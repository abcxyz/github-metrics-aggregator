module "gclb" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/gclb_cloud_run_backend?ref=1d5d7f3f166679b02cd3f1ec615d287d6b7002dc"

  project_id = data.google_project.default.project_id

  name             = "webhook"
  run_service_name = module.webhook_cloud_run.service_name
  domain           = var.webhook_domain
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
    "PROJECT_ID" : data.google_project.default.project_id,
    "TOPIC_ID" : google_pubsub_topic.default.name,
    "RETRY_LIMIT" : var.event_delivery_retry_limit,
  }
  secret_envvars = {
    "WEBHOOK_SECRET" : {
      name : "github-webhook-secret",
      version : "latest",
    }
  }
}
