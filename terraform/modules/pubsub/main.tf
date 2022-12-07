data "google_project" "project" {
  project_id = var.project_id
}

resource "google_pubsub_topic" "dead_letter" {
  name = "${var.name}-dead-letter"
}

resource "google_pubsub_topic_iam_member" "dead_letter_publisher" {
  project = google_pubsub_topic.dead_letter.project
  topic   = google_pubsub_topic.dead_letter.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_pubsub_subscription" "dead_letter" {
  name  = "${var.name}-dead-letter-sub"
  topic = google_pubsub_topic.dead_letter.name
}

resource "google_pubsub_schema" "default" {
  project    = var.project_id
  name       = var.name
  type       = "PROTOCOL_BUFFER"
  definition = file("${path.module}/pubsub_schema.proto")
}

resource "google_pubsub_topic" "default" {
  project = var.project_id
  name    = var.name
  schema_settings {
    schema   = google_pubsub_schema.default.id
    encoding = "JSON"
  }

  depends_on = [google_pubsub_schema.default]
}

resource "google_pubsub_topic_iam_binding" "bindings" {
  for_each = var.topic_iam
  project  = var.project_id
  topic    = google_pubsub_topic.default.name
  role     = each.key
  members  = each.value
}

resource "google_pubsub_subscription" "default" {
  project = var.project_id
  name    = "${var.name}-bq-sub"
  topic   = google_pubsub_topic.default.name

  bigquery_config {
    table            = var.bigquery_destination
    use_topic_schema = true
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.dead_letter.id
    max_delivery_attempts = 5
  }
}

resource "google_pubsub_subscription_iam_member" "editor" {
  project      = google_pubsub_topic.default.project
  subscription = google_pubsub_subscription.default.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}
