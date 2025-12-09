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

locals {
  # time helpers
  second = 1
  minute = 60 * local.second
  hour   = 60 * local.minute
  day    = 24 * local.hour
}

resource "google_cloud_run_v2_job" "default" {
  project = var.project_id

  name     = var.job_name
  location = var.region

  template {
    parallelism = 0
    task_count  = 1

    template {
      timeout = var.timeout
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
              secret  = var.github_private_key_secret_id
              version = "latest"
            }
          }
        }
        env {
          name  = "GITHUB_PRIVATE_KEY_SECRET_ID"
          value = var.github_private_key_secret_id
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
          value = var.bucket_name
        }
        dynamic "env" {
          for_each = var.additional_env_vars

          content {
            name  = env.key
            value = env.value
          }
        }
      }
      service_account = google_service_account.default.email
    }
  }

  depends_on = [
    google_project_iam_member.secret_accessor_role,
    google_service_account.default,
  ]
  lifecycle {
    ignore_changes = [
      client,
      client_version,
      template[0].template[0].containers[0].image,
    ]
  }
}

resource "google_cloud_run_v2_job_iam_binding" "retry_job_admins" {
  project = google_cloud_run_v2_job.default.project

  location = google_cloud_run_v2_job.default.location
  name     = google_cloud_run_v2_job.default.name
  role     = "roles/run.admin"
  members  = toset(var.retry_job_iam.admins)
}

resource "google_cloud_run_v2_job_iam_binding" "retry_job_developers" {
  project = google_cloud_run_v2_job.default.project

  location = google_cloud_run_v2_job.default.location
  name     = google_cloud_run_v2_job.default.name
  role     = "roles/run.developer"
  members  = toset(var.retry_job_iam.developers)
}

resource "google_cloud_run_v2_job_iam_binding" "retry_job_invokers" {
  project = google_cloud_run_v2_job.default.project

  location = google_cloud_run_v2_job.default.location
  name     = google_cloud_run_v2_job.default.name
  role     = "roles/run.invoker"
  members  = toset(var.retry_job_iam.invokers)
}

resource "google_service_account" "default" {
  project = var.project_id

  account_id = "${var.job_name}-sa"
}

resource "google_project_iam_member" "secret_accessor_role" {
  project = var.project_id

  member = google_service_account.default.member
  role   = "roles/secretmanager.secretAccessor"
}

// Give the service account invoker permission
resource "google_project_iam_member" "invoker_role" {
  project = var.project_id

  member = google_service_account.default.member
  role   = "roles/run.invoker"
}

// give the service account permission to run bigquery jobs
resource "google_project_iam_member" "bigquery_job_user_role" {
  count   = var.bigquery_infra_deploy ? 1 : 0

  project = var.project_id

  member = google_service_account.default.member
  role   = "roles/bigquery.jobUser"
}

// give the service account read access to bigquery data set
resource "google_bigquery_dataset_iam_member" "dataset_viewer_role" {
  count   = var.bigquery_infra_deploy ? 1 : 0

  project = var.project_id

  dataset_id = var.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = google_service_account.default.member
}

// give the service account read and write access to the checkpoint table
resource "google_bigquery_table_iam_member" "checkpoint_table_editor_role" {
  count   = var.bigquery_infra_deploy ? 1 : 0
  
  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = var.checkpoint_table_id
  role       = "roles/bigquery.dataEditor"
  member     = google_service_account.default.member
}

// give the service account permission to write storage objects
resource "google_project_iam_member" "storage_object_user_role" {
  project = var.project_id

  member = google_service_account.default.member
  role   = "roles/storage.objectUser"
}

resource "google_cloud_scheduler_job" "scheduler" {
  project = var.project_id

  name             = var.job_name
  region           = var.region
  schedule         = var.scheduler_cron
  time_zone        = var.scheduler_timezone
  attempt_deadline = var.scheduler_deadline_duration
  retry_config {
    retry_count = var.scheduler_retry_limit
  }

  http_target {
    http_method = "POST"
    uri         = "https://${google_cloud_run_v2_job.default.location}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.default.name}:run"
    oauth_token {
      service_account_email = google_service_account.default.email
    }
  }
}
