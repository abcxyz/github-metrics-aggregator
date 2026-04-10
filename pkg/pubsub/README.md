# Pub/Sub Utilities Package

This package provides utilities for interacting with Google Cloud Pub/Sub.

## Purpose
To provide a simple, interface-based way to send messages to Pub/Sub topics, facilitating testing and decoupling from the concrete Pub/Sub client.

## Files

- **`pubsub.go`**: Defines the `Messenger` interface and the `PubSubMessenger` struct that implements it. It handles client initialization and message publishing with a configurable timeout.
- **`pubsub_test.go`**: Unit tests for the Pub/Sub messenger.

## Design Patterns
- **Interface-based Design**: Defines the `Messenger` interface, allowing other packages (like `webhook`) to depend on an abstraction rather than a concrete implementation, making testing easier with mocks.
- **Wrapper Pattern**: Wraps the official `cloud.google.com/go/pubsub` client to provide a focused `Send` method.
