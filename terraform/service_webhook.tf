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

module "gclb" {
  count = var.enable_webhook_gclb ? 1 : 0

  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/gclb_cloud_run_backend?ref=129d7928a4ed16d7b51ea5aca9df77018d8e7632"

  project_id = data.google_project.default.project_id

  name             = "${var.prefix_name}-webhook"
  run_service_name = module.webhook_cloud_run.service_name
  domains          = var.webhook_domains
}

resource "google_service_account" "webhook_run_service_account" {
  project = data.google_project.default.project_id

  account_id   = "${var.prefix_name}-webhook-sa"
  display_name = "${var.prefix_name}-webhook-sa Cloud Run Service Account"
}

module "webhook_cloud_run" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=129d7928a4ed16d7b51ea5aca9df77018d8e7632"

  project_id = data.google_project.default.project_id

  name                  = "${var.prefix_name}-webhook"
  region                = var.region
  image                 = var.image
  args                  = ["webhook", "server"]
  ingress               = var.enable_webhook_gclb ? "internal-and-cloud-load-balancing" : "all"
  secrets               = ["github-webhook-secret"]
  service_account_email = google_service_account.webhook_run_service_account.email
  service_iam           = var.webhook_service_iam
  envvars = {
    "BIG_QUERY_PROJECT_ID" : var.bigquery_project_id,
    "DATASET_ID" : google_bigquery_dataset.default.dataset_id,
    "EVENTS_TABLE_ID" : google_bigquery_table.events_table.table_id,
    "FAILURE_EVENTS_TABLE_ID" : google_bigquery_table.failure_events_table.table_id,
    "PROJECT_ID" : data.google_project.default.project_id,
    "RETRY_LIMIT" : var.event_delivery_retry_limit,
    "EVENTS_TOPIC_ID" : google_pubsub_topic.default.name,
    "DLQ_EVENTS_TOPIC_ID" : google_pubsub_topic.dead_letter.name,
  }
  secret_envvars = {
    "GITHUB_WEBHOOK_SECRET" : {
      name : "github-webhook-secret",
      version : "latest",
    }
  }
}

# allow the ci service account to act as the webhook cloud run service account
# this allows the ci service account to deploy new revisions for the cloud run
# sevice
resource "google_service_account_iam_member" "webhook_run_sa_user" {
  service_account_id = google_service_account.webhook_run_service_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = local.automation_service_account_member
}
