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

resource "google_pubsub_topic_iam_member" "relay_webhook_topic_subscribers" {
  count = var.enable_relay_service ? 1 : 0

  project = google_pubsub_topic.default.project

  topic  = google_pubsub_topic.default.name
  role   = "roles/pubsub.subscriber"
  member = local.compute_service_account_member
}

module "relay_cloud_run" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=1467eaf0115f71613727212b0b51b3f99e699842"

  count = var.enable_relay_service ? 1 : 0

  project_id = data.google_project.default.project_id

  name                  = "${var.prefix_name}-relay"
  region                = var.region
  image                 = var.image
  args                  = ["relay"]
  ingress               = "internal-and-cloud-load-balancing"
  service_account_email = local.compute_service_account_email
  service_iam = {
    admins     = toset(var.relay_service_iam.admins)
    developers = toset(concat(var.relay_service_iam.developers, [var.automation_service_account_member]))
    invokers   = toset(concat(var.relay_service_iam.invokers, [local.compute_service_account_member]))
  }
  envvars = {
    "PROJECT_ID" : data.google_project.default.project_id,
    "RELAY_TOPIC_ID" : var.relay_topic_id,
    "RELAY_PROJECT_ID" : var.relay_project_id,
  }

  additional_service_annotations = { "run.googleapis.com/invoker-iam-disabled" : true }
}



resource "google_pubsub_subscription" "relay" {
  count = var.enable_relay_service ? 1 : 0

  project = var.project_id

  name  = "${var.prefix_name}-relay-sub"
  topic = google_pubsub_topic.default.name

  push_config {
    push_endpoint = module.relay_cloud_run[0].url
    oidc_token {
      service_account_email = local.compute_service_account_email
    }
  }

  expiration_policy {
    ttl = ""
  }

  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.dead_letter.id
    max_delivery_attempts = 5
  }
}
