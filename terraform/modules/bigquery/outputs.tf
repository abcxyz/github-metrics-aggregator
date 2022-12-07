output "bigquery_dataset" {
  description = "BigQuery dataset resource."
  value       = google_bigquery_dataset.default
}

output "bigquery_table" {
  description = "BigQuery table resource."
  value       = google_bigquery_table.default
}

output "bigquery_destination" {
  description = "BigQuery destination"
  value       = format("${google_bigquery_table.default.project}:${google_bigquery_table.default.dataset_id}.${google_bigquery_table.default.table_id}")
}

output "bigquery_default_views" {
  description = "Map of BigQuery default view resources."
  value       = google_bigquery_table.default_views
}

output "bigquery_views" {
  description = "Map of BigQuery view resources."
  value       = google_bigquery_table.views
}

