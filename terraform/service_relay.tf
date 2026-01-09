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

resource "google_service_account" "relay_run_service_account" {
  count = var.enable_relay_service ? 1 : 0

  project = data.google_project.default.project_id

  account_id   = "${var.prefix_name}-relay-sa"
  display_name = "${var.prefix_name}-relay-sa Cloud Run Service Account"
}

module "relay_cloud_run" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=1467eaf0115f71613727212b0b51b3f99e699842"

  count = var.enable_relay_service ? 1 : 0

  project_id = data.google_project.default.project_id

  name                  = "${var.prefix_name}-relay"
  region                = var.region
  image                 = var.image
  args                  = ["relay"]
  ingress               = "internal"
  service_account_email = google_service_account.relay_run_service_account[0].email
  service_iam = {
    admins     = toset(var.relay_service_iam.admins)
    developers = toset(concat(var.relay_service_iam.developers, [var.automation_service_account_member]))
    invokers   = toset(var.relay_service_iam.invokers)
  }
  envvars = {
    "PROJECT_ID" : data.google_project.default.project_id,
    "RELAY_TOPIC_ID" : var.relay_topic_id,
    "RELAY_PROJECT_ID" : var.relay_project_id,
  }

  additional_service_annotations = { "run.googleapis.com/invoker-iam-disabled" : true }
}

# allow the ci service account to act as the relay cloud run service account
# this allows the ci service account to deploy new revisions for the cloud run
# service
resource "google_service_account_iam_member" "relay_run_sa_user" {
  count = var.enable_relay_service ? 1 : 0

  service_account_id = google_service_account.relay_run_service_account[0].name
  role               = "roles/iam.serviceAccountUser"
  member             = var.automation_service_account_member
}

resource "google_service_account" "relay_sub_service_account" {
  count = var.enable_relay_service ? 1 : 0

  project = data.google_project.default.project_id

  account_id   = "${var.prefix_name}-relay-sub-sa"
  display_name = "${var.prefix_name}-relay-sub-sa PubSub Subscription Identity"
}

resource "google_pubsub_subscription" "relay" {
  count = var.enable_relay_service ? 1 : 0

  project = var.project_id

  name  = "${var.prefix_name}-relay-sub"
  topic = google_pubsub_topic.default.name

  push_config {
    push_endpoint = module.relay_cloud_run[0].url
    oidc_token {
      service_account_email = google_service_account.relay_sub_service_account[0].email
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

resource "google_cloud_run_service_iam_member" "relay_invoker" {
  count = var.enable_relay_service ? 1 : 0

  project = data.google_project.default.project_id

  location = var.region
  service  = module.relay_cloud_run[0].service_name

  role   = "roles/run.invoker"
  member = google_service_account.relay_sub_service_account[0].member
}
