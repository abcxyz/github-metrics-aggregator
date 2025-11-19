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


# Secret Manager secrets for the Cloud Run Retry service to use
resource "google_secret_manager_secret" "secrets" {
  for_each = toset(var.secrets_to_create)

  project = var.project_id

  secret_id = each.value
  replication {
    auto {}
  }

  depends_on = [
    google_project_service.services["secretmanager.googleapis.com"]
  ]
}

resource "google_secret_manager_secret_version" "secrets_default_version" {
  for_each = toset(var.secrets_to_create)

  secret = google_secret_manager_secret.secrets[each.key].id
  # default value used for initial revision to allow cloud run to map the secret
  # to manage this value and versions, use the google cloud web application
  secret_data = "DEFAULT_VALUE"

  lifecycle {
    ignore_changes = [
      enabled
    ]
  }
}

module "retry_job" {
  source = "./modules/retry"

  project_id = data.google_project.default.project_id

  job_name                          = "${var.prefix_name}-retry"
  region                            = var.region
  image                             = var.image
  dataset_id                        = var.dataset_id
  events_table_id                   = var.events_table_id
  checkpoint_table_id               = var.checkpoint_table_id
  bucket_name                       = google_storage_bucket.retry_lock.name
  github_app_id                     = var.github_app_id
  github_private_key_secret_id      = "github-private-key"
  github_private_key_secret_version = "latest"
  scheduler_cron                    = var.retry_job_schedule
  timeout                           = var.retry_job_timeout
  retry_job_iam = {
    admins     = toset(var.retry_service_iam.admins)
    developers = toset(concat(var.retry_service_iam.developers, [var.automation_service_account_member]))
    invokers   = toset(var.retry_service_iam.invokers)
  }

  # Make sure the secrets/versions are created before this module is provisioned
  depends_on = [
    google_secret_manager_secret_version.secrets_default_version
  ]
}

resource "google_storage_bucket" "retry_lock" {
  project = data.google_project.default.project_id

  name                        = "retry-lock-${random_id.default.hex}"
  location                    = "US"
  public_access_prevention    = "enforced"
  uniform_bucket_level_access = true

  depends_on = [
    google_project_service.default["storage.googleapis.com"]
  ]
}

resource "random_id" "default" {
  byte_length = 2
}

