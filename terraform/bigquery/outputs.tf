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
  value       = module.events_table.table_id
}

output "raw_events_table_id" {
  description = "The ID of the BigQuery table for raw events."
  value       = module.raw_events_table.table_id
}

output "optimized_events_table_id" {
  description = "The ID of the BigQuery table for optimized events."
  value       = module.optimized_events_table.table_id
}

output "checkpoint_table_id" {
  description = "The ID of the BigQuery table for checkpoints."
  value       = module.checkpoint_table.table_id
}

output "failure_events_table_id" {
  description = "The ID of the BigQuery table for failure events."
  value       = module.failure_events_table.table_id
}

output "artifacts_table_id" {
  description = "The ID of the BigQuery table for artifacts status."
  value       = module.artifacts_table.table_id
}

output "commit_review_status_table_id" {
  description = "The ID of the BigQuery table for commit review status."
  value       = module.commit_review_status_table.table_id
}
