module "artifacts_table" {
  source = "./modules/table"

  project_id = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id

  table_id            = var.artifacts_table_id
  deletion_protection = false
  schema = jsonencode([
    {
      "name" : "delivery_id",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GUID that represents the event that was ingested."
    },
    {
      "name" : "processed_at",
      "type" : "TIMESTAMP",
      "mode" : "REQUIRED",
      "description" : "Timestamp of when the event was processed."
    },
    {
      "name" : "status",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The status of the log ingestion."
    },
    {
      "name" : "workflow_uri",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The original workflow uri that trigger the ingestion."
    },
    {
      "name" : "logs_uri",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The GCS uri of the logs."
    },
    {
      "name" : "github_actor",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GitHub user that triggered the workflow event."
    },
    {
      "name" : "organization_name",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GitHub organization name."
    },
    {
      "name" : "repository_name",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GitHub repository name."
    },
    {
      "name" : "repository_slug",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "Combined org/repo_name of the repository."
    },
    {
      "name" : "job_name",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "Apache Beam job name of the pipeline that processed this event."
    },
  ])
  iam = var.artifacts_table_iam
}
