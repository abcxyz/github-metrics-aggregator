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
