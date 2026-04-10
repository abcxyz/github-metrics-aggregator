# Webhook Package

This package implements the core logic for receiving GitHub webhook events and securely persisting them.

## Purpose
To provide an HTTP endpoint that GitHub can call when events occur, validate those requests, and queue them for processing via Pub/Sub.

## Files

- **`bigquery.go`**: Contains functions for interacting with BigQuery to track delivery failures and determine when to send events to the Dead Letter Queue (DLQ).
- **`bigquery_mock.go`**: Mock implementation of the datastore interface for testing purposes.
- **`config.go`**: Defines the configuration required for the webhook server (e.g., port, secret names, retry limits).
- **`config_test.go`**: Unit tests for the configuration logic.
- **`server.go`**: Handles the HTTP server setup, routing, and graceful shutdown.
- **`webhook.go`**: Contains the main HTTP handler (`handleWebhook`) that validates the GitHub signature, reads the payload, and publishes the event to Pub/Sub.
- **`webhook_test.go`**: Unit tests for the webhook handler logic.

## Design Patterns
- **Signature Validation**: Uses HMAC-SHA256 to validate that requests truly come from GitHub.
- **Resilience with DLQ**: If publishing to the main Pub/Sub topic fails repeatedly (tracked in BigQuery), the event is sent to a Dead Letter Queue to prevent data loss.
