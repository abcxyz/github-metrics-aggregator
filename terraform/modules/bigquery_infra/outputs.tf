output "dataset_id" {
  description = "The ID of the BigQuery dataset."
  value       = google_bigquery_dataset.default.dataset_id
}

output "events_table_id" {
  description = "The ID of the BigQuery table for events."
  value       = google_bigquery_table.events_table.table_id
}

output "raw_events_table_id" {
  description = "The ID of the BigQuery table for raw events."
  value       = google_bigquery_table.raw_events_table.table_id
}

output "optimized_events_table_id" {
  description = "The ID of the BigQuery table for optimized events."
  value       = google_bigquery_table.optimized_events_table.table_id
}

output "checkpoint_table_id" {
  description = "The ID of the BigQuery table for checkpoints."
  value       = google_bigquery_table.checkpoint_table.table_id
}

output "failure_events_table_id" {
  description = "The ID of the BigQuery table for failure events."
  value       = google_bigquery_table.failure_events_table.table_id
}

output "bigquery_event_views" {
  description = "BigQuery event view resources."
  value       = module.metrics_views.bigquery_event_views
}

output "bigquery_resource_views" {
  description = "BigQuery resource view resources."
  value       = module.metrics_views.bigquery_resource_views
}