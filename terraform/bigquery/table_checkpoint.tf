module "checkpoint_table" {
  source = "./modules/table"

  project_id = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id

  table_id = var.checkpoint_table_id
  schema = jsonencode([
    {
      "name" : "delivery_id",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GUID that represents the last successfully redelivered event sent to GitHub."
    },
    {
      "name" : "created",
      "type" : "TIMESTAMP",
      "mode" : "REQUIRED",
      "description" : "Timestamp for when the checkpoint record was created."
    },
    {
      "name" : "github_instance_url",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The github instance the retry service is running for."
    },
  ])
  iam = var.checkpoint_table_iam
}
