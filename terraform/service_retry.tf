# Copyright 2023 The Authors (see AUTHORS file)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

resource "random_id" "default" {
  byte_length = 2
}

# GCS
resource "google_storage_bucket" "retry_lock" {
  project = data.google_project.default.project_id

  name                        = "retry-lock-${random_id.default.hex}"
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

  account_id   = "retry-invoker-sa"
  display_name = "retry-invoker-sa Cloud Run Service Account"
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

  name             = "retry-job"
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
      service_account_email = google_service_account.retry_invoker.email
    }
  }

  depends_on = [
    google_project_service.default["cloudscheduler.googleapis.com"],
  ]
}

# Cloud Run

resource "google_service_account" "retry_run_service_account" {
  project = data.google_project.default.project_id

  account_id   = "${var.prefix_name}-retry-sa"
  display_name = "${var.prefix_name}-retry-sa Cloud Run Service Account"
}

# This service is internal facing, and will only be invoked by the scheduler
module "retry_cloud_run" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=45975889dcd5bae12b527a6bf9d05e082472d790"

  project_id = data.google_project.default.project_id

  name                  = "${var.prefix_name}-retry"
  region                = var.region
  image                 = var.image
  args                  = ["retry", "server"]
  ingress               = "all"
  secrets               = ["github-private-key"]
  service_account_email = google_service_account.retry_run_service_account.email
  service_iam = {
    admins     = toset(var.retry_service_iam.admins)
    developers = toset(concat(var.retry_service_iam.developers, [var.automation_service_account_member]))
    invokers   = toset(concat([google_service_account.retry_invoker.member], var.retry_service_iam.invokers))
  }
  envvars = {
    "BIG_QUERY_PROJECT_ID" : var.bigquery_project_id,
    "BUCKET_NAME" : google_storage_bucket.retry_lock.name,
    "CHECKPOINT_TABLE_ID" : google_bigquery_table.checkpoint_table.table_id,
    "EVENTS_TABLE_ID" : google_bigquery_table.events_table.table_id,
    "DATASET_ID" : google_bigquery_dataset.default.dataset_id
    "GITHUB_APP_ID" : var.github_app_id,
    "LOCK_TTL" : var.lock_ttl,
    "LOCK_TTL_CLOCK_SKEW" : var.lock_ttl_clock_skew,
    "PROJECT_ID" : data.google_project.default.project_id,
    "LOG_MODE" : var.log_mode
    "LOG_LEVEL" : var.log_level
  }
  secret_envvars = {
    "GITHUB_PRIVATE_KEY" : {
      name : "github-private-key",
      version : "latest",
    },
  }

  depends_on = [
    google_storage_bucket.retry_lock,
    google_service_account.retry_invoker,
  ]
}

# allow the ci service account to act as the retry cloud run service account
# this allows the ci service account to deploy new revisions for the cloud run
# service
resource "google_service_account_iam_member" "retry_run_sa_user" {
  service_account_id = google_service_account.retry_run_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = var.automation_service_account_member
}

# Alerting and Monitoring

module "retry_alerts" {
  count = var.retry_alerts.enabled ? 1 : 0

  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/alerts_cloud_run?ref=8728b6384c551d82d5cb09aafa2bf1816179f394"

  project_id = var.project_id

  notification_channels = [for x in values(google_monitoring_notification_channel.non_paging) : x.id]
  cloud_run_resource = {
    service_name = module.retry_cloud_run.service_name
  }
  runbook_urls = {
    forward_progress = local.forward_progress_runbook
    cpu              = local.cpu_runbook
  }

  built_in_forward_progress_indicators = merge(
    {
      "request-count" = { metric = "request_count", window = 2 * local.hour + 10 * local.minute },
    },
    var.retry_alerts.built_in_forward_progress_indicators,
  )

  built_in_cpu_indicators = merge(
    {
      "cpu-utilization" = { metric = "utilizations", window = 10 * local.minute, threshold : 0.8 },
    },
    var.retry_alerts.built_in_cpu_indicators,
  )

  log_based_text_indicators = merge(
    {
      "scaling-failure" = {
        log_name_suffix = local.log_name_suffix_request
        severity        = local.error_severity
        textPayload     = local.auto_scaling_failure
      },
      "failed-request" : {
        log_name_suffix = local.log_name_suffix_request
        severity        = local.error_severity
        textPayload     = local.request_failure
      },
    },
    var.retry_alerts.log_based_text_indicators
  )
}
