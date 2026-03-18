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
  commit_review_status_window = 8 * local.hour + 10 * local.minute
}

resource "google_cloud_run_v2_job" "commit_review_status" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = var.project_id

  name     = var.commit_review_status.job_name
  location = var.region

  template {
    parallelism = 0
    task_count  = 1

    template {
      containers {
        image = var.image

        args = ["job", "review"]

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
          value = var.optimized_events_table_id # Wait, let's verify if they used optimized_events_table_id!
        }
        env {
          name  = "COMMIT_REVIEW_STATUS_TABLE_ID"
          value = var.commit_review_status.table_id
        }
        dynamic "env" {
          for_each = var.commit_review_status.job_additional_env_vars

          content {
            name  = env.key
            value = env.value
          }
        }
      }
      service_account = google_service_account.commit_review_status_sa[0].email
    }
  }

  depends_on = [
    google_project_iam_member.commit_review_status_secret_accessor,
    google_service_account.commit_review_status_sa,
  ]
  lifecycle {
    ignore_changes = [
      client,
      client_version,
      template[0].template[0].containers[0].image,
    ]
  }
}

resource "google_cloud_run_v2_job_iam_binding" "commit_review_status_job_admins" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = google_cloud_run_v2_job.commit_review_status[0].project

  location = google_cloud_run_v2_job.commit_review_status[0].location

  name = google_cloud_run_v2_job.commit_review_status[0].name


  role    = "roles/run.admin"
  members = toset(var.commit_review_status.job_iam.admins)
}

resource "google_cloud_run_v2_job_iam_binding" "commit_review_status_job_developers" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = google_cloud_run_v2_job.commit_review_status[0].project

  location = google_cloud_run_v2_job.commit_review_status[0].location

  name = google_cloud_run_v2_job.commit_review_status[0].name


  role    = "roles/run.developer"
  members = toset(var.commit_review_status.job_iam.developers)
}

resource "google_cloud_run_v2_job_iam_binding" "commit_review_status_job_invokers" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = google_cloud_run_v2_job.commit_review_status[0].project

  location = google_cloud_run_v2_job.commit_review_status[0].location

  name = google_cloud_run_v2_job.commit_review_status[0].name


  role    = "roles/run.invoker"
  members = toset(var.commit_review_status.job_iam.invokers)
}

resource "google_service_account" "commit_review_status_sa" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = var.project_id

  account_id = "${var.commit_review_status.job_name}-sa"
}

resource "google_project_iam_member" "commit_review_status_secret_accessor" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = var.project_id

  member = google_service_account.commit_review_status_sa[0].member
  role   = "roles/secretmanager.secretAccessor"
}

resource "google_project_iam_member" "commit_review_status_invoker" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = var.project_id

  member = google_service_account.commit_review_status_sa[0].member
  role   = "roles/run.invoker"
}

resource "google_project_iam_member" "commit_review_status_bigquery_job_user" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = var.project_id

  member = google_service_account.commit_review_status_sa[0].member
  role   = "roles/bigquery.jobUser"
}

resource "google_bigquery_dataset_iam_member" "commit_review_status_dataset_viewer" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = var.project_id

  dataset_id = var.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = google_service_account.commit_review_status_sa[0].member
}

resource "google_bigquery_table_iam_member" "commit_review_status_table_editor" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = var.commit_review_status.table_id
  role       = "roles/bigquery.dataEditor"
  member     = google_service_account.commit_review_status_sa[0].member
}

resource "google_cloud_scheduler_job" "commit_review_status_scheduler" {
  count = var.commit_review_status.enabled ? 1 : 0

  project = var.project_id

  name             = var.commit_review_status.job_name
  region           = var.region
  schedule         = var.commit_review_status.scheduler_cron
  time_zone        = "Etc/UTC"
  attempt_deadline = "180s"
  retry_config {
    retry_count = "0"
  }


  http_target {
    http_method = "POST"
    uri         = "https://${google_cloud_run_v2_job.commit_review_status[0].location}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.commit_review_status[0].name}:run"
    oauth_token {
      service_account_email = google_service_account.commit_review_status_sa[0].email
    }
  }
}
