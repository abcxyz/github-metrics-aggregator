# Version Package

This package provides version information for the GitHub Metrics Aggregator binary.

## Purpose
To store and expose version, commit SHA, and architecture information, typically set during the build process.

## Files

- **`version.go`**: Defines variables like `Name`, `Version`, `Commit`, `OSArch`, and `HumanVersion`. It attempts to read the VCS revision (commit SHA) from Go's build information if not explicitly set.

## Usage
This package is used by the CLI to display version information to the user.
