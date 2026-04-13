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

# ------------------------------------------------------------------------------
# Terraform Settings & Providers
# ------------------------------------------------------------------------------

terraform {
  required_version = ">= 1.5.7"

  required_providers {
    google = {
      version = ">= 5.19"
      source  = "hashicorp/google"
    }
  }
}

provider "google" {
  project = local.project_id
  region  = local.region
}

# ------------------------------------------------------------------------------
# Locals Configuration
# Use this block to store all variables. Do not hardcode values in module calls.
# ------------------------------------------------------------------------------

locals {
  # --- 1. Core / Shared Variables ---
  project_id          = "<YOUR_PROJECT_ID>"
  bigquery_project_id = "<YOUR_PROJECT_ID>"
  project_number      = "<YOUR_PROJECT_NUMBER>"
  region              = "us-central1"
  dataset_location    = "US"
  prefix_name         = "github-metrics"

  # --- 2. BigQuery Dataset & Tables ---
  dataset_id                            = "github_metrics"
  events_table_id                       = "events"
  raw_events_table_id                   = "raw_events"
  checkpoint_table_id                   = "checkpoint"
  failure_events_table_id               = "failure_events"
  optimized_events_table_id             = "optimized_events"
  artifacts_table_id                    = "artifacts_status"
  commit_review_status_table_id         = "commit_review_status"
  bigquery_events_partition_granularity = "DAY"


  # IAM Bindings for BigQuery (grouped as objects)


  # IAM for tables, adding service accounts as editors
  events_table_iam = {
    editors = [local.compute_service_account_member]
  }
  checkpoint_table_iam = {
    editors = [local.compute_service_account_member]
  }
  failure_events_table_iam = {
    editors = [local.compute_service_account_member]
  }

  # Generic IAM lists for BigQuery
  job_users                = [local.compute_service_account_member]
  dataset_metadata_viewers = [local.compute_service_account_member]


  # --- 3. GMA (Application) Config ---
  image                             = "<YOUR_DOCKER_IMAGE_PATH>" # e.g. gcr.io/your-project/gma:latest
  automation_service_account_member = "<AUTOMATION_SA_MEMBER>"   # e.g. serviceAccount:deployer@...
  github_app_id                     = "<YOUR_GITHUB_APP_ID>"
  github_private_key_secret_id      = "github-private-key"
  github_private_key_secret_version = "latest"
  enable_webhook_gclb               = true
  webhook_domains                   = [] # e.g., ["gma.yourdomain.com"]

  # Service Account emails/members used by bigquery to grant permissions
  # In a standard chained setup, these would come from module outputs.
  # Static string placeholders used here per strict local requirement.
  compute_service_account_member = "<COMPUTE_SA_MEMBER>" # e.g., serviceAccount:compute-sa@...

  # GMA IAM & Feature Toggles
  lock_ttl                        = "5m"
  lock_ttl_clock_skew             = "10s"
  enable_relay_service            = false
  relay_project_id                = ""
  relay_topic_id                  = ""
  relay_sub_service_account_email = ""
  dead_letter_topic_id            = ""

  # Dashboards and Analytics
  github_metrics_dashboard = {
    enabled = false
    viewers = []
  }
  enable_monitoring_dashboard = false

  # --- 4. Scheduled Queries Config ---
  prstats_service_account_email         = "<PRSTATS_SA_EMAIL>" # e.g., prstats@your-project.iam.gserviceaccount.com
  prstats_pull_requests_schedule        = "every 30 mins"
  prstats_pull_request_reviews_schedule = "every 30 mins"
  integration_events_schedule           = "every 30 mins"
  prstats_schedule                      = "every 30 mins"

  # Table IDs specific to Scheduled Queries mapping
  prstats_pull_requests_table_id        = "gma_prstats_pull_requests"
  prstats_pull_request_reviews_table_id = "gma_prstats_pull_request_reviews"
  prstats_table_id                      = "gma_prstats"
  integration_events_table_id           = "gma_integration_events"

  # IAM for Scheduled Queries tables
  bigquery_prstats_partition_granularity = "DAY"
}

# ------------------------------------------------------------------------------
# Module Definitions
# ------------------------------------------------------------------------------

module "bigquery" {
  source = "./bigquery"

  # Core
  project_id = local.bigquery_project_id

  prefix_name      = local.prefix_name
  dataset_id       = local.dataset_id
  dataset_location = local.dataset_location

