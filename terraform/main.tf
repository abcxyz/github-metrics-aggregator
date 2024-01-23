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

module "dataflow_vpc" {
  # This may be shared across multiple features and not exclusive to code audit.
  # If another feature takes a dependency on dataflow, then it should be added
  # to the conditional below.
  count = var.code_audit_dashboard.enabled ? 1 : 0

  project_id = var.project_id
}
