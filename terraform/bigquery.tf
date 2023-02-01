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

resource "google_bigquery_dataset" "default" {
  project    = data.google_project.default.project_id
  dataset_id = var.dataset_id
  location   = var.dataset_location

  depends_on = [
    google_project_service.default["bigquery.googleapis.com"]
  ]
}

resource "google_bigquery_dataset_iam_binding" "owners" {
  project    = data.google_project.default.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.dataOwner"
  members    = toset(var.dataset_iam.owners)
}

resource "google_bigquery_dataset_iam_binding" "editors" {
  project    = data.google_project.default.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.dataEditor"
  members    = toset(var.dataset_iam.editors)
}

resource "google_bigquery_dataset_iam_binding" "viewers" {
  project    = data.google_project.default.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.dataViewer"
  members    = toset(var.dataset_iam.viewers)
}

resource "google_bigquery_table" "default" {
  project             = data.google_project.default.project_id
  deletion_protection = true
  table_id            = var.table_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  schema = jsonencode([
    {
      "name" : "delivery_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "GUID from the GitHub webhook header (X-GitHub-Delivery)"
    },
    {
      "name" : "signature",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Signature from the GitHub webhook header (X-Hub-Signature-256)"
    },
    {
      "name" : "received",
      "type" : "TIMESTAMP",
      "mode" : "NULLABLE",
      "description" : "Timestamp for when an event is received"
    },
    {
      "name" : "event",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Event type from GitHub webhook header (X-GitHub-Event)"
    },
    {
      "name" : "payload",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Event payload JSON string"
    }
  ])
}

resource "google_bigquery_table" "default_views" {
  for_each = fileset("${path.module}/data/bq_views", "*")

  project             = data.google_project.default.project_id
  deletion_protection = true
  dataset_id          = google_bigquery_dataset.default.dataset_id
  friendly_name       = replace(each.value, ".sql", "")
  table_id            = replace(each.value, ".sql", "")

  view {
    query = templatefile("${path.module}/views/${each.value}", {
      dataset_id = google_bigquery_table.default.dataset_id,
      table_id   = google_bigquery_table.default.table_id
    })
    use_legacy_sql = false
  }
}

resource "google_bigquery_table_iam_binding" "owners" {
  project    = data.google_project.default.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = "roles/bigquery.dataOwner"
  members    = toset(var.table_iam.owners)
}

resource "google_bigquery_table_iam_binding" "editors" {
  project    = data.google_project.default.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = "roles/bigquery.dataEditor"
  members = toset(
    concat(
      ["serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"],
      var.table_iam.editors
    )
  )
}

resource "google_bigquery_table_iam_binding" "viewers" {
  project    = data.google_project.default.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = "roles/bigquery.dataViewer"
  members    = toset(var.table_iam.viewers)
}
