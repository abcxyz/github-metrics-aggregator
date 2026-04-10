# Artifact Package

This package implements a pipeline job that ingests GitHub Action workflow logs and stores them in Google Cloud Storage (GCS).

## Purpose
To download and persist workflow logs for further analysis, as GitHub only retains them for a limited time.

## Files

- **`config.go`**: Defines the configuration for the artifact ingestion job.
- **`ingest_logs.go`**: Contains the `LogIngester` struct and methods for downloading logs from GitHub and uploading them to GCS.
- **`ingest_logs_test.go`**: Unit tests for the log ingestion logic.
- **`job.go`**: The main entry point (`ExecuteJob`) that orchestrates the pipeline: querying BigQuery for events to process, fanning out the work to a worker pool, and writing completion records back to BigQuery.
- **`query.go`**: Generates the SQL query used to find events that need log ingestion.
- **`storage.go`**: Provides utilities for writing data to Google Cloud Storage.

## Design Patterns
- **Worker Pool**: Uses `github.com/abcxyz/pkg/workerpool` to process multiple log ingestions concurrently, handled by `ExecuteJob`.
- **Data Pipeline**: Follows a typical ETL (Extract, Transform, Load) pattern: extract event IDs from BigQuery, extract logs from GitHub, load logs to GCS, and load status to BigQuery.
