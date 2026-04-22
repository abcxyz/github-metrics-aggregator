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

module "webhook" {
  source = "../webhook"

  project_id = var.project_id

  prefix_name = var.prefix_name

  region = var.region

  image                             = var.image
  automation_service_account_member = var.automation_service_account_member
  events_topic_id                   = var.events_topic_id
  dlq_events_topic_id               = var.dead_letter_topic_id
  bigquery_project_id               = var.bigquery_project_id
  dataset_id                        = var.dataset_id
  optimized_events_table_id         = var.optimized_events_table_id
  failure_events_table_id           = var.failure_events_table_id
  enable_webhook_gclb               = var.enable_webhook_gclb
  webhook_domains                   = var.webhook_domains
}

module "backend" {
  source = "../backend"

  # Move to DB project!
  project_id = var.bigquery_project_id

  prefix_name = var.prefix_name

  region = var.region

  image                             = var.image
  automation_service_account_member = var.automation_service_account_member
  events_topic_id                   = var.events_topic_id
  dlq_events_topic_id               = var.dead_letter_topic_id
  relay_topic_id                    = var.relay_topic_id
  relay_project_id                  = var.relay_project_id
  secrets_to_create                 = var.secrets_to_create
  retry_job_timeout                 = var.retry_job_timeout
  github_app_id                     = var.github_app_id
  github_enterprise_server_url      = var.github_enterprise_server_url
  bigquery_project_id               = var.bigquery_project_id
  dataset_id                        = var.dataset_id
  optimized_events_table_id         = var.optimized_events_table_id
  checkpoint_table_id               = var.checkpoint_table_id
  bigquery_infra_deploy             = var.bigquery_infra_deploy
  retry_job_schedule                = var.retry_job_schedule
  retry_service_iam                 = var.retry_service_iam
  artifacts                         = var.artifacts
  commit_review_status              = var.commit_review_status
}
