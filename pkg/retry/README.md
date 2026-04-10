# Retry Package

This package implements a batch job that finds failed GitHub webhook deliveries and requests their redelivery.

## Purpose
To ensure data completeness by recovering events that failed to be delivered or processed initially. It polls GitHub's delivery log and triggers redeliveries for failed attempts.

## Files

- **`bigquery.go`**: Handles interactions with BigQuery for retrieving and writing checkpoints, and checking if a delivery event already exists.
- **`bigquery_mock.go`**: Mock implementation of the Datastore interface for testing.
- **`config.go`**: Defines the configuration for the retry job, including BigQuery table IDs, GCS bucket name, and GitHub app configuration.
- **`config_test.go`**: Unit tests for the configuration logic.
- **`github_mock.go`**: Mock implementation of the GitHubSource interface for testing.
- **`job.go`**: Contains the core logic of the retry job (`ExecuteJob`). It uses a GCS lock to prevent concurrent runs, tracks progress via checkpoints in BigQuery, and calls GitHub APIs to list and redeliver events.
- **`job_test.go`**: Unit tests for the execution job logic.
- **`lock_mock.go`**: Mock implementation of the GCS lock for testing.

## Design Patterns
- **Checkpointing**: Uses a BigQuery table to store the ID of the last successfully processed delivery, avoiding re-processing old events.
- **Distributed Locking**: Uses `go-gcslock` backed by Google Cloud Storage to ensure that only one instance of the job runs at a time.
- **Resilience**: Switches to a fresh context with a timeout when writing the final checkpoint to ensure progress is saved even if the main context was cancelled due to timeout.
