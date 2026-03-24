module "commit_review_status_table" {
  source = "./modules/table"

  project_id = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id

  table_id            = var.commit_review_status_table_id
  deletion_protection = false
  schema = jsonencode([
    {
      "name" : "author",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The author of the commit."
    },
    {
      "name" : "organization",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The GitHub organization to which the commit belongs."
    },
    {
      "name" : "repository",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The GitHub repository to which the commit belongs."
    },
    {
      "name" : "branch",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The GitHub branch to which the commit belongs."
    },
    {
      "name" : "visibility",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The repository visibility"
    },
    {
      "name" : "commit_sha",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The SHA Hash for the commit."
    },
    {
      "name" : "commit_timestamp",
      "type" : "TIMESTAMP",
      "mode" : "REQUIRED",
      "description" : "The Timestamp when the commit was made"
    },
    {
      "name" : "commit_html_url",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The URL for the commit in GitHub"
    },
    {
      "name" : "pull_request_id",
      "type" : "INT64",
      "mode" : "NULLABLE",
      "description" : "The id of the pull request that introduced the commit."
    },
    {
      "name" : "pull_request_number",
      "type" : "INT64",
      "mode" : "NULLABLE",
      "description" : "The number of the pull request that introduced the commit."
    },
    {
      "name" : "pull_request_html_url",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The html url of the pull request that introduced the commit."
    },
    {
      "name" : "approval_status",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The approval status of the commit in GitHub."
    },
    {
      "name" : "break_glass_issue_urls",
      "type" : "STRING",
      "mode" : "REPEATED",
      "description" : "The URLs of the break glass issues that the author had open during the time the commit was made."
    },
    {
      "name" : "note",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Optional context on the about the commit (e.g. a processing error message)"
    },
  ])
  iam = var.commit_review_status_table_iam
}
