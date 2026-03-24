module "optimized_events_table" {
  source = "./modules/table"

  project_id = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id

  table_id = var.optimized_events_table_id
  schema = jsonencode([
    {
      "name" : "delivery_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "GUID from the GitHub webhook header (X-GitHub-Delivery)"
    },
    {
      "name" : "signature",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Signature from the GitHub webhook header (X-Hub-Signature-256)"
    },
    {
      "name" : "received",
      "type" : "TIMESTAMP",
      "mode" : "NULLABLE",
      "description" : "Timestamp for when an event is received"
    },
    {
      "name" : "event",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Event type from GitHub webhook header (X-GitHub-Event)"
    },
    {
      "name" : "enterprise_name",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Enterprise name from payload"
    },
    {
      "name" : "enterprise_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Enterprise ID from payload"
    },
    {
      "name" : "organization_name",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Organization login from payload"
    },
    {
      "name" : "organization_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Organization ID from payload"
    },
    {
      "name" : "repository_name",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Repository name from payload"
    },
    {
      "name" : "repository_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Repository ID from payload"
    },
    {
      "name" : "payload",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Event payload JSON string"
    }
  ])

  time_partitioning = {
    field = "received"
    type  = var.bigquery_events_partition_granularity
  }

  clustering = ["event", "enterprise_name", "organization_name", "repository_name"]
  iam        = var.events_table_iam
}

resource "google_bigquery_table_iam_member" "optimized_event_pubsub_agent_editor" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id

  table_id = module.optimized_events_table.table_id
  role     = "roles/bigquery.dataEditor"
  member   = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_bigquery_table_iam_member" "optimized_events_relay_sa_editor" {
  count = var.enable_relay_service ? 1 : 0

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id

  table_id = module.optimized_events_table.table_id
  role     = "roles/bigquery.dataEditor"
  member   = "serviceAccount:${var.relay_sub_service_account_email}"
}
