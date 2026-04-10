# BigQuery Module `terraform/bigquery`

This module manages the core storage and dataset environment for the Google Metrics Aggregator (GMA) system.

## Purpose
To create a BigQuery dataset and define the underlying schema, partitioning, clustering presets, and IAM access controls for various event-tracking and reporting tables.

## Key Resources Defined

| Resource / Module | Description |
| :--- | :--- |
| **`google_bigquery_dataset`** | Creates the `github_metrics` dataset representing the single dataset pool. |
| **Table Modules (`table_*.tf`)** | Uses a child `./modules/table` structure to encapsulate schema configurations for things like `events`, `optimized_events`, `checkpoint`, `artifacts_status`, etc. |
| **IAM bindings** | Distributes specific dataset and table level accessor roles to reader/writer endpoints like Cloud Run or Pub/Sub handlers. |

## Sub-Modules Structure
- **`modules/table/`**: Encapsulates iterative creation for tables setting defaults for:
  - `time_partitioning` (usually standardizing on DAY boundaries).
  - `clustering` presets to optimize query scanning costs.

## Files

- **`main.tf`**: Defines the `google_bigquery_dataset` and shared resources.
- **`outputs.tf`**: Exports dataset and table IDs.
- **`terraform.tf`**: Terraform configuration and provider constraints.
- **`variables.tf`**: Input variables for the module.
- **`table_artifacts.tf`**: Defines the table for artifact ingestion status.
- **`table_checkpoint.tf`**: Defines the table for retry job checkpoints.
- **`table_commit_review_status.tf`**: Defines the table for commit review status.
- **`table_events.tf`**: Defines the main events table.
- **`table_failure_events.tf`**: Defines the table for tracking failed events.
- **`table_integration_events.tf`**: Defines the table for integration test events.
- **`table_invocation_comment.tf`**: Defines the table for tracking PR comments made by the teeth job.
- **`table_optimized_events.tf`**: Defines an optimized view or table for events.
- **`table_prstats.tf`**: Defines a table for PR statistics.
- **`table_prstats_pull_request_reviews.tf`**: Defines a table for PR reviews stats.
- **`table_prstats_pull_requests.tf`**: Defines a table for PR stats.
- **`table_raw_events.tf`**: Defines the table for raw un-enriched events.

## Notes & Design Patterns
- **Individual Table Files**: Each BigQuery table has its own dedicated `.tf` file (e.g., `table_optimized_events.tf`) to simplify maintenance of JSON schemas and table specific IAM bindings.
- **Aggregated View Queries**: Does not generally include views here; summary/computed setups tend to operate through the `scheduled_queries` companion module.
