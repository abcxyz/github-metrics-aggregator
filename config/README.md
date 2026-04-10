# Config Package

This directory contains JSON metadata files that define parameters for various jobs in the system.

## Purpose
To provide structured definitions of input parameters required by different data processing jobs, likely used for UI generation or validation.

## Files

- **`artifacts_metadata.json`**: Defines parameters for the artifact ingestion job, including batch size, table names, bucket name, and GitHub App credentials.
- **`commit-review-status-metadata.json`**: Defines parameters for the commit review status job.

## Usage
These files are likely used by a template rendering system or a orchestrator to know what parameters to pass to the Cloud Run jobs.
