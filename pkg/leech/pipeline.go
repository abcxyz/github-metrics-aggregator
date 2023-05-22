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

// Package leech contains an Apache Beam data pipeline that will read workflow
// event records from BigQuery and ingest any available logs into cloud
// storage. A mapping from the original GitHub event to the cloud storage
// location is persisted in BigQuery along with an indicator for the status
// of the copy. The pipeline acts as a GitHub App for authentication purposes.
package leech

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/abcxyz/pkg/githubapp"
	"github.com/abcxyz/pkg/logging"
	"github.com/apache/beam/sdks/v2/go/pkg/beam"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/io/bigqueryio"
	"github.com/apache/beam/sdks/v2/go/pkg/beam/register"
	"github.com/lestrrat-go/jwx/v2/jwk"
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

// LeechRecord is the output data structure that maps to the leech pipeline's
// output table schema.
type LeechRecord struct {
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

// sourceQuery is the driving BigQuery query that selects events
// that need to be processed.
const sourceQuery = `
SELECT 
	delivery_id,
	JSON_VALUE(payload, "$.repository.full_name") repo_slug,
	JSON_VALUE(payload, "$.repository.name") repo_name,
	JSON_VALUE(payload, "$.repository.owner.login") org_name,
	JSON_VALUE(payload, "$.workflow_run.logs_url") logs_url,
	JSON_VALUE(payload, "$.workflow_run.actor.login") github_actor,
	JSON_VALUE(payload, "$.workflow_run.html_url") workflow_url
FROM ` + "`%s`" + `
WHERE
event = "workflow_run"
AND JSON_VALUE(payload, "$.workflow_run.status") = "completed"
AND delivery_id NOT IN (
SELECT
  delivery_id
FROM ` + "`%s`" + `
LIMIT %d)
LIMIT %d
`

// errLogsExpired is a marker error so that upstream processing knows
// that the logs for a given event no longer exist.
var errLogsExpired = errors.New("GitHub logs expired")

// Pipeline represents an implementation of Beam Pipeline that queries
// for events, looks for their workflow logs in GitHub and then copies them
// to Google Cloud Storage.
type Pipeline struct {
	cfg     *Config
	ghApp   *githubapp.GitHubApp
	storage ObjectWriter
}

// ingestLogsFn is an object that implements beams "DoFn" interface to
// provide the main processing of the event.
type ingestLogsFn struct {
	pipe   *Pipeline
	client *http.Client
}

// BeamInit handles preregistration of pipeline shapes with their associated
// type signature and shape to speed up runtime execution.
func BeamInit() {
	register.DoFn1x1[EventRecord, LeechRecord](&ingestLogsFn{})
}

// New creates a new instance of the leech Pipeline.
func New(ctx context.Context, cfg *Config, storage ObjectWriter) (*Pipeline, error) {
	pk, err := readPrivateKey(cfg.GitHubPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	ghAppConfig := githubapp.NewConfig(cfg.GitHubAppID, cfg.GitHubInstallID, pk)
	ghApp := githubapp.New(ghAppConfig)

	return &Pipeline{
		cfg:     cfg,
		ghApp:   ghApp,
		storage: storage,
	}, nil
}

// Prepare connects the internal pipeline implementation to the beam
// scope object associated with the primary beam Pipeline.
func (pipe *Pipeline) Prepare(scope beam.Scope) {
	// BigQuery table notation is not consistent so it needs represented in a few different
	// formats to appease the BigQuery client.
	eventsTableDotNotation := fmt.Sprintf("%s.%s", pipe.cfg.EventsProjectID, pipe.cfg.EventsTable)
	leechTableDotNotation := fmt.Sprintf("%s.%s", pipe.cfg.LeechProjectID, pipe.cfg.LeechTable)
	leechTableColonNotation := fmt.Sprintf("%s:%s", pipe.cfg.LeechProjectID, pipe.cfg.LeechTable)
	batchSize := pipe.cfg.BatchSize
	client := http.Client{Timeout: 5 * time.Minute}

	var event EventRecord
	query := fmt.Sprintf(sourceQuery, eventsTableDotNotation, leechTableDotNotation, batchSize, batchSize)
	// step 1: query BigQuery for unprocessed events
	col := bigqueryio.Query(scope, pipe.cfg.LeechProjectID, query, reflect.TypeOf(event), bigqueryio.UseStandardSQL())
	// step 2: process the events in parallel, ingesting logs
	res := beam.ParDo(scope, &ingestLogsFn{pipe: pipe, client: &client}, col)
	// step 3: write all of the results back to BigQuery
	bigqueryio.Write(scope, pipe.cfg.LeechProjectID, leechTableColonNotation, res)
}

// readPrivateKey reads a PEM encoded private key from a string.
func readPrivateKey(privateKeyContent string) (*rsa.PrivateKey, error) {
	parsedKey, _, err := jwk.DecodePEM([]byte(privateKeyContent))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PEM formated key:  %w", err)
	}
	privateKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("failed to convert to *rsa.PrivateKey (got %T)", parsedKey)
	}
	return privateKey, nil
}

