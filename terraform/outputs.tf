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
    service_account_name   = google_service_account.webhook_run_service_account.name
    service_account_email  = google_service_account.webhook_run_service_account.email
    service_account_member = google_service_account.webhook_run_service_account.member
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

output "bigquery_unique_events_view_id" {
  description = "BigQuery unique events view resource."
  value       = "unique_${var.events_table_id}"
}

output "bigquery_commit_review_status_table_id" {
  description = "BigQuery commit_review_status table resource."
  value       = try(module.commit_review_status[0].commit_review_status_table_id, null)
}

output "bigquery_event_views" {
  description = "BigQuery event view resources."
  value       = local.bq_event_views
}

output "bigquery_resource_views" {
  description = "BigQuery resource view resources."
  value       = local.bq_resource_views
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
  description = "The Cloud Run Job for artifact data. Only populated when var.leech.enabled is set."
  value = {
    job_id                 = try(module.leech[0].job_id, null)
    job_name               = try(module.leech[0].job_name, null)
    service_account_name   = try(module.leech[0].google_service_account.name, null)
    service_account_email  = try(module.leech[0].google_service_account.email, null)
    service_account_member = try(module.leech[0].google_service_account.member, null)
  }
}

output "commit_review_status_job" {
  description = "The Cloud Run Job for commit review status data. Only populated when var.commit_review_status.enabled is set."
  value = {
    job_id                 = try(module.commit_review_status[0].job_id, null)
    job_name               = try(module.commit_review_status[0].job_name, null)
    service_account_name   = try(module.commit_review_status[0].google_service_account.name, null)
    service_account_email  = try(module.commit_review_status[0].google_service_account.email, null)
    service_account_member = try(module.commit_review_status[0].google_service_account.member, null)
  }
}

output "retry_run_job" {
  description = "The Cloud Run Job for retry data."
  value = {
    job_id                 = module.retry_job.job_id
    job_name               = module.retry_job.job_name
    service_account_name   = module.retry_job.google_service_account.name
    service_account_email  = module.retry_job.google_service_account.email
    service_account_member = module.retry_job.google_service_account.member
  }
}

output "relay_run_service" {
  description = "The Cloud Run webhook service data."
  value = {
    service_id             = module.relay_cloud_run[0].service_id
    service_url            = module.relay_cloud_run[0].url
    service_name           = module.relay_cloud_run[0].service_name
    service_account_name   = google_service_account.relay_run_service_account[0].name
    service_account_email  = google_service_account.relay_run_service_account[0].email
    service_account_member = google_service_account.relay_run_service_account[0].member
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

output "scheduled_queries" {
  description = "The scheduled queries module. Only populated when var.enable_scheduled_queries is set."
  value       = try(module.scheduled_queries[0], null)
}

output "prstats_service_account" {
  description = "The created prstats service account."
  value       = google_service_account.prstats
}
