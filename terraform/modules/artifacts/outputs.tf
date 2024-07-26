output "google_service_account" {
  value = google_service_account.default
}

output "job_id" {
  value = google_cloud_run_v2_job.default.id
}

output "job_name" {
  value = google_cloud_run_v2_job.default.name
}
