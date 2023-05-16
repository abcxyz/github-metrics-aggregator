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

# Enable all services to make queries against BQ tables
# https://cloud.google.com/bigquery/docs/access-control#bigquery.jobUser

resource "google_project_iam_member" "webhook_job_user" {
  project = data.google_project.default.project_id

  role   = "roles/bigquery.jobUser"
  member = google_service_account.webhook_run_service_account.member
}

resource "google_project_iam_member" "retry_job_user" {
  project = data.google_project.default.project_id

  role   = "roles/bigquery.jobUser"
  member = google_service_account.retry_run_service_account.member
}

# Dataset / IAM
resource "google_bigquery_dataset" "default" {
  project = data.google_project.default.project_id

  dataset_id = var.dataset_id
  location   = var.dataset_location

  depends_on = [
    google_project_service.default["bigquery.googleapis.com"]
  ]
}

resource "google_bigquery_dataset_iam_member" "default_owners" {
  for_each = toset(var.dataset_iam.owners)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_dataset_iam_member" "default_editors" {
  for_each = toset(var.dataset_iam.editors)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_dataset_iam_member" "default_viewers" {
  for_each = toset(var.dataset_iam.viewers)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

resource "google_bigquery_dataset_iam_member" "retry_dataset_metadata_viewer" {
  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.metadataViewer"
  member     = google_service_account.retry_run_service_account.member
}

resource "google_bigquery_dataset_iam_member" "webhook_dataset_metadata_viewer" {
  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.metadataViewer"
  member     = google_service_account.webhook_run_service_account.member
}

# Event Table / IAM

resource "google_bigquery_table" "events_table" {
  project = data.google_project.default.project_id

  deletion_protection = true
  table_id            = var.events_table_id
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

resource "google_bigquery_table_iam_member" "event_owners" {
  for_each = toset(var.events_table_iam.owners)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "event_editors" {
  for_each = toset(var.events_table_iam.editors)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "event_pubsub_agent_editor" {
  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_bigquery_table_iam_member" "event_webhook_editor" {
  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = google_service_account.webhook_run_service_account.member
}

resource "google_bigquery_table_iam_member" "event_viewers" {
  for_each = toset(var.events_table_iam.viewers)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

# Checkpoint Table / IAM

resource "google_bigquery_table" "checkpoint_table" {
  project = data.google_project.default.project_id

  deletion_protection = true
  table_id            = var.checkpoint_table_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  schema = jsonencode([
    {
      "name" : "delivery_id",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GUID that represents the last successfully redelivered event sent to GitHub."
    },
    {
      "name" : "created",
      "type" : "TIMESTAMP",
      "mode" : "REQUIRED",
      "description" : "Timestamp for when the checkpoint record was created."
    },
  ])
}

resource "google_bigquery_table_iam_member" "checkpoint_owners" {
  for_each = toset(var.checkpoint_table_iam.owners)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.checkpoint_table.table_id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "checkpoint_editors" {
  for_each = toset(var.checkpoint_table_iam.editors)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.checkpoint_table.table_id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "checkpoint_retry_editor" {
  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.checkpoint_table.table_id
  role       = "roles/bigquery.dataEditor"
  member     = google_service_account.retry_run_service_account.member
}

resource "google_bigquery_table_iam_member" "checkpoint_viewers" {
  for_each = toset(var.checkpoint_table_iam.viewers)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.checkpoint_table.table_id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

# Failure Events Table / IAM

resource "google_bigquery_table" "failure_events_table" {
  project = data.google_project.default.project_id

  deletion_protection = true
  table_id            = var.failure_events_table_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  schema = jsonencode([
    {
      "name" : "delivery_id",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GUID that represents the failed event."
    },
    {
      "name" : "created",
      "type" : "TIMESTAMP",
      "mode" : "REQUIRED",
      "description" : "Timestamp of when the event fails to process."
    },
  ])
}

resource "google_bigquery_table_iam_member" "failure_events_owners" {
  for_each = toset(var.failure_events_table_iam.owners)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.failure_events_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "failure_events_editors" {
  for_each = toset(var.failure_events_table_iam.editors)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.failure_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "failure_events_webhook_editor" {
  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.failure_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = google_service_account.webhook_run_service_account.member
}

resource "google_bigquery_table_iam_member" "failure_events_viewers" {
  for_each = toset(var.failure_events_table_iam.viewers)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.failure_events_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

// Unique Events View 

resource "google_bigquery_table" "event_views" {
  for_each = fileset("${path.module}/data/bq_views/events", "*")

  project = data.google_project.default.project_id

  deletion_protection = true
  dataset_id          = google_bigquery_dataset.default.dataset_id
  friendly_name       = replace(each.value, ".sql", "")
  table_id            = replace(each.value, ".sql", "")
  view {
    query = templatefile("${path.module}/data/bq_views/events/${each.value}", {
      dataset_id = google_bigquery_dataset.default.dataset_id
      table_id   = google_bigquery_table.events_table.table_id
    })
    use_legacy_sql = false
  }

  depends_on = [
    google_bigquery_table.events_table
  ]
}
