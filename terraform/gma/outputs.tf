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

output "gclb_external_ip_name" {
  description = "The external IPv4 name assigned to the global fowarding rule for the global load balancer fronting the webhook."
  value       = try(module.gclb[0].external_ip_name, null)
}

output "gclb_external_ip_address" {
  description = "The external IPv4 assigned to the global fowarding rule for the global load balancer fronting the webhook."
  value       = try(module.gclb[0].external_ip_address, null)
}

output "webhook_run_service" {
  description = "The Cloud Run webhook service data."
  value = {
    service_id             = module.webhook_cloud_run.service_id
    service_url            = module.webhook_cloud_run.url
    service_name           = module.webhook_cloud_run.service_name
    service_account_name   = local.compute_service_account_name
    service_account_email  = local.compute_service_account_email
    service_account_member = local.compute_service_account_member
  }
}

output "bigquery_dataset_id" {
  description = "BigQuery dataset resource."
  value       = var.dataset_id
}

output "bigquery_events_table_id" {
  description = "BigQuery events table resource."
  value       = var.events_table_id
}

output "bigquery_checkpoint_table_id" {
  description = "BigQuery checkpoint table resource."
  value       = var.checkpoint_table_id
}

output "bigquery_failure_events_table_id" {
  description = "BigQuery failure_events table resource."
  value       = var.failure_events_table_id
}


output "bigquery_commit_review_status_table_id" {
  description = "BigQuery commit_review_status table resource."
  value       = var.commit_review_status.table_id
}



output "bigquery_pubsub_destination" {
  description = "BigQuery PubSub destination"
  value       = "${var.project_id}:${var.dataset_id}.${var.events_table_id}"
}

output "github_metrics_looker_studio_report_link" {
  description = "The Looker Studio Linking API link for connecting the data sources for the GitHub Metrics dashboard."
  value = var.github_metrics_dashboard.enabled ? join("",
    [
      "https://lookerstudio.google.com/reporting/create",
      "?c.reportId=${var.github_metrics_dashboard.looker_report_id}",
      "&r.reportName=GitHub%20Metrics",
      "&ds.ds0.keepDatasourceName",
      "&ds.ds0.connector=bigQuery",
      "&ds.ds0.refreshFields",
      "&ds.ds0.projectId=${var.project_id}",
      "&ds.ds0.type=TABLE",
      "&ds.ds0.datasetId=${var.dataset_id}",
      "&ds.ds0.tableId=pull_request_reviews",
      "&ds.ds2.keepDatasourceName",
      "&ds.ds2.connector=bigQuery",
      "&ds.ds2.refreshFields",
      "&ds.ds2.projectId=${var.project_id}",
      "&ds.ds2.type=TABLE",
      "&ds.ds2.datasetId=${var.dataset_id}",
      "&ds.ds2.tableId=pull_requests",
    ]
  ) : null
}

output "artifacts_job" {
  description = "The Cloud Run Job for artifact data. Only populated when var.artifacts.enabled is set."
  value = {
    job_id                 = try(google_cloud_run_v2_job.artifacts[0].id, null)
    job_name               = try(google_cloud_run_v2_job.artifacts[0].name, null)
    service_account_name   = var.artifacts.enabled ? local.compute_service_account_name : null
    service_account_email  = var.artifacts.enabled ? local.compute_service_account_email : null
    service_account_member = var.artifacts.enabled ? local.compute_service_account_member : null
  }
}


output "commit_review_status_job" {
  description = "The Cloud Run Job for commit review status data. Only populated when var.commit_review_status.enabled is set."
  value = {
    job_id                 = try(google_cloud_run_v2_job.commit_review_status[0].id, null)
    job_name               = try(google_cloud_run_v2_job.commit_review_status[0].name, null)
    service_account_name   = var.commit_review_status.enabled ? local.compute_service_account_name : null
    service_account_email  = var.commit_review_status.enabled ? local.compute_service_account_email : null
    service_account_member = var.commit_review_status.enabled ? local.compute_service_account_member : null
  }
}


output "retry_run_job" {
  description = "The Cloud Run Job for retry data."
  value = {
    job_id                 = google_cloud_run_v2_job.retry.id
    job_name               = google_cloud_run_v2_job.retry.name
    service_account_name   = local.compute_service_account_name
    service_account_email  = local.compute_service_account_email
    service_account_member = local.compute_service_account_member
  }
}


output "relay_run_service" {
  description = "The Cloud Run webhook service data."
  value = {
    service_id             = module.relay_cloud_run[0].service_id
    service_url            = module.relay_cloud_run[0].url
    service_name           = module.relay_cloud_run[0].service_name
    service_account_name   = var.enable_relay_service ? local.compute_service_account_name : null
    service_account_email  = var.enable_relay_service ? local.compute_service_account_email : null
    service_account_member = var.enable_relay_service ? local.compute_service_account_member : null
  }
}

output "pubsub_schema_id" {
  description = "The ID of the Pub/Sub schema for events."
  value       = google_pubsub_schema.default.id
}

output "relay_pubsub_schema_id" {
  description = "The ID of the Pub/Sub schema for enriched relay events."
  value       = google_pubsub_schema.enriched.id
}
