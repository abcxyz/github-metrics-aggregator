package artifact

import (
	"context"
	"fmt"
	"runtime"

	"github.com/abcxyz/github-metrics-aggregator/pkg/version"
	"github.com/abcxyz/pkg/logging"
	"github.com/abcxyz/pkg/workerpool"
)

// ExecuteJob runs the ingestion pipeline job to read GitHub
// action workflow logs from GitHub and store them into GCS.
func ExecuteJob(ctx context.Context, cfg *Config) error {
	logger := logging.FromContext(ctx)

	bq, err := NewBigQuery(ctx, cfg.ProjectID, cfg.DatasetID)
	if err != nil {
		return fmt.Errorf("failed to create bigquery client: %w", err)
	}
	defer bq.Close()

	eventsTableDotNotation := formatGoogleSQL(cfg.ProjectID, cfg.DatasetID, cfg.EventsTableID)
	artifactsTableDotNotation := formatGoogleSQL(cfg.ProjectID, cfg.DatasetID, cfg.ArtifactsTableID)

	// Create a pool of workers to manage all of the log ingestions
	pool := workerpool.New[ArtifactRecord](&workerpool.Config{
		Concurrency: int64(runtime.NumCPU()),
		StopOnError: false,
	})

	// Setup a log ingester to process ingestion events
	logsFn, err := NewLogIngester(ctx, cfg.BucketName, cfg.GitHubAppID, cfg.GitHubInstallID, cfg.GitHubPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create log ingester: %w", err)
	}

	logger.InfoContext(ctx, "ingestion job starting",
		"name", version.Name,
		"commit", version.Commit,
		"version", version.Version)

	// Read up to `BatchSize` number of events that need to be processed
	query := fmt.Sprintf(SourceQuery, eventsTableDotNotation, artifactsTableDotNotation, cfg.BatchSize)
	events, err := Query[EventRecord](ctx, bq, query)
	if err != nil {
		return fmt.Errorf("failed to query bigquery for events: %w", err)
	}

	// Fan out the work of processing all of the events that were found
	for _, event := range events {
		pool.Do(ctx, func() (ArtifactRecord, error) {
			artifact := logsFn.ProcessElement(ctx, *event)
			// Errors are handled by the element processor and are flagged as special
			// artifact records. There is no possible error returned from processing.
			return artifact, nil
		})
	}
	// When all of the workers are complete, extract the result values
	results, err := pool.Done(ctx)
	if err != nil {
		return fmt.Errorf("failed to ingest logs for events: %w", err)
	}
	artifacts := Map(func(r *workerpool.Result[ArtifactRecord]) ArtifactRecord { return r.Value }, results)

	// Save all of the result records to the output table
	if err := Write(ctx, bq, cfg.ArtifactsTableID, artifacts); err != nil {
		return fmt.Errorf("failed to write artifacts to bigquery: %w", err)
	}

	return nil
}

// Map applies f to each element of xs, returning a new slice containing the results.
// Why is this not offered in the slices package in the standard library?
func Map[S any, E any](f func(S) E, xs []S) []E {
	ys := make([]E, len(xs))
	for i, x := range xs {
		ys[i] = f(x)
	}
	return ys
}
