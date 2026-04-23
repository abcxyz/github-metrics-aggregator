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

output "artifacts_job" {
  description = "The Cloud Run Job for artifact data. Only populated when var.artifacts.enabled is set."
  value = {
    job_id                 = try(google_cloud_run_v2_job.artifacts[0].id, null)
    job_name               = try(google_cloud_run_v2_job.artifacts[0].name, null)
    service_account_name   = var.artifacts.enabled ? google_service_account.backend.name : null
    service_account_email  = var.artifacts.enabled ? google_service_account.backend.email : null
    service_account_member = var.artifacts.enabled ? "serviceAccount:${google_service_account.backend.email}" : null
  }
}

output "commit_review_status_job" {
  description = "The Cloud Run Job for commit review status data. Only populated when var.commit_review_status.enabled is set."
  value = {
    job_id                 = try(google_cloud_run_v2_job.commit_review_status[0].id, null)
    job_name               = try(google_cloud_run_v2_job.commit_review_status[0].name, null)
    service_account_name   = var.commit_review_status.enabled ? google_service_account.backend.name : null
    service_account_email  = var.commit_review_status.enabled ? google_service_account.backend.email : null
    service_account_member = var.commit_review_status.enabled ? "serviceAccount:${google_service_account.backend.email}" : null
  }
}

output "retry_run_job" {
  description = "The Cloud Run Job for retry data."
  value = {
    job_id                 = google_cloud_run_v2_job.retry.id
    job_name               = google_cloud_run_v2_job.retry.name
    service_account_name   = google_service_account.backend.name
    service_account_email  = google_service_account.backend.email
    service_account_member = "serviceAccount:${google_service_account.backend.email}"
  }
}
