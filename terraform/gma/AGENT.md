# GMA (Google Metrics Aggregator) Module `terraform/gma`

This module manages the underlying compute and triggering infrastructure for the GMA service handlers.

## Purpose
To provision the environments running metrics processors, webhook listeners, and webhook push/pull dispatch workloads (represented as Cloud Run deployment resources).

## Key Components defined

| Component | Files | Description |
| :--- | :--- | :--- |
| **Cloud Run Jobs** | `job_artifacts.tf`<br>`job_commit_review_status.tf`<br>`job_retry.tf` | Jobs used for asynchronous bulk processing items (e.g. syncs, retries) generally triggered via Cloud Scheduler. |
| **Cloud Run Services** | `service_relay.tf`<br>`service_webhook.tf` | Long-lived endpoints handling push notifications (e.g. GitHub webhook ingress, event relaying). |
| **Pub/Sub config** | `pubsub.tf` | Associated topics dispatching trigger configs to matching subscribers/jobs. |

## Notes & Design Patterns
- **Standardized Job Lifecycles**: Jobs create distinct Service Accounts (SA) holding just enough BigQuery and Pub/Sub IAM access to strictly do their scope of work (e.g. `artifacts-job` having read on `events` stream and write on `artifacts_status` output).
- **GitHub App Key Access**: Requires standard environment secret mounting mapped against version endpoints configured at the root orchestrator level.
- **Service Account Tying**: Connects fully with service accounts used at module levels to secure resources effectively.
