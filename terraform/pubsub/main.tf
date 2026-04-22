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
  topic = google_pubsub_topic.relay.id

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
    dead_letter_topic     = google_pubsub_topic.dead_letter.id
    max_delivery_attempts = 5
  }
}

resource "google_pubsub_topic" "dead_letter" {
  project = var.project_id

  name = var.dead_letter_topic_id
}

resource "google_pubsub_subscription" "dead_letter" {
  project = var.project_id

  name = "${var.dead_letter_topic_id}-sub"

  topic = google_pubsub_topic.dead_letter.id

  expiration_policy {
    ttl = ""
  }
}

resource "google_pubsub_topic_iam_member" "dead_letter_publisher" {
  project = var.project_id

  topic = google_pubsub_topic.dead_letter.name

  role   = "roles/pubsub.publisher"
  member = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

data "google_project" "default" {
  project_id = var.project_id
}

resource "google_pubsub_topic" "relay" {

  project = var.relay_project_id

  name = var.relay_topic_id

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

resource "google_pubsub_subscription_iam_member" "relay_sub_dead_letter_ack" {
  project = var.project_id

  subscription = google_pubsub_subscription.relay_optimized_events.name
  role         = "roles/pubsub.subscriber"
  member       = "serviceAccount:service-${data.google_project.default.number}@gcp-sa-pubsub.iam.gserviceaccount.com"
}

# Resources moved from terraform/gma/pubsub.tf

resource "google_pubsub_schema" "enriched" {
  project = var.project_id

  name       = "${var.prefix_name}-enriched"
  type       = "PROTOCOL_BUFFER"
  definition = <<EOT
syntax = "proto3";

package github.metrics.aggregator;

message EnrichedEvent {
  string delivery_id = 1;
  string signature = 2;
  string received = 3;
  string event = 4;
  string payload = 5;
  string enterprise_id = 6;
  string enterprise_name = 7;
  string organization_id = 8;
  string organization_name = 9;
  string repository_id = 10;
  string repository_name = 11;
}
EOT
}

resource "google_pubsub_topic" "default" {
  project = var.project_id

  name = var.prefix_name
}

resource "google_pubsub_topic_iam_member" "topic_admins" {
  for_each = toset(var.events_topic_iam.admins)

  project = google_pubsub_topic.default.project

  topic = google_pubsub_topic.default.name

  role   = "roles/pubsub.admin"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "topic_editors" {
  for_each = toset(var.events_topic_iam.editors)

  project = google_pubsub_topic.default.project

  topic = google_pubsub_topic.default.name

  role   = "roles/pubsub.editor"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "topic_viewers" {
  for_each = toset(var.events_topic_iam.viewers)

  project = google_pubsub_topic.default.project

  topic = google_pubsub_topic.default.name

  role   = "roles/pubsub.viewer"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "topic_subscribers" {
  for_each = toset(var.events_topic_iam.subscribers)

  project = google_pubsub_topic.default.project

  topic = google_pubsub_topic.default.name

  role   = "roles/pubsub.subscriber"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "topic_publishers" {
  for_each = toset(var.events_topic_iam.publishers)

  project = google_pubsub_topic.default.project

  topic = google_pubsub_topic.default.name

  role   = "roles/pubsub.publisher"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "topic_publisher_webhook" {
  count = var.webhook_service_account_member != "" ? 1 : 0

  project = google_pubsub_topic.default.project

  topic = google_pubsub_topic.default.name

  role   = "roles/pubsub.publisher"
  member = var.webhook_service_account_member
}

# IAM for dead_letter topic (missing ones)
resource "google_pubsub_topic_iam_member" "dead_letter_admins" {
  for_each = toset(var.dlq_topic_iam.admins)

  project = google_pubsub_topic.dead_letter.project

  topic = google_pubsub_topic.dead_letter.name

  role   = "roles/pubsub.admin"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "dead_letter_editors" {
  for_each = toset(var.dlq_topic_iam.editors)

  project = google_pubsub_topic.dead_letter.project

  topic = google_pubsub_topic.dead_letter.name

  role   = "roles/pubsub.editor"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "dead_letter_viewers" {
  for_each = toset(var.dlq_topic_iam.viewers)

  project = google_pubsub_topic.dead_letter.project

  topic = google_pubsub_topic.dead_letter.name

  role   = "roles/pubsub.viewer"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "dead_letter_subscribers" {
  for_each = toset(var.dlq_topic_iam.subscribers)

  project = google_pubsub_topic.dead_letter.project

  topic = google_pubsub_topic.dead_letter.name

  role   = "roles/pubsub.subscriber"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "dead_letter_publishers" {
  for_each = toset(var.dlq_topic_iam.publishers)

  project = google_pubsub_topic.dead_letter.project

  topic = google_pubsub_topic.dead_letter.name

  role   = "roles/pubsub.publisher"
  member = each.value
}

resource "google_pubsub_topic_iam_member" "dead_letter_publisher_webhook" {
  count = var.webhook_service_account_member != "" ? 1 : 0

  project = google_pubsub_topic.dead_letter.project

  topic = google_pubsub_topic.dead_letter.name

  role   = "roles/pubsub.publisher"
  member = var.webhook_service_account_member
}

# IAM for dead_letter subscription
resource "google_pubsub_subscription_iam_member" "dead_letter_sub_admins" {
  for_each = toset(var.dead_letter_sub_iam.admins)

  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name

  role   = "roles/pubsub.admin"
  member = each.value
}

resource "google_pubsub_subscription_iam_member" "dead_letter_sub_editors" {
  for_each = toset(var.dead_letter_sub_iam.editors)

  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name

  role   = "roles/pubsub.editor"
  member = each.value
}

resource "google_pubsub_subscription_iam_member" "dead_letter_sub_viewers" {
  for_each = toset(var.dead_letter_sub_iam.viewers)

  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name

  role   = "roles/pubsub.viewer"
  member = each.value
}

resource "google_pubsub_subscription_iam_member" "dead_letter_sub_subscribers" {
  for_each = toset(var.dead_letter_sub_iam.subscribers)

  project = google_pubsub_subscription.dead_letter.project

  subscription = google_pubsub_subscription.dead_letter.name

  role   = "roles/pubsub.subscriber"
  member = each.value
}

