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

locals {
  webhook_service_window = 2 * local.hour + 10 * local.minute
}

module "gclb" {
  count = var.enable_webhook_gclb ? 1 : 0

  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/gclb_cloud_run_backend?ref=17dbc86af5b4e85237829515caad81da77289743"

  project_id = data.google_project.default.project_id

  name             = "${var.prefix_name}-webhook"
  run_service_name = module.webhook_cloud_run.service_name
  domains          = var.webhook_domains
}



module "webhook_cloud_run" {
  source = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=1467eaf0115f71613727212b0b51b3f99e699842"

  project_id = data.google_project.default.project_id

  name                  = "${var.prefix_name}-webhook"
  region                = var.region
  image                 = var.image
  args                  = ["webhook", "server"]
  ingress               = var.enable_webhook_gclb ? "internal-and-cloud-load-balancing" : "all"
  secrets               = ["github-webhook-secret"]
  service_account_email = local.compute_service_account_email
  service_iam = {
    admins     = toset(var.webhook_service_iam.admins)
    developers = toset(concat(var.webhook_service_iam.developers, [var.automation_service_account_member]))
    invokers   = toset(var.webhook_service_iam.invokers)
  }
  envvars = {
    "BIG_QUERY_PROJECT_ID" : var.bigquery_project_id,
    "DATASET_ID" : var.dataset_id,
    "EVENTS_TABLE_ID" : var.optimized_events_table_id,
    "FAILURE_EVENTS_TABLE_ID" : var.failure_events_table_id,
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

  additional_service_annotations = { "run.googleapis.com/invoker-iam-disabled" : true }

  max_instances = var.webhook_max_instances
}

