module "failure_events_table" {
  source = "./modules/table"

  project_id = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id

  table_id = var.failure_events_table_id
  schema = jsonencode([
    {
      "name" : "delivery_id",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GUID that represents the failed event."
    },
    {
      "name" : "created",
      "type" : "TIMESTAMP",
      "mode" : "REQUIRED",
      "description" : "Timestamp of when the event fails to process."
    },
  ])
  iam = var.failure_events_table_iam
}
