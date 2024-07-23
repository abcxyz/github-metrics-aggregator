resource "google_cloud_run_v2_job" "default" {
  project = var.project_id

  name     = var.job_name
  location = var.region

  template {
    parallelism = 0
    task_count  = 1

    template {
      containers {
        image = var.image

        env {
          name  = "GITHUB_APP_ID"
          value = var.github_app_id
        }
        env {
          name  = "GITHUB_INSTALL_ID"
          value = var.github_install_id
        }
        env {
          name = "GITHUB_PRIVATE_KEY_SECRET"
          value_source {
            secret_key_ref {
              secret  = var.github_private_key_secret_id
              version = var.github_private_key_secret_version
            }
          }
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
          value = google_bigquery_table.leech_table.table_id
        }
        env {
          name  = "BUCKET_NAME"
          value = google_storage_bucket.leech_storage_bucket.name
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
      template[0].template[0].containers[0].image,
    ]
  }
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
