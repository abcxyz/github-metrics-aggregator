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

data "google_project" "project" {
  project_id = var.project_id
}

resource "google_bigquery_dataset" "default" {
  project    = var.project_id
  dataset_id = var.dataset_id
  location   = var.dataset_location
}

resource "google_bigquery_dataset_iam_binding" "bindings" {
  for_each   = var.dataset_iam
  project    = var.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = each.key
  members    = each.value
}

resource "google_bigquery_table" "default" {
  project             = var.project_id
  deletion_protection = true
  table_id            = var.table_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  schema              = file("${path.module}/bq_schema.json")

  lifecycle {
    ignore_changes = [
      last_modified_time,
      num_rows,
      num_bytes,
      num_long_term_bytes,
    ]
  }

  depends_on = [
    google_bigquery_dataset.default
  ]
}

resource "google_bigquery_table" "default_views" {
  for_each = fileset("${path.module}/views", "*")


  project       = var.project_id
  dataset_id    = google_bigquery_dataset.default.dataset_id
  friendly_name = replace(each.value, ".sql", "")
  table_id      = replace(each.value, ".sql", "")

  view {
    query = templatefile("${path.module}/views/${each.value}", {
      dataset_id = google_bigquery_table.default.dataset_id,
      table_id   = google_bigquery_table.default.table_id
    })
    use_legacy_sql = false
  }

  depends_on = [
    google_bigquery_dataset.default,
    google_bigquery_table.default
  ]
}

resource "google_bigquery_table_iam_binding" "bindings" {
  for_each   = var.table_iam
  project    = var.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = each.key
  members    = each.value
}

resource "google_bigquery_table_iam_member" "default_editor" {
  project    = google_bigquery_table.default.project
  dataset_id = google_bigquery_table.default.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-pubsub.iam.gserviceaccount.com"

  depends_on = [
    google_bigquery_table.default
  ]
}
