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

output "retry_run_service" {
  description = "The Cloud Run retry service data."
  value = {
    service_id             = module.retry_cloud_run.service_id
    service_url            = module.retry_cloud_run.url
    service_name           = module.retry_cloud_run.service_name
    service_account_name   = google_service_account.retry_run_service_account.name
    service_account_email  = google_service_account.retry_run_service_account.email
    service_account_member = google_service_account.retry_run_service_account.member
  }
}

output "bigquery_dataset_id" {
  description = "BigQuery dataset resource."
  value       = google_bigquery_dataset.default.dataset_id
}

output "bigquery_events_table_id" {
  description = "BigQuery events table resource."
  value       = google_bigquery_table.events_table.table_id
}

output "bigquery_checkpoint_table_id" {
  description = "BigQuery checkpoint table resource."
  value       = google_bigquery_table.checkpoint_table.table_id
}

output "bigquery_failure_events_table_id" {
  description = "BigQuery failure_events table resource."
  value       = google_bigquery_table.failure_events_table.table_id
}

output "bigquery_unique_events_view_id" {
  description = "BigQuery unique events view resource."
  value       = google_bigquery_table.unique_events_view.table_id
}

output "bigquery_commit_review_status_table_id" {
  description = "BigQuery commit_review_status table resource."
  value       = var.commit_review_status.enabled ? module.commit_review_status[0].commit_review_status_table_id : null
}

output "bigquery_event_views" {
  description = "BigQuery event view resources."
  value       = module.metrics_views.bigquery_event_views
}

output "bigquery_resource_views" {
  description = "BigQuery resource view resources."
  value       = module.metrics_views.bigquery_resource_views
}

output "bigquery_pubsub_destination" {
  description = "BigQuery PubSub destination"
  value       = format("${google_bigquery_table.events_table.project}:${google_bigquery_table.events_table.dataset_id}.${google_bigquery_table.events_table.table_id}")
}

output "pr_stats_looker_studio_report_link" {
  description = "The Looker Studio Linking API link for connecting the data sources for the PR Stats dashboard."
  value = var.pr_stats_dashboard.enabled ? join("",
    [
      "https://lookerstudio.google.com/reporting/create",
      "?c.reportId=${var.pr_stats_dashboard.looker_report_id}",
      "&r.reportName=GitHub%20Metrics",
      "&ds.ds0.keepDatasourceName",
      "&ds.ds0.connector=bigQuery",
      "&ds.ds0.refreshFields",
      "&ds.ds0.projectId=${var.project_id}",
      "&ds.ds0.type=TABLE",
      "&ds.ds0.datasetId=${google_bigquery_dataset.default.dataset_id}",
      "&ds.ds0.tableId=pull_requests",
      "&ds.ds2.keepDatasourceName",
      "&ds.ds2.connector=bigQuery",
      "&ds.ds2.refreshFields",
      "&ds.ds2.projectId=${var.project_id}",
      "&ds.ds2.type=TABLE",
      "&ds.ds2.datasetId=${google_bigquery_dataset.default.dataset_id}",
      "&ds.ds2.tableId=pull_request_reviews",
    ]
  ) : null
}
