# Teeth Package

This package contains a job that reads leech and pull request event records from BigQuery and publishes available log invocations in a PR comment.

## Purpose
To provide automated feedback on pull requests by posting links to logs or status information derived from event processing.

## Subdirectories

- **[`sql`](./sql)**: Contains the SQL templates used by this package.

## Files

- **`bigquery.go`**: Implements the BigQuery client for this package, embedding the SQL query from `sql/publisher_source.sql` and handling execution and data insertion.
- **`bigquery_test.go`**: Unit tests for the BigQuery interaction logic.
- **`publish_logs.go`**: Defines the core interfaces and high-level logic for retrieving records to process and saving the results.

## Design Patterns
- **Embedded SQL**: Uses Go's `//go:embed` directive to include SQL files at compile time, keeping complex queries out of Go source files while avoiding runtime file loading issues.
- **Template Queries**: Uses `text/template` to dynamically inject table names into the embedded SQL query based on configuration.
