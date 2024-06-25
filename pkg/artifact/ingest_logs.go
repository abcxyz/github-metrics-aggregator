// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package artifact contains a data pipeline that will read workflow
// event records from BigQuery and ingest any available logs into cloud
// storage. A mapping from the original GitHub event to the cloud storage
// location is persisted in BigQuery along with an indicator for the status
// of the copy. The pipeline acts as a GitHub App for authentication purposes.
package artifact

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/abcxyz/pkg/githubauth"
	"github.com/abcxyz/pkg/logging"
)

// EventRecord maps the columns from the driving BigQuery query
// to a usable structure.
type EventRecord struct {
	DeliveryID       string `bigquery:"delivery_id" json:"delivery_id"`
	RepositorySlug   string `bigquery:"repo_slug" json:"repo_slug"`
	RepositoryName   string `bigquery:"repo_name" json:"repo_name"`
	OrganizationName string `bigquery:"org_name" json:"org_name"`
	LogsURL          string `bigquery:"logs_url" json:"logs_url"`
	GitHubActor      string `bigquery:"github_actor" json:"github_actor"`
	WorkflowURL      string `bigquery:"workflow_url" json:"workflow_url"`
}

// ArtifactRecord is the output data structure that maps to the leech pipeline's
// output table schema.
type ArtifactRecord struct {
	DeliveryID       string    `bigquery:"delivery_id" json:"delivery_id"`
	ProcessedAt      time.Time `bigquery:"processed_at" json:"processed_at"`
	Status           string    `bigquery:"status" json:"status"`
	WorkflowURI      string    `bigquery:"workflow_uri" json:"workflow_uri"`
	LogsURI          string    `bigquery:"logs_uri" json:"logs_uri"`
	GitHubActor      string    `bigquery:"github_actor" json:"github_actor"`
	OrganizationName string    `bigquery:"organization_name" json:"organization_name"`
	RepositoryName   string    `bigquery:"repository_name" json:"repository_name"`
	RepositorySlug   string    `bigquery:"repository_slug" json:"repository_slug"`
	JobName          string    `bigquery:"job_name" json:"job_name"`
}

// errLogsExpired is a marker error so that upstream processing knows
// that the logs for a given event no longer exist.
var errLogsExpired = errors.New("GitHub logs expired")

// logIngester is an object that provides the main processing of the event.
type logIngester struct {
	client             *http.Client
	githubInstallation *githubauth.AppInstallation
	storage            ObjectWriter
	bucketName         string
}

// NewLogIngester creates a logIngester and initializes the object store, GitHub app and http client.
func NewLogIngester(ctx context.Context, logsBucketName, gitHubAppID, gitHubInstallID, gitHubPrivateKey string) (*logIngester, error) {
	// create an object store
	store, err := NewObjectStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create object store client: %w", err)
	}

	app, err := githubauth.NewApp(gitHubAppID, gitHubPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create github app: %w", err)
	}

	installation, err := app.InstallationForID(ctx, gitHubInstallID)
	if err != nil {
		return nil, fmt.Errorf("failed to get github app installation: %w", err)
	}

	return &logIngester{
		storage: store,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
		githubInstallation: installation,
		bucketName:         logsBucketName,
	}, nil
}

// ProcessElement is the main processing function for the logIngester implementation that
// reads workflow logs from GitHub and stores them in Cloud Storage.
func (f *logIngester) ProcessElement(ctx context.Context, event EventRecord) ArtifactRecord {
	logger := logging.FromContext(ctx)

	logger.InfoContext(ctx, "process element", "delivery_id", event.DeliveryID)

	gcsPath := fmt.Sprintf("gs://%s/%s/%s/artifacts.tar.gz", f.bucketName, event.RepositorySlug, event.DeliveryID)
	result := ArtifactRecord{
		DeliveryID:       event.DeliveryID,
		ProcessedAt:      time.Now(),
		WorkflowURI:      event.WorkflowURL,
		GitHubActor:      event.GitHubActor,
		OrganizationName: event.OrganizationName,
		RepositoryName:   event.RepositoryName,
		RepositorySlug:   event.RepositorySlug,
		LogsURI:          gcsPath,
		Status:           "SUCCESS",
	}
	logger.InfoContext(ctx, "processing element",
		"delivery_id", event.DeliveryID,
		"event", event,
		"result", result)
	if err := f.handleMessage(ctx, event.RepositoryName, event.LogsURL, gcsPath); err != nil {
		// Expired logs can never be retrieved, mark them as gone and move on
		if errors.Is(err, errLogsExpired) {
			logger.InfoContext(ctx, "logs for workflow not available", "delivery_id", event.DeliveryID)
			result.Status = "NOT_FOUND"
		} else {
			// Other failures can be retried
			// @TODO(bradegler): These are currently filtered out of the query, need to figure
			// out a way to mark how many attempts have been made for a particular delivery id
			// preferably without causing an update to the row in BigQuery. The simplest approach
			// would be to just not write a FAILURE status and let the query try again. The drawback
			// is that something could be stuck forever in that situation and we wouldn't want to
			// keep processing it. I think a secondary FAILURE table might work that we can join
			// into the main query as WHERE count(failures where delivery_id = x) < 10 or something.
			// This adds complexity to the write operation though so it requires some thought.
			// For now just flag rows as FAILUREs and we can delete them from the table to trigger
			// reprocessing.
			logger.ErrorContext(ctx, "failed to retrieve logs for workflow",
				"error", err,
				"delivery_id", event.DeliveryID,
			)
			result.Status = "FAILURE"
		}
	}
	return result
}

// handleMessage is the main event processor. It generates a GitHub token, reads the workflow
// log files if they exist and persists them to Cloud Storage.
func (f *logIngester) handleMessage(ctx context.Context, repoName, ghLogsURL, gcsPath string) error {
	token, err := f.githubInstallation.AccessToken(ctx, &githubauth.TokenRequest{
		Repositories: []string{repoName},
		Permissions: map[string]string{
			"actions": "read",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get token for repository %q: %w", repoName, err)
	}

	// Create a request to the workflow logs endpoint. This will follow redirects
	// by default which is important since the endpoint returns a 302 w/ a short lived
	// url that expires.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ghLogsURL, nil)
	if err != nil {
		return fmt.Errorf("error creating http request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := f.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making http request %w", err)
	}
	defer res.Body.Close()

	// Check for not found conditions. This signals that the logs have expired
	// and there is nothing that can be done about it.
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusGone {
		return errLogsExpired
	}
	// If the request wasn't successful try to determine why and respond with
	// an error containing the response from GitHub if possible.
	if res.StatusCode != http.StatusOK {
		content, err := io.ReadAll(io.LimitReader(res.Body, 256_000))
		if err != nil {
			return fmt.Errorf("error response from GitHub - failed to read response body")
		}
		return fmt.Errorf("error response from GitHub - response body: %q", string(content))
	}

	if err := f.storage.Write(ctx, res.Body, gcsPath); err != nil {
		return fmt.Errorf("error copying logs to cloud storage: %w", err)
	}
	return nil
}
