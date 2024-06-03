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

data "google_project" "default" {
  project_id = var.project_id
}

resource "google_project_service" "default" {
  for_each = toset([
    "bigquery.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "cloudscheduler.googleapis.com",
    "dataflow.googleapis.com",
    "datapipelines.googleapis.com",
    "pubsub.googleapis.com",
    "storage.googleapis.com",
  ])

  project = var.project_id

  service            = each.value
  disable_on_destroy = false
}

module "artifacts" {
  count = var.leech.enabled ? 1 : 0

  source = "./modules/artifacts"

  project_id = var.project_id

  dataset_id                        = google_bigquery_dataset.default.dataset_id
  leech_bucket_name                 = var.leech.bucket_name
  leech_bucket_location             = var.leech.bucket_location
  leech_table_id                    = var.leech.table_id
  leech_table_iam                   = var.leech.table_iam
  events_table_id                   = var.events_table_id
  github_app_id                     = var.github_app_id
  github_install_id                 = var.github_install_id
  github_private_key_secret_id      = var.github_private_key_secret_id
  github_private_key_secret_version = var.github_private_key_secret_version
  job_name                          = var.artifacts_job_name
}
