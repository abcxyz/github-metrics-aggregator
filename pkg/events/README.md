# Events Package

This package defines the data structures for GitHub events used throughout the application.

## Purpose
To provide shared struct definitions for events, ensuring consistency across different services (webhook, relay, BigQuery).

## Files

- **`event.go`**: Defines the `Event` struct, which represents the raw event received from GitHub, including delivery ID, signature, timestamp, event type, and the raw JSON payload.
- **`enriched_event.go`**: Defines the `EnrichedEvent` struct, which extends `Event` by adding fields for enterprise, organization, and repository IDs and names, extracted by the relay service.

## Usage
These structs are used for JSON marshaling/unmarshaling when passing messages via Pub/Sub and writing to BigQuery.
