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

terraform {
  required_version = ">= 1.3.0"
}

data "google_project" "default" {
  project_id = var.project_id
}

# Enable all services to make queries against BQ tables
# https://cloud.google.com/bigquery/docs/access-control#bigquery.jobUser

resource "google_project_iam_member" "webhook_job_user" {
  project = var.project_id

  role   = "roles/bigquery.jobUser"
  member = var.webhook_run_service_account_member
}

resource "google_project_iam_member" "retry_job_user" {
  project = var.project_id

  role   = "roles/bigquery.jobUser"
  member = var.retry_run_service_account_member
}

# Dataset / IAM
resource "google_bigquery_dataset" "default" {
  project = var.project_id

  dataset_id = var.dataset_id
  location   = var.dataset_location
}

resource "google_bigquery_dataset_iam_member" "default_owners" {
  for_each = toset(var.dataset_iam.owners)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_dataset_iam_member" "default_editors" {
  for_each = toset(var.dataset_iam.editors)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_dataset_iam_member" "default_viewers" {
  for_each = toset(var.dataset_iam.viewers)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

resource "google_bigquery_dataset_iam_member" "retry_dataset_metadata_viewer" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.metadataViewer"
  member     = var.retry_run_service_account_member
}

resource "google_bigquery_dataset_iam_member" "webhook_dataset_metadata_viewer" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.metadataViewer"
  member     = var.webhook_run_service_account_member
}

# Event Table / IAM

resource "google_bigquery_table" "events_table" {
  project = var.project_id

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

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "event_editors" {
  for_each = toset(var.events_table_iam.editors)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "event_pubsub_agent_editor" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_bigquery_table_iam_member" "event_webhook_editor" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = var.webhook_run_service_account_member
}

resource "google_bigquery_table_iam_member" "event_viewers" {
  for_each = toset(var.events_table_iam.viewers)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.events_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

# Optimized Events table with clustering and partitioning / IAM
resource "google_bigquery_table" "optimized_events_table" {
  project = var.project_id

  deletion_protection = true
  table_id            = var.optimized_events_table_id
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
      "name" : "enterprise_name",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Enterprise name from payload"
    },
    {
      "name" : "enterprise_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Enterprise ID from payload"
    },
    {
      "name" : "organization_name",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Organization login from payload"
    },
    {
      "name" : "organization_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Organization ID from payload"
    },
    {
      "name" : "repository_name",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Repository name from payload"
    },
    {
      "name" : "repository_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Repository ID from payload"
    },
    {
      "name" : "payload",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Event payload JSON string"
    }
  ])

  time_partitioning {
    field = "received"
    type  = var.bigquery_events_partition_granularity
  }

  clustering = ["event", "enterprise_name", "organization_name", "repository_name"]
}

resource "google_bigquery_table_iam_member" "optimized_event_owners" {
  for_each = toset(var.events_table_iam.owners)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.optimized_events_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "optimized_event_editors" {
  for_each = toset(var.events_table_iam.editors)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.optimized_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "optimized_event_viewers" {
  for_each = toset(var.events_table_iam.viewers)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.optimized_events_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "optimized_event_pubsub_agent_editor" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.optimized_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

# Events table with clustering and partitioning / IAM
resource "google_bigquery_table" "raw_events_table" {
  project = var.project_id

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

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "raw_event_editors" {
  for_each = toset(var.events_table_iam.editors)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "raw_event_pubsub_agent_editor" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_bigquery_table_iam_member" "raw_event_webhook_editor" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = var.webhook_run_service_account_member
}

resource "google_bigquery_table_iam_member" "raw_event_viewers" {
  for_each = toset(var.events_table_iam.viewers)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.raw_events_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

# Checkpoint Table / IAM

resource "google_bigquery_table" "checkpoint_table" {
  project = var.project_id

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
    {
      "name" : "github_instance_url",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The github instance the retry service is running for."
    },
  ])
}

resource "google_bigquery_table_iam_member" "checkpoint_owners" {
  for_each = toset(var.checkpoint_table_iam.owners)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.checkpoint_table.table_id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "checkpoint_editors" {
  for_each = toset(var.checkpoint_table_iam.editors)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.checkpoint_table.table_id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "checkpoint_retry_editor" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.checkpoint_table.table_id
  role       = "roles/bigquery.dataEditor"
  member     = var.retry_run_service_account_member
}

resource "google_bigquery_table_iam_member" "checkpoint_viewers" {
  for_each = toset(var.checkpoint_table_iam.viewers)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.checkpoint_table.table_id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

# Failure Events Table / IAM

resource "google_bigquery_table" "failure_events_table" {
  project = var.project_id

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

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.failure_events_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "failure_events_editors" {
  for_each = toset(var.failure_events_table_iam.editors)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.failure_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "failure_events_webhook_editor" {
  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.failure_events_table.id
  role       = "roles/bigquery.dataEditor"
  member     = var.webhook_run_service_account_member
}

resource "google_bigquery_table_iam_member" "failure_events_viewers" {
  for_each = toset(var.failure_events_table_iam.viewers)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.failure_events_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

# Unique Events - deduplicate rows

resource "google_bigquery_table" "unique_events_view" {
  project = var.project_id

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
      JSON_VALUE(payload, "$.enterprise.name") enterprise,
      SAFE_CAST(JSON_VALUE(payload, "$.enterprise.id") AS INT64) enterprise_id,
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
  project = var.project_id

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
      LAX_STRING(payload.enterprise.name) enterprise,
      SAFE.INT64(payload.enterprise.id) enterprise_id,
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

  project_id = var.project_id

  dataset_id    = google_bigquery_dataset.default.dataset_id
  base_table_id = google_bigquery_table.unique_events_view.table_id
  base_tvf_id   = google_bigquery_routine.unique_events_by_date_type.routine_id
}

module "invocation_comment" {
  count = var.invocation_comment.enabled ? 1 : 0

  source = "./modules/invocation_comment"

  project_id = var.project_id

  dataset_id                   = google_bigquery_dataset.default.dataset_id
  invocation_comment_table_id  = var.invocation_comment.table_id
  invocation_comment_table_iam = var.invocation_comment.table_iam
}

# add all groups who need to view through lookerstudio to jobUser role
resource "google_project_iam_member" "github_metrics_dashboard_job_users" {
  for_each = var.github_metrics_dashboard.enabled ? toset(var.github_metrics_dashboard.viewers) : []

  project = var.project_id

  role   = "roles/bigquery.jobUser"
  member = each.value
}

# grant users access to the dataset used for displaying PR stats
resource "google_bigquery_dataset_iam_member" "github_metrics_dashboard_data_viewers" {
  for_each = var.github_metrics_dashboard.enabled ? toset(var.github_metrics_dashboard.viewers) : []

  project = var.project_id

  dataset_id = var.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "optimized_events_relay_sa_editor" {
  count = var.enable_relay_service ? 1 : 0

  project = var.project_id

  dataset_id = module.bigquery_dataset.dataset_id
  table_id   = google_bigquery_table.optimized_events_table.table_id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:${var.relay_sub_service_account_email}"
}