// ProcessElement is a DoFn implementation that reads workflow logs from GitHub
// and stores them in Cloud Storage.
func (f *ingestLogsFn) ProcessElement(event EventRecord) LeechRecord {
	ctx := context.Background()
	logger := logging.FromContext(ctx)

	gcsPath := fmt.Sprintf("gs://%s/%s/%s/artifacts.tar.gz", f.pipe.cfg.LogsBucketName, event.RepositorySlug, event.DeliveryID)
	result := LeechRecord{
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
	logger.Infof("processing element", "DeliveryID", event.DeliveryID, "event", event, "result", result)
	if err := f.handleMessage(ctx, event.RepositoryName, event.LogsURL, gcsPath); err != nil {
		// Expired logs can never be retrieved, mark them as gone and move on
		if errors.Is(err, errLogsExpired) {
			logger.Infof("logs for workflow not available", "DeliveryID", event.DeliveryID)
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
			logger.Errorf("failed to retrieve logs for workflow: %w", err, "DeliveryID", event.DeliveryID)
			result.Status = "FAILURE"
		}
	}
	return result
}

// handleMessage is the main event processor. It generates a GitHub token, reads the workflow
// log files if they exist and persists them to Cloud Storage.
func (f *ingestLogsFn) handleMessage(ctx context.Context, repoName, ghLogsURL, gcsPath string) error {
	token, err := f.repoAccessToken(ctx, repoName)
	if err != nil {
		return fmt.Errorf("error getting GitHub access token: %w", err)
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

	if err := f.pipe.storage.Write(ctx, res.Body, gcsPath); err != nil {
		return fmt.Errorf("error copying logs to cloud storage: %w", err)
	}
	return nil
}

func (f *ingestLogsFn) repoAccessToken(ctx context.Context, repoName string) (string, error) {
	tokenRequest := githubapp.TokenRequest{
		Repositories: []string{repoName},
		Permissions: map[string]string{
			"actions": "read",
		},
	}
	// @TODO(bradegler): This could use some caching. Requests to the same repo
	// can reuse a token without requesting a new one until it expires. Might be
	// better to implement that in pkg so that GHTM can take advantage of it as well.
	ghAppJWT, err := f.pipe.ghApp.AccessToken(ctx, &tokenRequest)
	if err != nil {
		return "", fmt.Errorf("error creating GitHub access token: %w", err)
	}
	// The token response is a json doc with a lot of information about the
	// token. All that is needed is the token itself.
	var tokenResp map[string]any
	if err := json.Unmarshal([]byte(ghAppJWT), &tokenResp); err != nil {
		return "", fmt.Errorf("malformed GitHub token response: %w", err)
	}
	t, ok := tokenResp["token"]
	if !ok {
		return "", fmt.Errorf("malformed GitHub token response - missing token attribute")
	}
	token, ok := t.(string)
	if !ok {
		return "", fmt.Errorf("malformed GitHub token: wanted string got %T", token)
	}
	return token, nil
}
