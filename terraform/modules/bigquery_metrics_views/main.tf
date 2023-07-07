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

resource "google_bigquery_table" "event_views" {
  for_each = fileset("${path.module}/data/bq_views/events", "*.sql")

  project = var.project_id

  deletion_protection = false
  dataset_id          = var.dataset_id
  friendly_name       = replace(each.value, ".sql", "")
  table_id            = replace(each.value, ".sql", "")
  view {
    query = templatefile("${path.module}/data/bq_views/events/${each.value}", {
      dataset_id = var.dataset_id
      table_id   = var.base_table_id
    })
    use_legacy_sql = false
  }
}

resource "google_bigquery_table" "resource_views" {
  for_each = fileset("${path.module}/data/bq_views/resources", "*.sql")

  project = var.project_id

  deletion_protection = false
  dataset_id          = var.dataset_id
  friendly_name       = replace(each.value, ".sql", "")
  table_id            = replace(each.value, ".sql", "")
  view {
    query = templatefile("${path.module}/data/bq_views/resources/${each.value}", {
      dataset_id = var.dataset_id
    })
    use_legacy_sql = false
  }

  # Must wait for all events tables before creating views on them
  depends_on = [google_bigquery_table.event_views]
}
