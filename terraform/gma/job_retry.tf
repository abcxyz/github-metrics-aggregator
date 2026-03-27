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
    google_project_service.default["secretmanager.googleapis.com"]
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

resource "google_cloud_run_v2_job" "retry" {
  project = var.project_id

  name     = "${var.prefix_name}-retry"
  location = var.region

  template {
    parallelism = 0
    task_count  = 1

    template {
      timeout = var.retry_job_timeout
      containers {
        image = var.image

        args = ["job", "retry"]

        env {
          name  = "GITHUB_APP_ID"
          value = var.github_app_id
        }
        env {
          name = "GITHUB_PRIVATE_KEY"
          value_source {
            secret_key_ref {
              secret  = "github-private-key"
              version = "latest"
            }
          }
        }
        env {
          name  = "GITHUB_ENTERPRISE_SERVER_URL"
          value = var.github_enterprise_server_url
        }
        env {
          name  = "PROJECT_ID"
          value = var.project_id
        }
        env {
          name  = "BIG_QUERY_PROJECT_ID"
          value = coalesce(var.bigquery_project_id, var.project_id)
        }
        env {
          name  = "DATASET_ID"
          value = var.dataset_id
        }
        env {
          name  = "EVENTS_TABLE_ID"
          value = var.events_table_id
        }
        env {
          name  = "CHECKPOINT_TABLE_ID"
          value = var.checkpoint_table_id
        }
        env {
          name  = "BUCKET_NAME"
          value = google_storage_bucket.retry_lock.name
        }
      }
      service_account = local.compute_service_account_email
    }
  }

  depends_on = [
    google_project_iam_member.retry_secret_accessor,
    google_secret_manager_secret_version.secrets_default_version,
  ]
  lifecycle {
    ignore_changes = [
      template[0].template[0].containers[0].image,
    ]
  }
}

resource "google_cloud_run_v2_job_iam_binding" "retry_job_admins" {
  project = google_cloud_run_v2_job.retry.project

  location = google_cloud_run_v2_job.retry.location


  name = google_cloud_run_v2_job.retry.name

  role    = "roles/run.admin"
  members = toset(var.retry_service_iam.admins)
}

resource "google_cloud_run_v2_job_iam_binding" "retry_job_developers" {
  project = google_cloud_run_v2_job.retry.project

  location = google_cloud_run_v2_job.retry.location


  name = google_cloud_run_v2_job.retry.name

  role    = "roles/run.developer"
  members = toset(concat(var.retry_service_iam.developers, [var.automation_service_account_member]))
}

resource "google_cloud_run_v2_job_iam_binding" "retry_job_invokers" {
  project = google_cloud_run_v2_job.retry.project

  location = google_cloud_run_v2_job.retry.location


  name = google_cloud_run_v2_job.retry.name

  role    = "roles/run.invoker"
  members = toset(var.retry_service_iam.invokers)
}

resource "google_project_iam_member" "retry_secret_accessor" {
  project = var.project_id

  member = local.compute_service_account_member

  role = "roles/secretmanager.secretAccessor"
}

resource "google_project_iam_member" "retry_invoker" {
  project = var.project_id

  member = local.compute_service_account_member

  role = "roles/run.invoker"
}

resource "google_project_iam_member" "retry_bigquery_job_user" {
  count = var.bigquery_infra_deploy ? 1 : 0

  project = var.project_id

  member = local.compute_service_account_member

  role = "roles/bigquery.jobUser"
}

resource "google_bigquery_dataset_iam_member" "retry_dataset_viewer" {
  count = var.bigquery_infra_deploy ? 1 : 0

  project = var.project_id

  dataset_id = var.dataset_id


  role = "roles/bigquery.dataViewer"

  member = local.compute_service_account_member
}

resource "google_bigquery_table_iam_member" "retry_checkpoint_table_editor" {
  count = var.bigquery_infra_deploy ? 1 : 0

  project = var.project_id

  dataset_id = var.dataset_id

  table_id = var.checkpoint_table_id

  role = "roles/bigquery.dataEditor"

  member = local.compute_service_account_member
}

resource "google_project_iam_member" "retry_storage_object_user" {
  project = var.project_id

  member = local.compute_service_account_member

  role = "roles/storage.objectUser"
}

resource "google_cloud_scheduler_job" "retry_scheduler" {
  project = var.project_id

  name             = "${var.prefix_name}-retry"
  region           = var.region
  schedule         = var.retry_job_schedule
  time_zone        = "Etc/UTC"
  attempt_deadline = "180s"
  retry_config {
    retry_count = "0"
  }

  http_target {
    http_method = "POST"
    uri         = "https://${google_cloud_run_v2_job.retry.location}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.retry.name}:run"
    oauth_token {
      service_account_email = local.compute_service_account_email
    }
  }
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
