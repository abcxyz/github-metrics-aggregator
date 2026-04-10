# Relay Package

This package implements a service that receives GitHub webhook events from a Pub/Sub push subscription, enriches them with metadata, and publishes them to another Pub/Sub topic.

## Purpose
To process raw webhook events and add useful attributes (like organization, repository, and enterprise IDs and names) before forwarding them. This allows for better filtering and routing of events downstream.

## Files

- **`config.go`**: Defines the configuration for the relay service (port, topic IDs, etc.).
- **`config_test.go`**: Unit tests for the configuration logic.
- **`enricher.go`**: Contains the core logic for parsing the raw GitHub event payload and extracting organization, repository, and enterprise information to create an `EnrichedEvent`. It also generates Pub/Sub message attributes.
- **`relay.go`**: Contains the HTTP handler (`handleRelay`) that receives the Pub/Sub message, calls the enricher, and sends the enriched message to the target topic.
- **`relay_test.go`**: Unit tests for the relay handler and enrichment logic.
- **`server.go`**: Handles the HTTP server setup and routing for the relay service.

## Design Patterns
- **Event Enrichment**: Extracts key fields from a nested JSON payload and elevates them to top-level fields and message attributes, making downstream processing more efficient.
- **Push-to-Push Processing**: Acts as a bridge between two Pub/Sub topics, triggered by a push subscription and publishing to another topic.
