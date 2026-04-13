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
  artifact_window = 45 * local.minute + 5 * local.minute

  default_p_value = 99
}


resource "google_cloud_run_v2_job" "artifacts" {
  count = var.artifacts.enabled ? 1 : 0

  project = var.project_id

  name     = var.artifacts.job_name
  location = var.region

  template {
    parallelism = 0
    task_count  = 1

    template {
      containers {
        image = var.image

        args = ["job", "artifact"]

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
          value = var.events_table_id
        }
        env {
          name  = "ARTIFACTS_TABLE_ID"
          value = var.artifacts.table_id
        }
        env {
          name  = "BUCKET_NAME"
          value = google_storage_bucket.artifacts_storage_bucket[0].name
        }
        dynamic "env" {
          for_each = var.artifacts.job_additional_env_vars

          content {
            name  = env.key
            value = env.value
          }
        }
      }
      service_account = local.compute_service_account_email
    }
  }

  depends_on = [
    google_project_iam_member.artifacts_secret_accessor,
  ]
  lifecycle {
    ignore_changes = [
      client,
      client_version,
      template[0].template[0].containers[0].image,
    ]
  }
}

resource "google_cloud_run_v2_job_iam_binding" "artifacts_job_admins" {
  count = var.artifacts.enabled ? 1 : 0

  project = google_cloud_run_v2_job.artifacts[0].project

  location = google_cloud_run_v2_job.artifacts[0].location


  name = google_cloud_run_v2_job.artifacts[0].name

  role    = "roles/run.admin"
  members = toset(var.artifacts.job_iam.admins)
}

resource "google_cloud_run_v2_job_iam_binding" "artifacts_job_developers" {
  count = var.artifacts.enabled ? 1 : 0

  project = google_cloud_run_v2_job.artifacts[0].project

  location = google_cloud_run_v2_job.artifacts[0].location


  name = google_cloud_run_v2_job.artifacts[0].name

  role    = "roles/run.developer"
  members = toset(var.artifacts.job_iam.developers)
}

resource "google_cloud_run_v2_job_iam_binding" "artifacts_job_invokers" {
  count = var.artifacts.enabled ? 1 : 0

  project = google_cloud_run_v2_job.artifacts[0].project

  location = google_cloud_run_v2_job.artifacts[0].location


  name = google_cloud_run_v2_job.artifacts[0].name

  role    = "roles/run.invoker"
  members = toset(var.artifacts.job_iam.invokers)
}

resource "google_project_iam_member" "artifacts_secret_accessor" {
  count = var.artifacts.enabled ? 1 : 0

  project = var.project_id

  member = local.compute_service_account_member
  role   = "roles/secretmanager.secretAccessor"
}

resource "google_project_iam_member" "artifacts_invoker" {
  count = var.artifacts.enabled ? 1 : 0

  project = var.project_id

  member = local.compute_service_account_member
  role   = "roles/run.invoker"
}

resource "google_project_iam_member" "artifacts_bigquery_job_user" {
  count = var.artifacts.enabled ? 1 : 0

  project = var.project_id

  member = local.compute_service_account_member
  role   = "roles/bigquery.jobUser"
}

resource "google_bigquery_dataset_iam_member" "artifacts_dataset_viewer" {
  count = var.artifacts.enabled ? 1 : 0

  project = var.project_id

  dataset_id = var.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = local.compute_service_account_member
}

resource "google_bigquery_table_iam_member" "artifacts_table_editor" {
  count = var.artifacts.enabled ? 1 : 0

  project = var.project_id

  dataset_id = var.dataset_id
  table_id   = var.artifacts.table_id
  role       = "roles/bigquery.dataEditor"
  member     = local.compute_service_account_member
}

resource "google_project_iam_member" "artifacts_storage_object_user" {
  count = var.artifacts.enabled ? 1 : 0

  project = var.project_id

  member = local.compute_service_account_member
  role   = "roles/storage.objectUser"
}

resource "google_cloud_scheduler_job" "artifacts_scheduler" {
  count = var.artifacts.enabled ? 1 : 0

  project = var.project_id

  name             = var.artifacts.job_name
  region           = var.region
  schedule         = var.artifacts.scheduler_cron
  time_zone        = "Etc/UTC"
  attempt_deadline = "180s"
  retry_config {
    retry_count = "0"
  }


  http_target {
    http_method = "POST"
    uri         = "https://${google_cloud_run_v2_job.artifacts[0].location}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.artifacts[0].name}:run"
    oauth_token {
      service_account_email = local.compute_service_account_email
    }
  }
}



resource "google_storage_bucket" "artifacts_storage_bucket" {

  count = var.artifacts.enabled ? 1 : 0

  project = var.project_id

  name     = var.artifacts.bucket_name
  location = var.artifacts.bucket_location

  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
}
