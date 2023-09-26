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
    field = "received" # TODO: would we rather extract an actual time from the payload?
    type  = var.bigquery_events_partition_granularity
  }

  clustering = ["event", "received"] # TODO: would we rather use the actual time extracted from event?
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
      SAFE_CAST(JSON_QUERY(payload, "$.repository.id") AS INT64) repository_id,
      JSON_VALUE(payload, "$.repository.name") repository,
      JSON_VALUE(payload, "$.repository.visibility") repository_visibility,
      JSON_VALUE(payload, "$.sender.login") sender,
      SAFE_CAST(JSON_QUERY(payload, "$.sender.id") AS INT64) sender_id,
    FROM
      `${google_bigquery_dataset.default.dataset_id}.${google_bigquery_table.events_table.table_id}`
    GROUP BY
      delivery_id,
      signature,
      received,
      event,
      payload,
      JSON_VALUE(payload, "$.organization.login"),
      SAFE_CAST(JSON_VALUE(payload, "$.organization.id") AS INT64),
      JSON_VALUE(payload, "$.repository.full_name"),
      SAFE_CAST(JSON_QUERY(payload, "$.repository.id") AS INT64),
      JSON_VALUE(payload, "$.repository.name"),
      JSON_VALUE(payload, "$.repository.visibility"),
      JSON_VALUE(payload, "$.sender.login"),
      SAFE_CAST(JSON_QUERY(payload, "$.sender.id") AS INT64)
    EOT
    use_legacy_sql = false
  }

  depends_on = [
    google_bigquery_table.events_table
  ]
}

# Unique Events -deduplicate rows but as a table function
resource "google_bigquery_routine" "unique_events_by_date_type" {
  dataset_id      = google_bigquery_dataset.default.dataset_id
  routine_id      = "unique_events_by_date_type"
  definition_body = <<EOT
    SELECT
      delivery_id,
      signature,
      received,
      event,
      payload,
      JSON_VALUE(payload, "$.organization.login") organization,
      SAFE_CAST(JSON_VALUE(payload, "$.organization.id") AS INT64) organization_id,
      JSON_VALUE(payload, "$.repository.full_name") repository_full_name,
      SAFE_CAST(JSON_QUERY(payload, "$.repository.id") AS INT64) repository_id,
      JSON_VALUE(payload, "$.repository.name") repository,
      JSON_VALUE(payload, "$.repository.visibility") repository_visibility,
      JSON_VALUE(payload, "$.sender.login") sender,
      SAFE_CAST(JSON_QUERY(payload, "$.sender.id") AS INT64) sender_id,
    FROM
      `${google_bigquery_dataset.default.dataset_id}.${google_bigquery_table.events_table.table_id}`
    WHERE
      received >= start
      AND received <= end
      AND eventTypeFilter == event
    GROUP BY
      delivery_id,
      signature,
      received,
      event,
      payload,
      JSON_VALUE(payload, "$.organization.login"),
      SAFE_CAST(JSON_VALUE(payload, "$.organization.id") AS INT64),
      JSON_VALUE(payload, "$.repository.full_name"),
      SAFE_CAST(JSON_QUERY(payload, "$.repository.id") AS INT64),
      JSON_VALUE(payload, "$.repository.name"),
      JSON_VALUE(payload, "$.repository.visibility"),
      JSON_VALUE(payload, "$.sender.login"),
      SAFE_CAST(JSON_QUERY(payload, "$.sender.id") AS INT64)
    EOT

  arguments = [
    {
      name      = "start"
      data_type = jsonencode({ typeKind : "TIMESTAMP" })
    },
    {
      name      = "end"
      data_type = jsonencode({ typeKind : "TIMESTAMP" })
    },
    {
      name      = "eventTypeFilter"
      data_type = jsonencode({ typeKind : "STRING" })
    }
  ]
}

# Create all the required metrics views for dashboards
module "metrics_views" {
  source = "./modules/bigquery_metrics_views"

  project_id = data.google_project.default.project_id

  dataset_id    = google_bigquery_dataset.default.dataset_id
  base_table_id = google_bigquery_table.unique_events_view.table_id
}

# Leech Table / IAM

resource "google_bigquery_table" "leech_table" {
  project = data.google_project.default.project_id

  deletion_protection = false
  table_id            = var.leech_table_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  schema = jsonencode([
    {
      "name" : "delivery_id",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GUID that represents the event that was ingested."
    },
    {
      "name" : "processed_at",
      "type" : "TIMESTAMP",
      "mode" : "REQUIRED",
      "description" : "Timestamp of when the event was processed."
    },
    {
      "name" : "status",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The status of the log ingestion."
    },
    {
      "name" : "workflow_uri",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The original workflow uri that trigger the ingestion."
    },
    {
      "name" : "logs_uri",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The GCS uri of the logs."
    },
    {
      "name" : "github_actor",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GitHub user that triggered the workflow event."
    },
    {
      "name" : "organization_name",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GitHub organization name."
    },
    {
      "name" : "repository_name",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "GitHub repository name."
    },
    {
      "name" : "repository_slug",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "Combined org/repo_name of the repository."
    },
    {
      "name" : "job_name",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "Apache Beam job name of the pipeline that processed this event."
    },
  ])
}

resource "google_bigquery_table_iam_member" "leech_owners" {
  for_each = toset(var.leech_table_iam.owners)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.leech_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "leech_editors" {
  for_each = toset(var.leech_table_iam.editors)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.leech_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "leech_viewers" {
  for_each = toset(var.leech_table_iam.viewers)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.leech_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

resource "google_storage_bucket" "leech_storage_bucket" {
  project = data.google_project.default.project_id

  name     = var.leech_bucket_name
  location = var.leech_bucket_location

  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
}


# Commit Review Status Table / IAM

resource "google_bigquery_table" "commit_review_status_table" {
  project = data.google_project.default.project_id

  deletion_protection = false
  table_id            = var.commit_review_status_table_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  schema = jsonencode([
    {
      name : "organization",
      type : "STRING",
      mode : "REQUIRED",
      description : "The GitHub organization to which the commit belongs."
    },
    {
      name : "repository",
      type : "STRING",
      mode : "REQUIRED",
      description : "The GitHub repository to which the commit belongs."
    },
    {
      name : "branch",
      type : "STRING",
      mode : "REQUIRED",
      description : "The GitHub branch to which the commit belongs."
    },
    {
      name : "commit_sha",
      type : "STRING",
      mode : "REQUIRED",
      description : "The SHA Hash for the commit."
    },
    {
      name : "commit_html_url",
      type : "STRING",
      mode : "REQUIRED",
      description : "The URL for the commit in GitHub"
    },
    {
      name : "pull_request_id",
      type : "INT64",
      mode : "NULLABLE",
      description : "The id of the pull request that introduced the commit."
    },
    {
      name : "approval_status",
      type : "STRING",
      mode : "REQUIRED",
      description : "The approval status of the commit in GitHub."
    },
    {
      name : "break_glass_issue_url",
      type : "STRING",
      mode : "NULLABLE",
      description : "The URL of the break glass issue that was used to introduce the commit."
    },
  ])
}

resource "google_bigquery_table_iam_member" "commit_review_status_owners" {
  for_each = toset(var.commit_review_status_iam.owners)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.commit_review_status_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "commit_review_status_editors" {
  for_each = toset(var.commit_review_status_iam.editors)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.commit_review_status_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "commit_review_status_viewers" {
  for_each = toset(var.commit_review_status_iam.viewers)

  project = data.google_project.default.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.commit_review_status_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}
