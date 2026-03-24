module "invocation_comment_table" {
  count  = var.invocation_comment.enabled ? 1 : 0
  source = "./modules/table"

  project_id = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id

  table_id            = var.invocation_comment.table_id
  deletion_protection = false
  schema = jsonencode([
    {
      "name" : "pull_request_id",
      "type" : "INT64",
      "mode" : "REQUIRED",
      "description" : "ID of pull request."
    },
    {
      "name" : "pull_request_html_url",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "URL of pull request."
    },
    {
      "name" : "processed_at",
      "type" : "TIMESTAMP",
      "mode" : "REQUIRED",
      "description" : "Timestamp of when the analyzer pipeline processed the PR."
    },
    {
      "name" : "comment_id",
      "type" : "INT64",
      "mode" : "NULLABLE",
      "description" : "ID of pull request comment."
    },
    {
      "name" : "status",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The status of invocation comment operation."
    },
    {
      "name" : "job_name",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "Job name of the analyzer that processed this event."
    },
  ])
  iam = try(var.invocation_comment.table_iam, {})
}
