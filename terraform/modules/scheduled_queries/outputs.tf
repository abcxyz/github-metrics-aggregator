output "prstats_pull_requests_schedule" {
  description = "The created prstats_pull_requests_schedule."
  value       = google_bigquery_data_transfer_config.prstats_pull_requests_schedule
}

output "prstats_pull_request_reviews_schedule" {
  description = "The created prstats_pull_request_reviews_schedule."
  value       = google_bigquery_data_transfer_config.prstats_pull_request_reviews_schedule
}
