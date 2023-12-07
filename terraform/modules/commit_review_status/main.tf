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

resource "google_bigquery_table" "commit_review_status_table" {
  project = var.project_id

  deletion_protection = false
  table_id            = var.commit_review_status_table_id
  dataset_id          = var.dataset_id
  schema = jsonencode([
    {
      name : "author",
      type : "STRING",
      mode : "REQUIRED",
      description : "The author of the commit."
    },
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
      name : "commit_timestamp",
      type : "TIMESTAMP",
      mode : "REQUIRED",
      description : "The Timestamp when the commit was made"
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
      name : "pull_request_number",
      type : "INT64",
      mode : "NULLABLE",
      description : "The number of the pull request that introduced the commit."
    },
    {
      name : "pull_request_html_url",
      type : "STRING",
      mode : "NULLABLE",
      description : "The html url of the pull request that introduced the commit."
    },
    {
      name : "approval_status",
      type : "STRING",
      mode : "REQUIRED",
      description : "The approval status of the commit in GitHub."
    },
    {
      name : "break_glass_issue_urls",
      type : "STRING",
      mode : "REPEATED",
      description : "The URLs of the break glass issues that the author had open during the time the commit was made."
    },
    {
      name : "note",
      type : "STRING",
      mode : "NULLABLE",
      description : "Optional context on the about the commit (e.g. a processing error message)"
    },
  ])
}

resource "google_bigquery_table_iam_member" "commit_review_status_owners" {
  for_each = toset(var.commit_review_status_table_iam.owners)

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = google_bigquery_table.commit_review_status_table.id
  role       = "roles/bigquery.dataOwner"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "commit_review_status_editors" {
  for_each = toset(var.commit_review_status_table_iam.editors)

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = google_bigquery_table.commit_review_status_table.id
  role       = "roles/bigquery.dataEditor"
  member     = each.value
}

resource "google_bigquery_table_iam_member" "commit_review_status_viewers" {
  for_each = toset(var.commit_review_status_table_iam.viewers)

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = google_bigquery_table.commit_review_status_table.id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}
