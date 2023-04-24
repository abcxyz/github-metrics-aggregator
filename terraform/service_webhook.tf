module "gclb" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/gclb_cloud_run_backend?ref=e4e2ad79ae2cf833540f890ac8241220144057d0"

  project_id = data.google_project.default.project_id

  name             = "${var.prefix_name}-webhook"
  run_service_name = module.webhook_cloud_run.service_name
  domains          = var.webhook_domains
}

resource "google_service_account" "webhook_run_service_account" {
  project = data.google_project.default.project_id

  account_id   = "${var.prefix_name}-webhook-sa"
  display_name = "${var.prefix_name}-webhook-sa Cloud Run Service Account"
}

module "webhook_cloud_run" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=e4e2ad79ae2cf833540f890ac8241220144057d0"

  project_id = data.google_project.default.project_id

  name                  = "${var.prefix_name}-webhook"
  region                = var.region
  image                 = var.image
  args                  = ["webhook", "server"]
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
