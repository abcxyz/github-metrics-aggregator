# Review Package

This package implements a pipeline job that checks GitHub commits to verify if they were properly reviewed according to organizational policies.

## Purpose
To audit commits and ensure they meet review requirements, identifying unreviewed commits or those made using "breakglass" (emergency) procedures.

## Files

- **`breakglass_query.go`**: Functions to query BigQuery for breakglass issues.
- **`breakglass_query_test.go`**: Unit tests for breakglass query logic.
- **`commit_query.go`**: Functions to query BigQuery for commits that need to be processed.
- **`commit_query_test.go`**: Unit tests for commit query logic.
- **`commit_review_status.go`**: Contains the core logic for processing a commit, using the GitHub GraphQL API to fetch pull request and review information, and determining the review status.
- **`commit_review_status_test.go`**: Extensive tests for the commit review status logic, using mock data.
- **`config.go`**: Defines the configuration for the review job.
- **`issue_fetcher.go`**: Defines interfaces and BigQuery implementations for fetching breakglass issues.
- **`job.go`**: The main entry point (`ExecuteJob`) that orchestrates the pipeline: querying commits, processing them in parallel using a worker pool, checking for breakglass justification, and writing results back to BigQuery.

## Design Patterns
- **Worker Pool**: Uses `github.com/abcxyz/pkg/workerpool` to process commits concurrently, improving performance when dealing with many commits and API calls.
- **GraphQL**: Uses GitHub's GraphQL API (via `CommitReviewStatus`) to efficiently fetch complex relational data (commits, pull requests, reviews) in fewer requests than the REST API would require.
- **Breakglass Detection**: Correlates commits with "breakglass" issues (emergency declarations) to justify unreviewed commits.
