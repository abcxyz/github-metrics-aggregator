locals {
  project_id                        = "REPLACE_PROJECT_ID"
  automation_service_account_member = "REPLACE_AUTOMATION_SERVICE_ACCOUNT_MEMBER"
}

module "REPLACE_MODULE_NAME" {
  source = "git::https://github.com/abcxyz/github-metrics-aggregator.git//terraform?ref=v0.0.15"

  project_id = local.project_id

  image                             = "gcr.io/cloudrun/placeholder@sha256:f1586972ac147796d60ee2c5a0d6cc78067fc862b6d715d6d2a96826455c3423"
  bigquery_project_id               = local.project_id
  automation_service_account_member = local.automation_service_account_member
  webhook_domains                   = ["REPLACE_DOMAIN"]
  github_app_id                     = "REPLACE_GITHUB_APP_ID"
  log_level                         = "info"
  retry_service_iam = {
    admins     = []
    developers = []
    invokers   = []
  }
  webhook_service_iam = {
    admins     = []
    developers = []
    invokers   = ["allUsers"] # public access, called by github webhook
  }
  dataset_iam = {
    owners  = []
    editors = []
    viewers = []
  }

  depends_on = [
    google_project_iam_member.REPLACE_MODULE_NAME_actuators
  ]
}
