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

locals {
  # time helpers
  second = 1
  minute = 60 * local.second
  hour   = 60 * local.minute
  day    = 24 * local.hour

  # runbooks
  runbook_url_prefix       = "https://github.com/abcxyz/github-metrics-aggregator/blob/main/docs/playbooks/alerts"
  forward_progress_runbook = "${local.runbook_url_prefix}/ForwardProgressFailed.md"
  container_util_runbook   = "${local.runbook_url_prefix}/ContainerUsage.md"
  bad_request_runbook      = "${local.runbook_url_prefix}/BadRequests.md"
  server_fault_runbook     = "${local.runbook_url_prefix}/ServerFaults.md"
  request_latency_runbook  = "${local.runbook_url_prefix}/RequestLatency.md"

  # cloud run error logs
  request_failure      = "The request failed because either the HTTP response was malformed or connection to the instance had an error."
  auto_scaling_failure = "The request was aborted because there was no available instance."

  error_severity = "ERROR"

  log_name_suffix_requests      = "requests"
  log_name_suffix_stderr        = "stderr"
  log_name_suffix_stdout        = "stdout"
  log_name_suffix_varlog_system = "varlog/system"

  default_threshold_ms               = 5 * 1000
  default_utilization_threshold_rate = 0.8

  default_log_based_condition_threshold = {
    window    = 5 * local.minute
    threshold = 0
  }
}

data "google_project" "default" {
  project_id = var.project_id
}

resource "google_project_service" "default" {
  for_each = toset([
    "bigquery.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "cloudscheduler.googleapis.com",
    "dataflow.googleapis.com",
    "datapipelines.googleapis.com",
    "logging.googleapis.com",
    "pubsub.googleapis.com",
    "stackdriver.googleapis.com",
    "storage.googleapis.com",
  ])

  project = var.project_id

  service            = each.value
  disable_on_destroy = false
}

module "leech" {
  count = var.leech.enabled ? 1 : 0

  source = "./modules/artifacts"

  project_id = var.project_id

  image                                = var.image
  dataset_id                           = google_bigquery_dataset.default.dataset_id
  leech_bucket_name                    = var.leech.bucket_name
  leech_bucket_location                = var.leech.bucket_location
  leech_table_id                       = var.leech.table_id
  leech_table_iam                      = var.leech.table_iam
  artifacts_job_iam                    = var.leech.job_iam
  events_table_id                      = var.events_table_id
  github_app_id                        = var.github_app_id
  github_install_id                    = var.github_install_id
  github_private_key_secret_id         = var.github_private_key_secret_id
  github_private_key_secret_version    = var.github_private_key_secret_version
  job_name                             = var.leech.job_name
  scheduler_cron                       = var.leech.scheduler_cron
  additional_env_vars                  = var.leech.job_additional_env_vars
  alerts_enabled                       = var.leech.alerts.enabled
  built_in_forward_progress_indicators = var.leech.alerts.built_in_forward_progress_indicators
  built_in_container_util_indicators   = var.leech.alerts.built_in_container_util_indicators
  notification_channels_non_paging     = [for x in values(google_monitoring_notification_channel.non_paging) : x.id]
}

# Allow the ci service account to act as the artifacts job service account.
# This allows the ci service account to deploy new revisions for the cloud run job.
resource "google_service_account_iam_member" "artifacts_job_sa_user" {
  count = var.leech.enabled ? 1 : 0

  service_account_id = module.leech[0].google_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = var.automation_service_account_member
}

module "commit_review_status" {
  count = var.commit_review_status.enabled ? 1 : 0

  source = "./modules/commit_review_status"

  project_id = var.project_id

  image                                = var.image
  dataset_id                           = google_bigquery_dataset.default.dataset_id
  github_app_id                        = var.github_app_id
  github_install_id                    = var.github_install_id
  github_private_key_secret_id         = var.github_private_key_secret_id
  github_private_key_secret_version    = var.github_private_key_secret_version
  push_events_table_id                 = module.metrics_views.bigquery_event_views["push_events.sql"]
  issues_table_id                      = module.metrics_views.bigquery_resource_views["issues.sql"]
  commit_review_status_table_id        = var.commit_review_status.table_id
  commit_review_status_table_iam       = var.commit_review_status.table_iam
  commit_review_status_job_iam         = var.commit_review_status.job_iam
  job_name                             = var.commit_review_status.job_name
  scheduler_cron                       = var.commit_review_status.scheduler_cron
  additional_env_vars                  = var.commit_review_status.job_additional_env_vars
  alerts_enabled                       = var.commit_review_status.alerts.enabled
  built_in_forward_progress_indicators = var.commit_review_status.alerts.built_in_forward_progress_indicators
  built_in_container_util_indicators   = var.commit_review_status.alerts.built_in_container_util_indicators
  notification_channels_non_paging     = [for x in values(google_monitoring_notification_channel.non_paging) : x.id]
}

# Allow the ci service account to act as the commit review status job service account.
# This allows the ci service account to deploy new revisions for the cloud run job.
resource "google_service_account_iam_member" "commit_review_status_job_sa_user" {
  count = var.commit_review_status.enabled ? 1 : 0

  service_account_id = module.commit_review_status[0].google_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = var.automation_service_account_member
}

resource "google_logging_project_bucket_config" "basic" {
  project = var.project_id

  location         = var.default_log_bucket_configuration.location
  retention_days   = var.default_log_bucket_configuration.retention_period
  enable_analytics = var.default_log_bucket_configuration.enable_analytics
  bucket_id        = "_Default"

  depends_on = [
    google_project_service.default["logging.googleapis.com"],
    google_project_service.default["stackdriver.googleapis.com"],
  ]
}
