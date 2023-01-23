# Copyright 2023 The Authors (see AUTHORS file)
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

resource "google_pubsub_topic" "dead_letter" {
  project = var.project_id
  name    = "${var.name}-dead-letter"
  depends_on = [
    google_project_service.default["pubsub.googleapis.com"],
  ]
}

resource "google_pubsub_topic_iam_member" "dead_letter_publisher" {
  project = google_pubsub_topic.dead_letter.project
  topic   = google_pubsub_topic.dead_letter.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_pubsub_subscription" "dead_letter" {
  project = var.project_id
  name    = "${var.name}-dead-letter-sub"
  topic   = google_pubsub_topic.dead_letter.name
}

resource "google_pubsub_schema" "default" {
  project    = var.project_id
  name       = var.name
  type       = "PROTOCOL_BUFFER"
  definition = file("${path.module}/../protos/pubsub_schemas/event.proto")
  depends_on = [
    google_project_service.default["pubsub.googleapis.com"],
  ]
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
    table            = format("${google_bigquery_table.default.project}:${google_bigquery_table.default.dataset_id}.${google_bigquery_table.default.table_id}")
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
  member       = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}
