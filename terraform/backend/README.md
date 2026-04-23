# Backend Module

This module provisions the backend services for the GitHub Metrics Aggregator.

It includes several services running on Cloud Run or as Cloud Run jobs:
- **Relay**: Processes events from Pub/Sub.
- **Retry**: Retries failed events.
- **Artifacts**: Manages job artifacts.
- **Commit Review Status**: Updates commit review status.

## Resources

- Cloud Run services and jobs.
- IAM bindings.
- Secrets (if any).
