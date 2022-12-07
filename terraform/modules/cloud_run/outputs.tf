output "service" {
  description = "Cloud Run service."
  value       = google_cloud_run_service.default
}

output "service_account" {
  description = "Cloud Run service account."
  value       = google_service_account.default
}

output "service_account_email" {
  description = "Cloud Run service account email."
  value       = google_service_account.default.email
}

output "service_account_iam_email" {
  description = "Cloud Run service account email iam string."
  value       = format("serviceAccount:%s", google_service_account.default.email)
}
