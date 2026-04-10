# CLI Package

This package implements the command-line interface for the GitHub Metrics Aggregator.

## Purpose
To provide a structured CLI with subcommands for running the different components of the system (webhook server, retry job, relay service, etc.). It follows the `abcxyz` Go CLI pattern.

## Files

- **`artifact.go`**: Defines the command for artifact processing.
- **`relay.go`**: Defines the command for running the event relay service.
- **`retry.go`**: Defines the command for running the retry job.
- **`retry_test.go`**: Unit tests for the retry command.
- **`review.go`**: Defines the command for commit review status processing.
- **`root.go`**: Defines the root command and registers all subcommands.
- **`root_test.go`**: Unit tests for the root command.
- **`webhook.go`**: Defines the command for starting the webhook receiver server.
- **`webhook_test.go`**: Unit tests for the webhook command.

## Design Patterns
- **Command Pattern**: Each subcommand is defined in its own file and registered with the root command.
- **CLI Framework**: Uses `github.com/abcxyz/pkg/cli` for handling flags, arguments, and command execution.
