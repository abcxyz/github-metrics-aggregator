resource "random_id" "default" {
  byte_length = 2
}

# GCS
resource "google_storage_bucket" "retry_lock" {
  project = data.google_project.default.project_id

  name                        = "${var.component_names.retry_name}-lock-${random_id.default.hex}"
  location                    = "US"
  public_access_prevention    = "enforced"
  uniform_bucket_level_access = true

  depends_on = [
    google_project_service.default["storage.googleapis.com"]
  ]
}

resource "google_storage_bucket_iam_member" "member" {
  bucket = google_storage_bucket.retry_lock.name
  role   = "roles/storage.admin"
  member = google_service_account.retry_run_service_account.member
}

# Cloud Scheduler

resource "google_service_account" "retry_invoker" {
  project = data.google_project.default.project_id

  account_id   = "${var.component_names.retry_name}-invoker-sa"
  display_name = "${var.component_names.retry_name}-invoker-sa Cloud Run Service Account"
}

// Give the scheduler invoker permission to the cloud run instance
resource "google_cloud_run_service_iam_member" "retry_invoker" {
  project = data.google_project.default.project_id

  location = var.region
  service  = module.retry_cloud_run.service_name
  role     = "roles/run.invoker"
  member   = google_service_account.retry_invoker.member

  depends_on = [
    google_cloud_scheduler_job.retry_scheduler
  ]
}

resource "google_cloud_scheduler_job" "retry_scheduler" {
  project = data.google_project.default.project_id

  name             = "${var.component_names.retry_name}-job"
  region           = var.region
  schedule         = var.cloud_scheduler_schedule_cron
  time_zone        = var.cloud_scheduler_timezone
  attempt_deadline = var.cloud_scheduler_deadline_duration
  retry_config {
    retry_count = var.cloud_scheduler_retry_limit
  }

  http_target {
    http_method = "GET"
    uri         = "${module.retry_cloud_run.url}/retry"
    oidc_token {
      audience              = module.retry_cloud_run.url
      service_account_email = google_service_account.retry_run_service_account.email
    }
  }

  depends_on = [
    google_project_service.default["cloudscheduler.googleapis.com"],
  ]
}

# Cloud Run 

resource "google_service_account" "retry_run_service_account" {
  project = data.google_project.default.project_id

  account_id   = "${var.component_names.retry_name}-run-sa"
  display_name = "${var.component_names.retry_name}-run-sa Cloud Run Service Account"
}

# This service is internal facing, and will only be invoked by the scheduler
module "retry_cloud_run" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=1d5d7f3f166679b02cd3f1ec615d287d6b7002dc"

  project_id = data.google_project.default.project_id

  name                  = var.component_names.retry_name
  region                = var.region
  image                 = var.retry_image
  ingress               = "all"
  secrets               = ["github-ssh-key"]
  service_account_email = google_service_account.retry_run_service_account.email
  service_iam = {
    admins     = var.service_iam.retry.admins
    developers = var.service_iam.retry.developers
    invokers = toset(
      concat(
        [
          google_service_account.retry_invoker,
        ],
        var.service_iam.retry.invokers,
      )
    )
  }
  envvars = {
    "PROJECT_ID" : data.google_project.default.project_id,
    "BIG_QUERY_ID" : var.bigquery_project_id,
    "BUCKET_URL" : google_storage_bucket.retry_lock.url,
    "EXECUTION_INTERVAL_MINUTES" : var.execution_interval_minutes,
    "EXECUTION_INTERVAL_CLOCK_SKEW_MS" : var.execution_interval_clock_skew_ms,
    "GITHUB_APP_ID" : var.github_app_id,
    "GITHUB_WEBHOOK_ID" : var.github_webhook_id,
  }
  secret_envvars = {
    "GITHUB_SSH_KEY" : {
      name : "github-ssh-key",
      version : "latest",
    },
  }

  depends_on = [
    google_storage_bucket.retry_lock,
  ]
}
