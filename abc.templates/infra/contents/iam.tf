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

resource "google_project_iam_member" "REPLACE_MODULE_NAME_actuators" {
  for_each = toset([
    "projects/${local.project_id}/roles/${local.cloudscheduler_job_creator}",
    "projects/${local.project_id}/roles/${local.cloudstorage_bucket_creator}",
    "projects/${local.project_id}/roles/${local.secretmanager_secret_creator}",
  ])

  project = local.project_id

  role   = each.value
  member = "REPLACE_TERRAFORM_ACTUATOR_MEMBER"

  depends_on = [
    google_project_iam_custom_role.cloudscheduler_job_creator,
    google_project_iam_custom_role.cloudstorage_bucket_creator,
    google_project_iam_custom_role.secretmanager_secret_creator,
  ]
}

locals { cloudscheduler_job_creator = "cloudschedulerJobCreator" }
resource "google_project_iam_custom_role" "cloudscheduler_job_creator" {
  project = local.project_id

  role_id     = local.cloudscheduler_job_creator
  title       = "Cloud Scheduler Job Creator"
  description = "Access to create Cloud Scheduler jobs"
  permissions = [
    "cloudscheduler.jobs.create",
  ]
}

locals { cloudstorage_bucket_creator = "cloudstorageBucketCreator" }
resource "google_project_iam_custom_role" "cloudstorage_bucket_creator" {
  project = local.project_id

  role_id     = local.cloudstorage_bucket_creator
  title       = "Cloud Storage Bucket Creator"
  description = "Access to create GCS buckets"
  permissions = [
    "storage.buckets.create",
  ]
}

locals { secretmanager_secret_creator = "secretmanagerSecretCreator" }
resource "google_project_iam_custom_role" "secretmanager_secret_creator" {
  project = local.project_id

  role_id     = local.secretmanager_secret_creator
  title       = "Secret Manager Secret Creator"
  description = "Access to create secrets in Secret Manager"
  permissions = [
    "secretmanager.secrets.create",
  ]
}
