# Copyright 2026 The Authors (see AUTHORS file)
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

resource "google_bigquery_table" "default" {
  project = var.project_id

  deletion_protection = var.deletion_protection
  table_id            = var.table_id
  dataset_id          = var.dataset_id
  schema              = var.schema

  dynamic "time_partitioning" {
    for_each = var.time_partitioning != null ? [var.time_partitioning] : []
    content {
      type  = time_partitioning.value.type
      field = lookup(time_partitioning.value, "field", null)
    }
  }

  clustering = length(var.clustering) > 0 ? var.clustering : null
}

resource "google_bigquery_table_iam_member" "owners" {
  for_each = toset(lookup(var.iam, "owners", []))

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "editors" {
  for_each = toset(lookup(var.iam, "editors", []))

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "viewers" {
  for_each = toset(lookup(var.iam, "viewers", []))

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}
