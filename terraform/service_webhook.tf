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
  webhook_service_window = 2 * local.hour + 10 * local.minute
}

module "gclb" {
  count = var.enable_webhook_gclb ? 1 : 0

  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/gclb_cloud_run_backend?ref=45975889dcd5bae12b527a6bf9d05e082472d790"

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
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=1467eaf0115f71613727212b0b51b3f99e699842"

  project_id = data.google_project.default.project_id

  name                  = "${var.prefix_name}-webhook"
  region                = var.region
  image                 = var.image
  args                  = ["webhook", "server"]
  ingress               = var.enable_webhook_gclb ? "internal-and-cloud-load-balancing" : "all"
  secrets               = ["github-webhook-secret"]
  service_account_email = google_service_account.webhook_run_service_account.email
  service_iam = {
    admins     = toset(var.webhook_service_iam.admins)
    developers = toset(concat(var.webhook_service_iam.developers, [var.automation_service_account_member]))
    invokers   = toset(var.webhook_service_iam.invokers)
  }
  envvars = {
    "BIG_QUERY_PROJECT_ID" : var.bigquery_project_id,
    "DATASET_ID" : var.dataset_id,
    "EVENTS_TABLE_ID" : var.events_table_id,
    "FAILURE_EVENTS_TABLE_ID" : var.failure_events_table_id,
    "PROJECT_ID" : data.google_project.default.project_id,
    "RETRY_LIMIT" : var.event_delivery_retry_limit,
    "EVENTS_TOPIC_ID" : google_pubsub_topic.default.name,
    "DLQ_EVENTS_TOPIC_ID" : google_pubsub_topic.dead_letter.name,
  }
  secret_envvars = {
    "GITHUB_WEBHOOK_SECRET" : {
      name : "github-webhook-secret",
      version : "latest",
    }
  }

  additional_service_annotations = { "run.googleapis.com/invoker-iam-disabled" : true }

  max_instances = var.webhook_max_instances
}

# allow the ci service account to act as the webhook cloud run service account
# this allows the ci service account to deploy new revisions for the cloud run
# service
resource "google_service_account_iam_member" "webhook_run_sa_user" {
  service_account_id = google_service_account.webhook_run_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = var.automation_service_account_member
}

# Alerting and Monitoring

module "webhook_alerts" {
  count = var.webhook_alerts.enabled ? 1 : 0

  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/alerts_cloud_run?ref=e73b4187678673411403a6f5f0969b1d17126e27"

  project_id = var.project_id

  notification_channels_non_paging            = [for x in values(google_monitoring_notification_channel.non_paging) : x.id]
  enable_built_in_forward_progress_indicators = false
  enable_built_in_container_indicators        = true
  enable_log_based_text_indicators            = true
  enable_log_based_json_indicators            = true

  cloud_run_resource = {
    service_name = module.webhook_cloud_run.service_name
  }
  runbook_urls = {
    forward_progress = local.forward_progress_runbook
    container_util   = local.container_util_runbook
    bad_request      = local.bad_request_runbook
    server_fault     = local.server_fault_runbook
    request_latency  = local.request_latency_runbook
  }

  built_in_container_util_indicators = merge(
    {
      "cpu" = {
        metric    = "container/cpu/utilizations"
        window    = local.webhook_service_window
        threshold = local.default_utilization_threshold_rate
        p_value   = 99
      },
      "memory" = {
        metric    = "container/memory/utilizations"
        window    = local.webhook_service_window
        threshold = local.default_utilization_threshold_rate
        p_value   = 99
      },
      "all-container-count" = {
        metric    = "container/instance_count"
        window    = local.webhook_service_window
        threshold = 10
      },
    },
    var.webhook_alerts.built_in_container_util_indicators,
  )

  log_based_text_indicators = merge(
    {
      "scaling-failure" = {
        log_name_suffix      = local.log_name_suffix_requests
        severity             = local.error_severity
        text_payload_message = local.auto_scaling_failure
        condition_threshold  = local.default_log_based_condition_threshold
      },
      "failed-request" : {
        log_name_suffix      = local.log_name_suffix_requests
        severity             = local.error_severity
        text_payload_message = local.request_failure,
        condition_threshold  = local.default_log_based_condition_threshold
      },
    },
    var.webhook_alerts.log_based_text_indicators
  )

  log_based_json_indicators = merge(
    {
      "write-failed-event-failure" : {
        log_name_suffix     = local.log_name_suffix_stdout
        severity            = local.error_severity
        additional_filters  = "jsonPayload.message=\"failed to call BigQuery\" AND jsonPayload.method=WriteFailureEvent"
        condition_threshold = local.default_log_based_condition_threshold
      }
    },
    var.webhook_alerts.log_based_json_indicators
  )

  service_5xx_configuration = {
    enabled   = true
    window    = 300
    threshold = 0
  }

  service_latency_configuration = {
    enabled      = true
    window       = local.webhook_service_window
    threshold_ms = local.default_threshold_ms
    p_value      = 99
  }

  service_max_conns_configuration = {
    enabled   = true
    window    = local.webhook_service_window
    threshold = 60
    p_value   = 99
  }
}
