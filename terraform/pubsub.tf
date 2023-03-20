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

  name = "${var.prefix_name}-dead-letter"
  depends_on = [
    google_project_service.default["pubsub.googleapis.com"],
  ]
}

resource "google_pubsub_topic_iam_binding" "dead_letter_publishers" {
  project = google_pubsub_topic.dead_letter.project

  topic   = google_pubsub_topic.dead_letter.name
  role    = "roles/pubsub.publisher"
  members = ["serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"]
}

resource "google_pubsub_subscription" "dead_letter" {
  project = var.project_id
  name    = "${var.prefix_name}-dead-letter-sub"
  topic   = google_pubsub_topic.dead_letter.name
}

resource "google_pubsub_subscription_iam_binding" "dead_letter_sub_admins" {
  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name
  role         = "roles/pubsub.admin"
  members      = toset(var.dead_letter_sub_iam.admins)
}

resource "google_pubsub_subscription_iam_binding" "dead_letter_sub_editors" {
  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name
  role         = "roles/pubsub.editor"
  members      = toset(var.dead_letter_sub_iam.editors)
}

resource "google_pubsub_subscription_iam_binding" "dead_letter_sub_viewers" {
  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name
  role         = "roles/pubsub.viewer"
  members      = toset(var.dead_letter_sub_iam.viewers)
}

resource "google_pubsub_subscription_iam_binding" "dead_letter_sub_subscribers" {
  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name
  role         = "roles/pubsub.subscriber"
  members      = toset(var.dead_letter_sub_iam.subscribers)
}

resource "google_pubsub_schema" "default" {
  project = var.project_id

  name       = var.prefix_name
  type       = "PROTOCOL_BUFFER"
  definition = file("${path.module}/../protos/pubsub_schemas/event.proto")
  depends_on = [
    google_project_service.default["pubsub.googleapis.com"],
  ]
}

resource "google_pubsub_topic" "default" {
  project = var.project_id

  name = var.prefix_name
  schema_settings {
    schema   = google_pubsub_schema.default.id
    encoding = "JSON"
  }

  depends_on = [google_pubsub_schema.default]
}


resource "google_pubsub_topic_iam_binding" "topic_admins" {
  project = google_pubsub_topic.default.project

  topic   = google_pubsub_topic.default.name
  role    = "roles/pubsub.admin"
  members = toset(var.topic_iam.admins)
}

resource "google_pubsub_topic_iam_binding" "topic_editors" {
  project = google_pubsub_topic.default.project

  topic   = google_pubsub_topic.default.name
  role    = "roles/pubsub.editor"
  members = toset(var.topic_iam.editors)
}

resource "google_pubsub_topic_iam_binding" "topic_viewers" {
  project = google_pubsub_topic.default.project

  topic   = google_pubsub_topic.default.name
  role    = "roles/pubsub.viewer"
  members = toset(var.topic_iam.viewers)
}

resource "google_pubsub_topic_iam_binding" "topic_publishers" {
  project = google_pubsub_topic.default.project

  topic = google_pubsub_topic.default.name
  role  = "roles/pubsub.publisher"
  members = toset(
    concat(
      [google_service_account.webhook_run_service_account.member],
      var.topic_iam.publishers
    )
  )
}

resource "google_pubsub_topic_iam_binding" "topic_subscribers" {
  project = google_pubsub_topic.default.project

  topic   = google_pubsub_topic.default.name
  role    = "roles/pubsub.subscriber"
  members = toset(var.topic_iam.subscribers)
}

resource "google_pubsub_subscription" "default" {
  project = var.project_id

  name  = "${var.prefix_name}-bq-sub"
  topic = google_pubsub_topic.default.name

  bigquery_config {
    table            = format("${google_bigquery_table.events_table.project}:${google_bigquery_table.events_table.dataset_id}.${google_bigquery_table.events_table.table_id}")
    use_topic_schema = true
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.dead_letter.id
    max_delivery_attempts = 5
  }
}

resource "google_pubsub_subscription_iam_member" "editor" {
  project = google_pubsub_topic.default.project

  subscription = google_pubsub_subscription.default.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}
