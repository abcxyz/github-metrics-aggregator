# SQL Templates Package

This directory contains SQL template files used by the `teeth` package.

## Purpose
To store complex SQL queries that are used by the Go code, allowing them to be maintained as separate files rather than embedded strings.

## Files

- **`publisher_source.sql`**: A templated SQL query that retrieves pull request events that need processing (specifically, those that have not yet been commented on). It uses Go template syntax (e.g., `{{.PullRequestEventsTable}}`) for table names.

## Usage
These SQL files are read by the Go application, executed after filling in the template parameters, and the results are processed by the application logic.
