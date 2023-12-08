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

resource "google_bigquery_table" "invocation_comment_table" {
  project = var.project_id

  deletion_protection = false
  table_id            = var.invocation_comment_table_id
  dataset_id          = var.dataset_id
  schema = jsonencode([
    {
      "name" : "pull_request_id",
      "type" : "INT64",
      "mode" : "REQUIRED",
      "description" : "ID of pull request."
    },
    {
      "name" : "pull_request_html_url",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "URL of pull request."
    },
    {
      "name" : "processed_at",
      "type" : "TIMESTAMP",
      "mode" : "REQUIRED",
      "description" : "Timestamp of when the analyzer pipeline processed the PR."
    },
    {
      "name" : "comment_id",
      "type" : "INT64",
      "mode" : "NULLABLE",
      "description" : "ID of pull request comment."
    },
    {
      "name" : "status",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "The status of invocation comment operation."
    },
    {
      "name" : "job_name",
      "type" : "STRING",
      "mode" : "REQUIRED",
      "description" : "Job name of the analyzer that processed this event."
    },
  ])
}

resource "google_bigquery_table_iam_member" "invocation_comment_owners" {
  for_each = toset(var.invocation_comment_table_iam.owners)

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = google_bigquery_table.invocation_comment_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "invocation_comment_editors" {
  for_each = toset(var.invocation_comment_table_iam.editors)

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = google_bigquery_table.invocation_comment_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "invocation_comment_viewers" {
  for_each = toset(var.invocation_comment_table_iam.viewers)

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = google_bigquery_table.invocation_comment_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}
