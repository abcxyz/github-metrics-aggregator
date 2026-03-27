# Copyright 2026 The Authors (see AUTHORS file)
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

resource "google_service_account" "compute_sa" {
  count = var.compute_service_account_email == "" ? 1 : 0

  project = data.google_project.default.project_id

  account_id   = "${var.prefix_name}-compute"
  display_name = "${var.prefix_name} Compute Service Account"
}

locals {
  compute_service_account_email  = var.compute_service_account_email == "" ? google_service_account.compute_sa[0].email : var.compute_service_account_email
  compute_service_account_name   = var.compute_service_account_email == "" ? google_service_account.compute_sa[0].name : "projects/${data.google_project.default.project_id}/serviceAccounts/${var.compute_service_account_email}"
  compute_service_account_member = "serviceAccount:${local.compute_service_account_email}"
}

# Allow the automation service account to act as the compute service account.
# This enables the automation service account to deploy new revisions for the compute workloads.
resource "google_service_account_iam_member" "compute_sa_user" {
  count = var.compute_service_account_email == "" ? 1 : 0

  service_account_id = google_service_account.compute_sa[0].name
  role               = "roles/iam.serviceAccountUser"
  member             = var.automation_service_account_member
}
