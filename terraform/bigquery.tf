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

# Events table with clustering and partitioning / IAM
resource "google_bigquery_table" "raw_events_table" {
  project = data.google_project.default.project_id

  deletion_protection = true
  table_id            = var.raw_events_table_id
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
      "type" : "JSON",
      "mode" : "NULLABLE",
      "description" : "Event payload JSON"
    }
  ])

  time_partitioning {
    field = "received"
    type  = var.bigquery_events_partition_granularity
  }

  clustering = ["event", "received"]
}

resource "google_bigquery_table_iam_member" "raw_event_owners" {
  for_each = toset(var.events_table_iam.owners)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "raw_event_editors" {
  for_each = toset(var.events_table_iam.editors)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "raw_event_pubsub_agent_editor" {
  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_bigquery_table_iam_member" "raw_event_webhook_editor" {
  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = google_service_account.webhook_run_service_account.member
}

resource "google_bigquery_table_iam_member" "raw_event_viewers" {
  for_each = toset(var.events_table_iam.viewers)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
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

# Unique Events - deduplicate rows

resource "google_bigquery_table" "unique_events_view" {
  project = data.google_project.default.project_id

  deletion_protection = false
  dataset_id          = google_bigquery_dataset.default.dataset_id
  friendly_name       = "unique_${var.events_table_id}"
  table_id            = "unique_${var.events_table_id}"
  view {
    query          = <<EOT
    SELECT
      delivery_id,
      signature,
      received,
      event,
      payload,
      JSON_VALUE(payload, "$.organization.login") organization,
      SAFE_CAST(JSON_VALUE(payload, "$.organization.id") AS INT64) organization_id,
      JSON_VALUE(payload, "$.repository.full_name") repository_full_name,
      SAFE_CAST(JSON_VALUE(payload, "$.repository.id") AS INT64) repository_id,
      JSON_VALUE(payload, "$.repository.name") repository,
      JSON_VALUE(payload, "$.repository.visibility") repository_visibility,
      JSON_VALUE(payload, "$.sender.login") sender,
      SAFE_CAST(JSON_VALUE(payload, "$.sender.id") AS INT64) sender_id,
    FROM (
       SELECT ROW_NUMBER() OVER (PARTITION BY delivery_id ORDER BY received DESC) as row_id, *
       FROM `${google_bigquery_table.events_table.project}.${google_bigquery_table.events_table.dataset_id}.${google_bigquery_table.events_table.table_id}`)
    WHERE row_id = 1;
    EOT
    use_legacy_sql = false
  }

  depends_on = [
    google_bigquery_table.events_table
  ]
}

# Unique Events -deduplicate rows but as a table function
resource "google_bigquery_routine" "unique_events_by_date_type" {
  project = data.google_project.default.project_id

  dataset_id      = google_bigquery_dataset.default.dataset_id
  routine_id      = "unique_events_by_date_type"
  routine_type    = "TABLE_VALUED_FUNCTION"
  language        = "SQL"
  definition_body = <<EOT
    SELECT
      delivery_id,
      signature,
      received,
      event,
      payload,
      LAX_STRING(payload.organization.login) organization,
      SAFE.INT64(payload.organization.id) organization_id,
      LAX_STRING(payload.repository.full_name) repository_full_name,
      SAFE.INT64(payload.repository.id) repository_id,
      LAX_STRING(payload.repository.name) repository,
      LAX_STRING(payload.repository.visibility) repository_visibility,
      LAX_STRING(payload.sender.login) sender,
      SAFE.INT64(payload.sender.id) sender_id,
    FROM ( SELECT ROW_NUMBER() OVER (PARTITION BY delivery_id ORDER BY received DESC) as row_id, *
      FROM
      `${google_bigquery_table.raw_events_table.project}.${google_bigquery_table.raw_events_table.dataset_id}.${google_bigquery_table.raw_events_table.table_id}`
      WHERE
        received >= startTimestamp
        AND received <= endTimestamp
        AND event = eventTypeFilter
      )
    WHERE row_id = 1
    EOT

  arguments {
    name      = "startTimestamp"
    data_type = jsonencode({ typeKind : "TIMESTAMP" })
  }
  arguments {
    name      = "endTimestamp"
    data_type = jsonencode({ typeKind : "TIMESTAMP" })
  }
  arguments {
    name      = "eventTypeFilter"
    data_type = jsonencode({ typeKind : "STRING" })
  }
}

# Create all the required metrics views for dashboards
module "metrics_views" {
  source = "./modules/bigquery_metrics_views"

  project_id = data.google_project.default.project_id

  dataset_id    = google_bigquery_dataset.default.dataset_id
  base_table_id = google_bigquery_table.unique_events_view.table_id
  base_tvf_id   = google_bigquery_routine.unique_events_by_date_type.routine_id
}

module "leech" {
  count = var.leech.enabled ? 1 : 0

  source = "./modules/leech"

  project_id = var.project_id

  dataset_id            = google_bigquery_dataset.default.dataset_id
  leech_bucket_name     = var.leech.bucket_name
  leech_bucket_location = var.leech.bucket_location
  leech_table_id        = var.leech.table_id
  leech_table_iam       = var.leech.table_iam
}

module "commit_review_status" {
  count = var.commit_review_status.enabled ? 1 : 0

  source = "./modules/commit_review_status"

  project_id = var.project_id

  dataset_id                     = google_bigquery_dataset.default.dataset_id
  commit_review_status_table_id  = var.commit_review_status.table_id
  commit_review_status_table_iam = var.commit_review_status.table_iam
}

module "invocation_comment" {
  count = var.invocation_comment.enabled ? 1 : 0

  source = "./modules/invocation_comment"

  project_id = var.project_id

  dataset_id                   = google_bigquery_dataset.default.dataset_id
  invocation_comment_table_id  = var.invocation_comment.table_id
  invocation_comment_table_iam = var.invocation_comment.table_iam
}

module "pr_stats_dashboard" {
  count = var.pr_stats_dashboard.enabled ? 1 : 0

  source = "./modules/pr_stats_dashboard"

  project_id       = var.project_id
  dataset_id       = google_bigquery_dataset.default.dataset_id
  looker_report_id = var.pr_stats_dashboard.looker_report_id
  viewers          = var.pr_stats_dashboard.viewers
}
