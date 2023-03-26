module "gclb" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/gclb_cloud_run_backend?ref=142849110ddd7158796b399be1cf52802df46abd"

  project_id = data.google_project.default.project_id

  name             = "webhook"
  run_service_name = module.webhook_cloud_run.service_name
  domains          = var.webhook_domains
}

resource "google_service_account" "webhook_run_service_account" {
  project = data.google_project.default.project_id

  account_id   = "webhook-run-sa"
  display_name = "webhook-run-sa Cloud Run Service Account"
}

module "webhook_cloud_run" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=1d5d7f3f166679b02cd3f1ec615d287d6b7002dc"

  project_id = data.google_project.default.project_id

  name                  = "webhook"
  image                 = var.webhook_image
  ingress               = "internal-and-cloud-load-balancing"
  secrets               = ["github-webhook-secret"]
  service_account_email = google_service_account.webhook_run_service_account.email
  service_iam           = var.webhook_service_iam
  envvars = {
    "BIG_QUERY_PROJECT_ID" : var.bigquery_project_id,
    "DATASET_ID" : google_bigquery_dataset.default.dataset_id,
    "EVENTS_TABLE_ID" : google_bigquery_table.events_table.table_id,
    "FAILURE_EVENTS_TABLE_ID" : google_bigquery_table.failure_events_table.table_id,
    "PROJECT_ID" : data.google_project.default.project_id,
    "RETRY_LIMIT" : var.event_delivery_retry_limit,
    "TOPIC_ID" : google_pubsub_topic.default.name,
  }
  secret_envvars = {
    "WEBHOOK_SECRET" : {
      name : "github-webhook-secret",
      version : "latest",
    }
  }
}
