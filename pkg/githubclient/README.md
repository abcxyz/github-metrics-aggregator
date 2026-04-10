# GitHub Client Package

This package provides a wrapper around the GitHub API client, specifically tailored for use as a GitHub App.

## Purpose
To simplify authentication and interaction with the GitHub API, providing specific methods needed by other packages in this project (like `retry`).

## Files

- **`config.go`**: Defines the configuration for the GitHub client, supporting authentication via raw private key, Secret Manager, or Google Cloud KMS.
- **`config_test.go`**: Unit tests for the configuration logic.
- **`githubclient.go`**: The core client implementation. It handles authentication as a GitHub App and provides methods to list webhook deliveries and request redeliveries.
- **`lifecycle_test.go`**: Contains tests for the client lifecycle and integration with GitHub.

## Design Patterns
- **Wrapper Pattern**: Wraps the `google/go-github` client to provide a simpler interface and handle authentication details.
- **Flexible Authentication**: Supports multiple ways to provide the GitHub App's private key, making it suitable for different deployment environments.
