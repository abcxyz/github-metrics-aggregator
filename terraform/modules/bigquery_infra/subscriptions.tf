
resource "google_pubsub_subscription" "relay_optimized_events" {
  count = var.enable_relay_service ? 1 : 0

  project = var.project_id

  name  = "${var.prefix_name}-relay-optimized-events-sub"
  topic = "projects/${var.relay_project_id}/topics/${var.relay_topic_id}"

  bigquery_config {
    table                 = "${var.project_id}:${google_bigquery_dataset.default.dataset_id}.${google_bigquery_table.optimized_events_table.table_id}"
    use_topic_schema      = true
    service_account_email = var.relay_sub_service_account_email
  }

  # set to never expire
  expiration_policy {
    ttl = ""
  }

  dead_letter_policy {
    dead_letter_topic     = var.dead_letter_topic_id
    max_delivery_attempts = 5
  }
}
