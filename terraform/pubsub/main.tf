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

resource "google_pubsub_subscription" "relay_optimized_events" {

  project = var.project_id

  name  = "${var.prefix_name}-relay-optimized-events-sub"
  topic = "projects/${var.relay_project_id}/topics/${var.relay_topic_id}"

  bigquery_config {
    table                 = "${var.project_id}:${var.dataset_id}.${var.optimized_events_table_id}"
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

data "google_project" "default" {
  project_id = var.project_id
}

resource "google_pubsub_topic" "relay" {

  project = var.relay_project_id

  name = var.relay_topic_id != "" ? var.relay_topic_id : "${var.prefix_name}-relay"

  schema_settings {
    schema   = var.relay_schema_id
    encoding = "JSON"
  }
}

resource "google_pubsub_topic_iam_member" "relay_topic_remote_subscriber" {
  project = google_pubsub_topic.relay.project

  topic  = google_pubsub_topic.relay.name
  role   = "roles/pubsub.subscriber"
  member = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_pubsub_topic_iam_member" "relay_publisher" {
  project = google_pubsub_topic.relay.project

  topic  = google_pubsub_topic.relay.name
  role   = "roles/pubsub.publisher"
  member = var.relay_publisher_member
}
