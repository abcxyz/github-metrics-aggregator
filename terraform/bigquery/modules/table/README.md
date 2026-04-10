# BigQuery Table Module

This directory contains a reusable Terraform module for creating BigQuery tables with IAM permissions.

## Purpose
To standardize the creation of BigQuery tables, ensuring consistent application of schemas, partitioning, clustering, and access control.

## Files

- **`main.tf`**: Defines the `google_bigquery_table` resource and associated `google_bigquery_table_iam_member` resources for owners, editors, and viewers.
- **`outputs.tf`**: Exports attributes of the created table, such as `table_id` and `id`.
- **`variables.tf`**: Defines input variables for project ID, dataset ID, table ID, schema, partitioning options, and IAM bindings.

## Design Patterns
- **Dynamic Partitioning**: Uses a `dynamic` block for `time_partitioning` to allow optional configuration.
- **Granular IAM**: Applies IAM roles at the table level rather than the dataset level where appropriate, following the principle of least privilege.
