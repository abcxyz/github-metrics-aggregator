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

package artifact

import (
	"context"
	"fmt"
	"runtime"

	"github.com/abcxyz/github-metrics-aggregator/pkg/bq"
	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/workerpool"
)

// ExecuteJob runs the ingestion pipeline job to read GitHub
// action workflow logs from GitHub and store them into GCS.
func ExecuteJob(ctx context.Context, cfg *Config) error {
	logger := logging.FromContext(ctx)

	bqClient, err := bq.NewBigQuery(ctx, cfg.ProjectID, cfg.DatasetID)
	if err != nil {
		return fmt.Errorf("failed to create bigquery client: %w", err)
	}
	defer bqClient.Close()

	// Create a pool of workers to manage all of the log ingestions
	pool := workerpool.New[ArtifactRecord](&workerpool.Config{
		Concurrency: int64(runtime.NumCPU()),
		StopOnError: false,
	})

	// Setup a log ingester to process ingestion events
	logsFn, err := NewLogIngester(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create log ingester: %w", err)
	}

	logger.InfoContext(ctx, "ingestion job starting",
		"name", version.Name,
		"commit", version.Commit,
		"version", version.Version)

	// Read up to `BatchSize` number of events that need to be processed
	query, err := makeQuery(bqClient, cfg.EventsTableID, cfg.ArtifactsTableID, cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to populate query template: %w", err)
	}
	events, err := bq.Query[EventRecord](ctx, bqClient, query)
	if err != nil {
		return fmt.Errorf("failed to query bigquery for events: %w", err)
	}

	// Fan out the work of processing all of the events that were found
	for _, event := range events {
		if err := pool.Do(ctx, func() (ArtifactRecord, error) {
			artifact := logsFn.ProcessElement(ctx, *event)
			// Errors are handled by the element processor and are flagged as special
			// artifact records. There is no possible error returned from processing.
			return artifact, nil
		}); err != nil {
			return fmt.Errorf("failed to submit job to worker pool: %w", err)
		}
	}
	// When all of the workers are complete, extract the result values
	results, err := pool.Done(ctx)
	if err != nil {
		return fmt.Errorf("failed to ingest logs for events: %w", err)
	}
	artifacts := make([]*ArtifactRecord, 0, len(results))
	for _, v := range results {
		artifacts = append(artifacts, &v.Value)
	}

	// Save all of the result records to the output table
	if err := bq.Write[ArtifactRecord](ctx, bqClient, cfg.ArtifactsTableID, artifacts); err != nil {
		return fmt.Errorf("failed to write artifacts to bigquery: %w", err)
	}

	return nil
}
