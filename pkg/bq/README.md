# BigQuery Utilities Package

This package provides generic utilities for interacting with Google BigQuery.

## Purpose
To reduce boilerplate code when executing queries and writing data to BigQuery, providing a simple, generic interface for mapping rows to Go structs.

## Files

- **`bigquery.go`**: Defines the `BigQuery` wrapper struct and generic functions `Query[T]` and `Write[T]` that handle executing SQL and inserting rows, respectively.
- **`bigquery_test.go`**: Unit tests for the BigQuery utilities, likely using mocks or small-scale tests.

## Design Patterns
- **Generics**: Uses Go generics to allow mapping BigQuery rows to any user-defined struct, providing type safety and reducing manual mapping code.
- **Wrapper Pattern**: Wraps the official `cloud.google.com/go/bigquery` client to provide a more focused API for this project's needs.
