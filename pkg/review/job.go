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
	"github.com/abcxyz/github-metrics-aggregator/pkg/githubclient"
	"github.com/abcxyz/github-metrics-aggregator/pkg/kms"
	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/githubauth"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/workerpool"
)

// ExecuteJob runs the pipeline job to read GitHub commits to check if they were
// properly reviewed.
func ExecuteJob(ctx context.Context, cfg *Config) error {
	logger := logging.FromContext(ctx)

	bqClient, err := bq.NewBigQuery(ctx, cfg.ProjectID, cfg.DatasetID)
	if err != nil {
		return fmt.Errorf("failed to create bigquery client: %w", err)
	}
	defer bqClient.Close()

	var app *githubauth.App
	if cfg.GitHubPrivateKeyKMSKeyID != "" {
		kmc, err := kms.NewKeyManagement(ctx)
		if err != nil {
			return fmt.Errorf("failed to create kms client: %w", err)
		}
		defer kmc.Close()

		signer, err := kmc.CreateSigner(ctx, cfg.GitHubPrivateKeyKMSKeyID)
		if err != nil {
			return fmt.Errorf("failed to create app signer: %w", err)
		}
		app, err = githubclient.NewGitHubAppFromSigner(ctx, signer, cfg.GitHubAppID, cfg.GitHubEnterpriseServerURL)
		if err != nil {
			return fmt.Errorf("failed to create github app from kms: %w", err)
		}
	} else if cfg.GitHubPrivateKeySecret != "" {
		app, err = githubclient.NewGitHubApp(ctx, cfg.GitHubEnterpriseServerURL, cfg.GitHubAppID, cfg.GitHubPrivateKeySecret)
		if err != nil {
			return fmt.Errorf("failed to create github app: %w", err)
		}
	}

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
			installation, err := app.InstallationForOrg(ctx, commit.Organization)
			if err != nil {
				return nil, fmt.Errorf("failed to get github app installation for org %s: %w", commit.Organization, err)
			}

			githubTokenSource := installation.AllReposTokenSource(map[string]string{
				"contents":      "read",
				"pull_requests": "read",
			})

			gitHubToken, err := githubTokenSource.GitHubToken(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get github token: %w", err)
			}
			gitHubClient := NewGitHubEnterpriseGraphQLClient(ctx, cfg.GitHubEnterpriseServerURL, gitHubToken)
			return processCommit(ctx, gitHubClient, commit), nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to process commits: %w", err)
	}

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

	// Step 4: Write the commit review status information to BigQuery.
	if err := bq.Write[CommitReviewStatus](ctx, bqClient, cfg.CommitReviewStatusTableID, taggedReviewStatuses); err != nil {
		return fmt.Errorf("failed to write commit review statuses to bigquery: %w", err)
	}

	return nil
}

// pooledTransform transforms each input element of type E into an element of
// type V using the given transform function. The transform is fanned out using
// a worker pool so that each input element may be processed asynchronously from
// the others.
//
// Any nil elements or nil results are excluded from the returned values.
func pooledTransform[E, V any](ctx context.Context, elements []*E, transform func(*E) (*V, error)) ([]*V, error) {
	// Create a pool of workers to manage the transformation
	workerPool := workerpool.New[*V](&workerpool.Config{
		Concurrency: int64(runtime.NumCPU()),
		StopOnError: false,
	})

	// schedule each element transformation in the worker pool
	for _, e := range elements {
		if e == nil {
			continue
		}

		if err := workerPool.Do(ctx, func() (*V, error) {
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
	values := make([]*V, 0, len(results))
	for _, v := range results {
		if v.Value == nil {
			continue
		}

		values = append(values, v.Value)
	}

	return values, nil
}
