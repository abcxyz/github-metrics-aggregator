# Integration Tests Package

This directory contains integration tests for the GitHub Metrics Aggregator system.

## Purpose
To perform end-to-end testing by simulating real webhook events from GitHub and verifying that they are correctly received, processed, and stored in BigQuery.

## Files

- **`bigquery.go`**: Helper functions for executing queries and polling for results in BigQuery during tests.
- **`config.go`**: Defines the configuration required for integration tests, loading environment variables for endpoint URLs and Google Cloud project details.
- **`main_test.go`**: The core integration test. It reads a sample payload, signs it, sends it to the deployed webhook service endpoint, and then queries BigQuery to confirm the record was written.

## Usage
These tests require a deployed environment and proper credentials to run. They are typically skipped during normal unit testing (via `testutil.SkipIfNotIntegration`).
