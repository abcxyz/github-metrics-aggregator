# Packages

This directory contains the core Go packages for the GitHub Metrics Aggregator application.

## Subdirectories

- **[`artifact`](./artifact)**: Job for ingesting GitHub Action workflow logs to GCS.
- **[`bq`](./bq)**: Generic BigQuery utilities.
- **[`cli`](./cli)**: CLI structure and commands.
- **[`events`](./events)**: Shared event data structures.
- **[`githubclient`](./githubclient)**: Wrapper around the GitHub API client.
- **[`pubsub`](./pubsub)**: Shared Pub/Sub utilities.
- **[`relay`](./relay)**: Service for enriching and relaying events.
- **[`retry`](./retry)**: Job for retrying failed event deliveries.
- **[`review`](./review)**: Job for auditing commit reviews.
- **[`teeth`](./teeth)**: Job for publishing log links in PR comments.
- **[`version`](./version)**: Version information.
- **[`webhook`](./webhook)**: Core logic for receiving and persisting webhooks.
