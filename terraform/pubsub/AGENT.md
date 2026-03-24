# Pub/Sub Module `terraform/pubsub`

This module manages the core messaging backbone focusing on event streams within the Google Metrics Aggregator (GMA) infrastructure.

## Purpose
To provision Pub/Sub topics and subscription sinks that decouple upstream webhooks or processing jobs from downstream data pipelines (like BigQuery insert streaming or DLQs).

## Key Components

| Resource | Description |
| :--- | :--- |
| **`google_pubsub_topic`** | Topics like `relay` mapping upstream streams into subscriptions. |
| **`google_pubsub_subscription`** | Subscriptions driving message pushes into target BigQuery tables or runner endpoints (e.g. `relay_optimized_events`). |
| **Dead Letter Queues** | Setups for failure stream redirection handling. |

## Notes & Design Patterns
- **Standardized Retries**: Subscriptions usually include standard dead-letter policies mapping failing inserts to dedicated retry endpoints for backpressure handling or triage.
- **Relay Dispatch**: Defines specific push/pull config items designed to coordinate workflow steps (e.g., matching SA execution tokens and endpoint subscription URL hooks).
- **Cross-Component Hooks**: Subscribed trigger types tie neatly to app endpoint controllers (such as the webhook Cloud Run listener).
