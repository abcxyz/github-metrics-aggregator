resource "google_cloud_run_v2_job" "default" {
  project = var.project_id

  name     = var.job_name
  location = var.region

  template {
    parallelism = 0
    task_count  = 1

    template {
      containers {
        image = "us-docker.pkg.dev/cloudrun/container/hello"

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
          name  = "PUSH_EVENTS_TABLE_ID"
          value = var.push_events_table_id
        }
        env {
          name  = "ISSUES_TABLE_ID"
          value = var.issues_table_id
        }
        env {
          name  = "COMMIT_REVIEW_STATUS_TABLE_ID"
          value = google_bigquery_table.commit_review_status_table.id
        }
      }
    }
  }

  depends_on = [
    google_project_iam_member.default,
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

resource "google_project_iam_member" "default" {
  project = var.project_id

  member = google_service_account.default.member
  role   = "roles/secretmanager.secretAccessor"
}
