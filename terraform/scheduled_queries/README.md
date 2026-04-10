# Scheduled Queries Module `terraform/scheduled_queries`

This module manages BigQuery Data Transfer configurations for the Google Metrics Aggregator (GMA) system.

## Purpose
To execute scheduled SQL queries that process raw or optimized event logs and populate aggregate/summary tables used for dashboarding and reporting.

## Key Components

| Resource | Description |
| :--- | :--- |
| **`google_bigquery_data_transfer_config`** | Configures the interval, SQL query body, and destination table for a periodic sync job (e.g. `prstats_schedule`). |
| **IAM Permission (`google_project_iam_member`)** | Grants specific ShortTermTokenMinter permissions to the BigQuery Data Transfer Service Agent to allows it execution privileges on behalf of run SA accounts. |

## Queries Defined

- **`prstats.tf`**: Computes pull request metrics (insertions, deletions, reviewer interactions) by joining closed actions with comments streams.
- **`prstats_pull_requests.tf`**: Filters specific PR open/close triggers for aggregators.
- **`prstats_pull_request_reviews.tf`**: Tracks review activity timestamps.
- **`integration_events.tf`**: Aggregate event types for general integrations.

## Other Files

- **`main.tf`**: Grants IAM permissions (`roles/iam.serviceAccountShortTermTokenMinter`) to the BigQuery Data Transfer Service Agent.
- **`outputs.tf`**: Defines outputs for this module.
- **`terraform.tf`**: Terraform configuration, specifying required providers.
- **`variables.tf`**: Defines variables used in this module.

## Notes & Design Patterns
- **Standardized Retrying & Triggers**: Relies on BigQuery's built-in scheduler runtime dispatching background nodes.
- **Incremental inserts**: Queries tend to compute `WHERE received > (SELECT COALESCE(MAX(received), ...))` to avoid full table rewrites and optimize processing fee scanning on optimized partitioned logs.
- **Dedicated SAs**: SAs run discrete query sets isolating token scoping.
- **Location Alignment**: Schedule config matches dataset location bounds (e.g., "US").
