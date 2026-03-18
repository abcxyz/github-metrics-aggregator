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

## Notes & Design Patterns
- **Individual Table Files**: Each BigQuery table has its own dedicated `.tf` file (e.g., `table_optimized_events.tf`) to simplify maintenance of JSON schemas and table specific IAM bindings.
- **Aggregated View Queries**: Does not generally include views here; summary/computed setups tend to operate through the `scheduled_queries` companion module.
