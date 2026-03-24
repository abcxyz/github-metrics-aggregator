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
    "secretmanager.googleapis.com"
  ])

  project = var.project_id

  service            = each.value
  disable_on_destroy = false
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
