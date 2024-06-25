// Copyright 2024 The Authors (see AUTHORS file)
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

package review

import (
	"context"
	"fmt"
	"runtime"

	"github.com/abcxyz/github-metrics-aggregator/pkg/bq"
	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/githubauth"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/workerpool"
)

// ExecuteJob runs the pipeline job to read GitHub
// commits to check if they were properly reviewed.
func ExecuteJob(ctx context.Context, cfg *Config) error {
	logger := logging.FromContext(ctx)

	bqClient, err := bq.NewBigQuery(ctx, cfg.ProjectID, cfg.DatasetID)
	if err != nil {
		return fmt.Errorf("failed to create bigquery client: %w", err)
	}
	defer bqClient.Close()

	app, err := githubauth.NewApp(cfg.GitHubAppID, cfg.GitHubPrivateKeySecret)
	if err != nil {
		return fmt.Errorf("failed to create github app: %w", err)
	}

	installation, err := app.InstallationForID(ctx, cfg.GitHubInstallID)
	if err != nil {
		return fmt.Errorf("failed to get github app installation: %w", err)
	}

	githubTokenSource := installation.AllReposTokenSource(map[string]string{
		"actions":       "read",
		"contents":      "read",
		"pull_requests": "read",
	})

	gitHubToken, err := githubTokenSource.GitHubToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get github token: %w", err)
	}
	gitHubClient := NewGitHubGraphQLClient(ctx, gitHubToken)

	// Create a pool of workers to manage all of the log ingestions
	commitPool := workerpool.New[*CommitReviewStatus](&workerpool.Config{
		Concurrency: int64(runtime.NumCPU()),
		StopOnError: false,
	})

	logger.InfoContext(ctx, "review job starting",
		"name", version.Name,
		"commit", version.Commit,
		"version", version.Version)

	// Step 1: Get commits that need to be processed from BigQuery.
	query, err := makeCommitQuery(cfg)
	if err != nil {
		return fmt.Errorf("failed to created commit query: %w", err)
	}
	commits, err := bq.Query[Commit](ctx, bqClient, query)
	if err != nil {
		return fmt.Errorf("failed to query bigquery for events: %w", err)
	}

	// Step 2: Get review status information for each commit.
	// Fan out the work of processing all of the commits that were found
	for _, commit := range commits {
		if err := commitPool.Do(ctx, func() (*CommitReviewStatus, error) {
			result := processCommit(ctx, *commit, gitHubClient)
			return result, nil
		}); err != nil {
			return fmt.Errorf("failed to submit job to worker pool: %w", err)
		}
	}
	// When all of the workers are complete, extract the result values
	results, err := commitPool.Done(ctx)
	if err != nil {
		return fmt.Errorf("failed to ingest logs for events: %w", err)
	}
	statuses := make([]*CommitReviewStatus, 0, len(results))
	for _, v := range results {
		// Statuses that were failed to be retrieved will be nil
		if v.Value != nil {
			statuses = append(statuses, v.Value)
		}
	}

	// Step 3: Look up break glass issue if necessary.
	statusPool := workerpool.New[*CommitReviewStatus](&workerpool.Config{
		Concurrency: int64(runtime.NumCPU()),
		StopOnError: false,
	})
	fetcher := BigQueryBreakGlassIssueFetcher{
		client: bqClient,
	}
	// Fan out the work of processing all of the commits that were found
	for _, status := range statuses {
		if err := statusPool.Do(ctx, func() (*CommitReviewStatus, error) {
			result := processReviewStatus(ctx, &fetcher, cfg, *status)
			return result, nil
		}); err != nil {
			return fmt.Errorf("failed to submit job to worker pool: %w", err)
		}
	}
	// When all of the workers are complete, extract the result values
	reviewStatuses, err := commitPool.Done(ctx)
	if err != nil {
		return fmt.Errorf("failed to ingest logs for events: %w", err)
	}
	taggedReviewStatuses := make([]*CommitReviewStatus, 0, len(reviewStatuses))
	for _, v := range results {
		// Statuses that were failed to be retrieved will be nil
		if v.Value != nil {
			statuses = append(statuses, v.Value)
		}
	}
	// Step 4: Write the commit review status information to BigQuery.
	if err := bq.Write[CommitReviewStatus](ctx, bqClient, cfg.CommitReviewStatusTableID, taggedReviewStatuses); err != nil {
		return fmt.Errorf("failed to write artifacts to bigquery: %w", err)
	}

	return nil
}
