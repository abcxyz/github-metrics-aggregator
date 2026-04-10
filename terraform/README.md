# Terraform Root Module `terraform/`

The root Terraform module acts as the **orchestrator** for the entire Google Metrics Aggregator (GMA) infrastructure. It defines global provider settings, centralizes configuration using local variables, and calls specialized submodules.

## Purpose
To provision and manage the full lifecycle of resources required to collect, process, and query GitHub metrics using Google Cloud Platform (GCP).

## Submodules Orchestrated

| Module | Location | Description |
| :--- | :--- | :--- |
| **`bigquery`** | `./bigquery` | Dataset, tables, and granular BigQuery IAM bindings. |
| **`pubsub`** | `./pubsub` | Pub/Sub topics for relay nodes and subscription sinks. |
| **`gma`** | `./gma` | Core application running Cloud Run jobs (e.g., `artifacts`). |
| **`scheduled_queries`** | `./scheduled_queries` | Data transfer configs that populate aggregate calculation tables. |

## Key Config Variables (`example_main.tf` Locals)
The example root module aggregates all options into a `locals` block in `example_main.tf`. Key groupings include:
- **Core / Shared Settings**: `project_id`, `region`, `dataset_location`
- **Application Specs**: `image` paths, target `endpoints`, `github_app_id` secrets.
- **Table IDs Reference**: Unique addresses mappings used to hook queries up with datasets correctly.

## Files

- **`example_main.tf`**: Calls the 4 core submodules, piping variables or linking outputs. Used as an example for consumers.
- **`example_moved.tf`**: Centralized historical log instructions preserving state during migrations.

## Quality & Linting

To maintain code standards and satisfy CI checks, you should run the Terraform linter locally before committing changes.

### Running the Linter

Execute the following command on the target directory (e.g., `.`, `./bigquery`, `./gma`):

```bash
# Run from within the terraform/ directory or pass direct folder path
go run github.com/abcxyz/terraform-linter/cmd/terraform-linter@main lint <directory_path>
```

### Resolving Common Lint Errors

- **`TF050` (Missing Newline Error)**: This frequently occurs when a meta-block (e.g., `lifecycle`, `depends_on`, `count`, `for_each`) or provider-specific attribute sits immediately adjacent to general attributes without an empty line spacer separating them.
  - **Fix**: Insert a single empty newline above the meta-block or provider stanza to restore compliance layout boundaries.
- **Formatting**: Always run standard `terraform fmt` to fix indentation and alignment before running opinionated structural linters.
