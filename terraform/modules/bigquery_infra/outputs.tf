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

output "dataset_id" {
  description = "The ID of the BigQuery dataset."
  value       = google_bigquery_dataset.default.dataset_id
}

output "events_table_id" {
  description = "The ID of the BigQuery table for events."
  value       = google_bigquery_table.events_table.table_id
}

output "raw_events_table_id" {
  description = "The ID of the BigQuery table for raw events."
  value       = google_bigquery_table.raw_events_table.table_id
}

output "checkpoint_table_id" {
  description = "The ID of the BigQuery table for checkpoints."
  value       = google_bigquery_table.checkpoint_table.table_id
}

output "failure_events_table_id" {
  description = "The ID of the BigQuery table for failure events."
  value       = google_bigquery_table.failure_events_table.table_id
}

output "bigquery_event_views" {
  description = "BigQuery event view resources."
  value       = module.metrics_views.bigquery_event_views
}

output "bigquery_resource_views" {
  description = "BigQuery resource view resources."
  value       = module.metrics_views.bigquery_resource_views
}

output "events_dashboard_mv_table_id" {
  description = "The ID of the BigQuery Materialized View for CL Stats."
  value       = google_bigquery_table.events_dashboard_mv.table_id
}
