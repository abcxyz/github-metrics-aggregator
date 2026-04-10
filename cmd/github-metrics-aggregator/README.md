# GitHub Metrics Aggregator Command

This directory contains the main entry point for the GitHub Metrics Aggregator application.

## Purpose
To provide the executable binary that users run. It sets up signal handling and logging, and delegates execution to the `pkg/cli` package.

## Files

- **`main.go`**: The entry point file. It handles `SIGINT` and `SIGTERM` for graceful shutdown and calls `cli.Run` with command-line arguments.

## Usage
This is the main command invoked to run the webhook server, retry job, or other subcommands defined in the CLI.
