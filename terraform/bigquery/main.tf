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

resource "google_project_iam_member" "job_users" {
  for_each = toset(var.job_users)

  project = var.project_id

  role   = "roles/bigquery.jobUser"
  member = each.value
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

resource "google_bigquery_dataset_iam_member" "metadata_viewers" {
  for_each = toset(var.dataset_metadata_viewers)

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = "roles/bigquery.metadataViewer"
  member     = each.value
}
