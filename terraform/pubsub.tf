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

resource "google_pubsub_topic_iam_member" "dead_letter_admins" {
  for_each = toset(var.dlq_topic_iam.admins)

  project = google_pubsub_topic.dead_letter.project

  topic  = google_pubsub_topic.dead_letter.name
  role   = "roles/pubsub.admin"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "dead_letter_editors" {
  for_each = toset(var.dlq_topic_iam.editors)

  project = google_pubsub_topic.dead_letter.project

  topic  = google_pubsub_topic.dead_letter.name
  role   = "roles/pubsub.editor"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "dead_letter_viewers" {
  for_each = toset(var.dlq_topic_iam.viewers)

  project = google_pubsub_topic.dead_letter.project

  topic  = google_pubsub_topic.dead_letter.name
  role   = "roles/pubsub.viewer"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "dead_letter_subscribers" {
  for_each = toset(var.dlq_topic_iam.subscribers)

  project = google_pubsub_topic.dead_letter.project

  topic  = google_pubsub_topic.dead_letter.name
  role   = "roles/pubsub.subscriber"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "dead_letter_publishers" {
  for_each = toset(var.dlq_topic_iam.publishers)

  project = google_pubsub_topic.dead_letter.project

  topic  = google_pubsub_topic.dead_letter.name
  role   = "roles/pubsub.publisher"
  member = each.value
}

# Purposely not combined with the dead_letter_publishers resource as it will cause a TF error
resource "google_pubsub_topic_iam_member" "dead_letter_publisher_webhook" {
  project = google_pubsub_topic.dead_letter.project

  topic  = google_pubsub_topic.dead_letter.name
  role   = "roles/pubsub.publisher"
  member = google_service_account.webhook_run_service_account.member
}

# Allow the PubSub SA to publish the DLQ
resource "google_pubsub_topic_iam_member" "dead_letter_publisher_default" {
  project = google_pubsub_topic.dead_letter.project

  topic  = google_pubsub_topic.dead_letter.name
  role   = "roles/pubsub.publisher"
  member = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

resource "google_pubsub_subscription" "dead_letter" {
  project = var.project_id

  name  = "${var.prefix_name}-dead-letter-sub"
  topic = google_pubsub_topic.dead_letter.name

  # set to never expire
  expiration_policy {
    ttl = ""
  }
}

resource "google_pubsub_subscription_iam_member" "dead_letter_sub_admins" {
  for_each = toset(var.dead_letter_sub_iam.admins)

  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name
  role         = "roles/pubsub.admin"
  member       = each.value
}

resource "google_pubsub_subscription_iam_member" "dead_letter_sub_editors" {
  for_each = toset(var.dead_letter_sub_iam.editors)

  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name
  role         = "roles/pubsub.editor"
  member       = each.value
}

resource "google_pubsub_subscription_iam_member" "dead_letter_sub_viewers" {
  for_each = toset(var.dead_letter_sub_iam.viewers)

  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name
  role         = "roles/pubsub.viewer"
  member       = each.value
}

resource "google_pubsub_subscription_iam_member" "dead_letter_sub_subscribers" {
  for_each = toset(var.dead_letter_sub_iam.subscribers)

  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name
  role         = "roles/pubsub.subscriber"
  member       = each.value
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

resource "google_pubsub_topic_iam_member" "topic_admins" {
  for_each = toset(var.events_topic_iam.admins)

  project = google_pubsub_topic.default.project

  topic  = google_pubsub_topic.default.name
  role   = "roles/pubsub.admin"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "topic_editors" {
  for_each = toset(var.events_topic_iam.editors)

  project = google_pubsub_topic.default.project

  topic  = google_pubsub_topic.default.name
  role   = "roles/pubsub.editor"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "topic_viewers" {
  for_each = toset(var.events_topic_iam.viewers)

  project = google_pubsub_topic.default.project

  topic  = google_pubsub_topic.default.name
  role   = "roles/pubsub.viewer"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "topic_subscribers" {
  for_each = toset(var.events_topic_iam.subscribers)

  project = google_pubsub_topic.default.project

  topic  = google_pubsub_topic.default.name
  role   = "roles/pubsub.subscriber"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "topic_publishers" {
  for_each = toset(var.events_topic_iam.publishers)

  project = google_pubsub_topic.default.project

  topic  = google_pubsub_topic.default.name
  role   = "roles/pubsub.publisher"
  member = each.value
}

# Purposely not combined with the topic_publishers resource as it will cause a TF error
resource "google_pubsub_topic_iam_member" "topic_publisher_webhook" {
  project = google_pubsub_topic.default.project

  topic  = google_pubsub_topic.default.name
  role   = "roles/pubsub.publisher"
  member = google_service_account.webhook_run_service_account.member
}

resource "google_pubsub_subscription" "default" {
  project = var.project_id

  name  = "${var.prefix_name}-bq-sub"
  topic = google_pubsub_topic.default.name

  bigquery_config {
    table            = format("${google_bigquery_table.events_table.project}:${google_bigquery_table.events_table.dataset_id}.${google_bigquery_table.events_table.table_id}")
    use_topic_schema = true
  }

  # set to never expire
  expiration_policy {
    ttl = ""
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.dead_letter.id
    max_delivery_attempts = 5
  }
}

resource "google_pubsub_subscription" "json" {
  project = var.project_id

  name  = "${var.prefix_name}-bq-json-sub"
  topic = google_pubsub_topic.default.name

  bigquery_config {
    # TODO: fix once I get raw events merged in so its pointing to correct tabletable            = format("${google_bigquery_table.events_table.project}:${google_bigquery_table.events_table.dataset_id}.${google_bigquery_table.events_table.table_id}")
    use_topic_schema = true
  }

  # set to never expire
  expiration_policy {
    ttl = ""
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
