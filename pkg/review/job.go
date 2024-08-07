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
		return fmt.Errorf("failed to query bigquery for commits: %w", err)
	}

	// Step 2: Get review status information for each commit.
	commitReviewStatuses, err := pooledTransform(ctx, commits,
		func(commit *Commit) (*CommitReviewStatus, error) {
			return processCommit(ctx, gitHubClient, commit), nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to process commits: %w", err)
	}
	// statuses that should not be further processed will be nil, so we should exclude them
	commitReviewStatuses = removeNil(commitReviewStatuses)

	// Step 3: Look up break glass issue if necessary and tag the review status with it if found.
	fetcher := &BigQueryBreakGlassIssueFetcher{
		client: bqClient,
	}
	taggedReviewStatuses, err := pooledTransform(ctx, commitReviewStatuses,
		func(status *CommitReviewStatus) (*CommitReviewStatus, error) {
			return processReviewStatus(ctx, fetcher, cfg, status), nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to process commit review statuses: %w", err)
	}
	// statuses that should not be further processed will be nil
	taggedReviewStatuses = removeNil(taggedReviewStatuses)

	// Step 4: Write the commit review status information to BigQuery.
	if err := bq.Write[CommitReviewStatus](ctx, bqClient, cfg.CommitReviewStatusTableID, taggedReviewStatuses); err != nil {
		return fmt.Errorf("failed to write commit review statuses to bigquery: %w", err)
	}

	return nil
}

// pooledTransform transforms each input element of type T into an element of type V using the given transform function.
// The transform is fanned out using a worker pool so that each input element may be processed asynchronously from the
// others.
func pooledTransform[T, V any](ctx context.Context, elements []T, transform func(T) (V, error)) ([]V, error) {
	// Create a pool of workers to manage the transformation
	workerPool := workerpool.New[V](&workerpool.Config{
		Concurrency: int64(runtime.NumCPU()),
		StopOnError: false,
	})

	// schedule each element transformation in the worker pool
	for _, e := range elements {
		if err := workerPool.Do(ctx, func() (V, error) {
			return transform(e)
		}); err != nil {
			return nil, fmt.Errorf("failed to submit job to worker pool: %w", err)
		}
	}

	// When all the workers are complete, extract the result values
	results, err := workerPool.Done(ctx)
	if err != nil {
		return nil, fmt.Errorf("worker pool failed: %w", err)
	}
	values := make([]V, 0, len(results))
	for _, v := range results {
		values = append(values, v.Value)
	}

	return values, nil
}

func removeNil[T any](elements []*T) []*T {
	var filtered []*T
	for _, e := range elements {
		if e != nil {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