  # Dataset & Tables Configuration
  dataset_iam                           = local.dataset_iam
  events_table_id                       = local.events_table_id
  raw_events_table_id                   = local.raw_events_table_id
  bigquery_events_partition_granularity = local.bigquery_events_partition_granularity
  events_table_iam                      = local.events_table_iam
  checkpoint_table_id                   = local.checkpoint_table_id
  checkpoint_table_iam                  = local.checkpoint_table_iam
  failure_events_table_id               = local.failure_events_table_id
  failure_events_table_iam              = local.failure_events_table_iam
  optimized_events_table_id             = local.optimized_events_table_id
  artifacts_table_id                    = local.artifacts_table_id
  commit_review_status_table_id         = local.commit_review_status_table_id


  # Service Integration IAM
  job_users                = local.job_users
  dataset_metadata_viewers = local.dataset_metadata_viewers


  # Relay / Subscriptions (Optional)
  enable_relay_service            = local.enable_relay_service
  relay_sub_service_account_email = local.relay_sub_service_account_email

  # Scheduled Queries Table Mapping (For query creations or setups)
  prstats_pull_requests_table_id         = local.prstats_pull_requests_table_id
  prstats_pull_request_reviews_table_id  = local.prstats_pull_request_reviews_table_id
  prstats_table_id                       = local.prstats_table_id
  integration_events_table_id            = local.integration_events_table_id
  bigquery_prstats_partition_granularity = local.bigquery_prstats_partition_granularity
}

module "pubsub" {
  source = "./pubsub"

  project_id = local.bigquery_project_id

  prefix_name = local.prefix_name

  relay_project_id                = local.relay_project_id
  relay_topic_id                  = local.relay_topic_id
  relay_sub_service_account_email = local.relay_sub_service_account_email
  dead_letter_topic_id            = local.dead_letter_topic_id

  dataset_id                = module.bigquery.dataset_id
  optimized_events_table_id = module.bigquery.optimized_events_table_id

  relay_schema_id        = try(module.gma.relay_pubsub_schema_id, "")
  relay_publisher_member = try(module.gma.relay_run_service.service_account_member, "")
}









module "gma" {
  source = "./gma"

  # Core
  project_id = local.project_id

  region      = local.region
  prefix_name = local.prefix_name

  # Deploy & Image Parameters
  image                             = local.image
  automation_service_account_member = local.automation_service_account_member
  github_app_id                     = local.github_app_id
  github_private_key_secret_id      = local.github_private_key_secret_id

  # Networking / Load Balancing
  enable_webhook_gclb = local.enable_webhook_gclb
  webhook_domains     = local.webhook_domains

  # Dataset / Storage references
  dataset_id              = local.dataset_id
  events_table_id         = local.events_table_id
  raw_events_table_id     = local.raw_events_table_id
  checkpoint_table_id     = local.checkpoint_table_id
  failure_events_table_id = local.failure_events_table_id

  # Cloud Runner config/locks
  bigquery_project_id = local.bigquery_project_id # Using core project ID

  # Relay setup
  enable_relay_service = local.enable_relay_service
  relay_topic_id       = try(module.pubsub.relay_topic_name, "")
  relay_project_id     = try(module.pubsub.relay_topic_project, "")
  relay_service_iam    = local.relay_service_iam
}

module "scheduled_queries" {
  source = "./scheduled_queries"

  # Core
  project_id = local.project_id

  project_number = local.project_number
  location       = local.dataset_location # Mapped to dataset_location from scheduled queries `location`

  # Service Account Runner
  prstats_service_account_email = local.prstats_service_account_email

  # Dataset Config
  dataset_id = local.dataset_id

  # Source & Targets (Maps optimized structure)
  optimized_events_table_name             = local.optimized_events_table_id
  prstats_source_table_name               = local.optimized_events_table_id # Set source to optimized_events per default
  prstats_pull_requests_table_name        = local.prstats_pull_requests_table_id
  prstats_pull_request_reviews_table_name = local.prstats_pull_request_reviews_table_id
  prstats_table_name                      = local.prstats_table_id
  integration_events_table_name           = local.integration_events_table_id

  # Schedule Config
  prstats_pull_requests_schedule        = local.prstats_pull_requests_schedule
  prstats_pull_request_reviews_schedule = local.prstats_pull_request_reviews_schedule
  prstats_schedule                      = local.prstats_schedule
  integration_events_schedule           = local.integration_events_schedule
}
