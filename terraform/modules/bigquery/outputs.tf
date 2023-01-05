/**
 * Copyright 2023 The Authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

output "bigquery_dataset_id" {
  description = "BigQuery dataset resource."
  value       = google_bigquery_dataset.default.id
}

output "bigquery_table_id" {
  description = "BigQuery table resource."
  value       = google_bigquery_table.default.id
}

output "bigquery_destination" {
  description = "BigQuery destination"
  value       = format("${google_bigquery_table.default.project}:${google_bigquery_table.default.dataset_id}.${google_bigquery_table.default.table_id}")
}

output "bigquery_default_views" {
  description = "Map of BigQuery default view resources."
  value = {
    for k, v in google_bigquery_table.default_views : v.table_id => v.id
  }
}
