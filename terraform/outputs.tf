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
  value       = module.gclb.external_ip_name
}

output "gclb_external_ip_address" {
  description = "The external IPv4 assigned to the global fowarding rule for the global load balancer fronting the webhook."
  value       = module.gclb.external_ip_address
}

output "run_service_url" {
  description = "The Cloud Run webhook service url."
  value       = module.webhook_cloud_run.url
}

output "run_service_id" {
  description = "The Cloud Run webhook service id."
  value       = module.webhook_cloud_run.service_id
}

output "run_service_name" {
  description = "The Cloud Run webhook service name."
  value       = module.webhook_cloud_run.service_name
}

output "run_service_account_name" {
  description = "Cloud Run service account name."
  value       = google_service_account.webhook_run_service_account.name
}

output "run_service_account_email" {
  description = "Cloud Run service account email."
  value       = google_service_account.webhook_run_service_account.email
}

output "run_service_account_member" {
  description = "Cloud Run service account email iam string."
  value       = google_service_account.webhook_run_service_account.member
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
  value       = google_bigquery_table.event_views["unique_events.sql"].table_id
}

output "bigquery_pubsub_destination" {
  description = "BigQuery PubSub destination"
  value       = format("${google_bigquery_table.events_table.project}:${google_bigquery_table.events_table.dataset_id}.${google_bigquery_table.events_table.table_id}")
}
