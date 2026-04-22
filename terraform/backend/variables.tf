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

variable "project_id" {
  description = "The project ID to deploy resources into."
  type        = string
}

variable "prefix_name" {
  description = "The prefix name for resources."
  type        = string
}

variable "region" {
  description = "The region to deploy resources into."
  type        = string
  default     = "us-central1"
}

variable "image" {
  description = "The container image to use for Cloud Run services and jobs."
  type        = string
}

variable "automation_service_account_member" {
  description = "The IAM member string for the automation service account."
  type        = string
}

# Relay Service Variables
variable "enable_relay_service" {
  description = "Whether to enable the relay service."
  type        = bool
  default     = false
}

variable "relay_service_iam" {
  description = "IAM bindings for the relay service."
  type = object({
    admins     = list(string)
    developers = list(string)
    invokers   = list(string)
  })
  default = {
    admins     = []
    developers = []
    invokers   = []
  }
}

variable "relay_topic_id" {
  description = "The Pub/Sub topic ID for the relay."
  type        = string
}

variable "relay_project_id" {
  description = "The project ID for the relay Pub/Sub topic."
  type        = string
}

variable "events_topic_id" {
  description = "The Pub/Sub topic ID for events (incoming from webhook)."
  type        = string
}

variable "dlq_events_topic_id" {
  description = "The Pub/Sub topic ID for dead-letter events."
  type        = string
}

# Retry Job Variables
variable "secrets_to_create" {
  description = "List of secrets to create in Secret Manager."
  type        = list(string)
  default     = []
}

variable "retry_job_timeout" {
  description = "The timeout for the retry job."
  type        = string
  default     = "600s"
}

variable "github_app_id" {
  description = "The GitHub App ID."
  type        = string
}

variable "github_enterprise_server_url" {
  description = "The GitHub Enterprise Server URL."
  type        = string
  default     = ""
}

variable "bigquery_project_id" {
  description = "The project ID for BigQuery dataset."
  type        = string
}

variable "dataset_id" {
  description = "The BigQuery dataset ID."
  type        = string
}

variable "optimized_events_table_id" {
  description = "The BigQuery table ID for optimized events."
  type        = string
}

variable "checkpoint_table_id" {
  description = "The BigQuery table ID for checkpoints."
  type        = string
}

variable "bigquery_infra_deploy" {
  description = "Whether to deploy BigQuery infrastructure permissions."
  type        = bool
  default     = true
}

variable "retry_job_schedule" {
  description = "The schedule for the retry job."
  type        = string
  default     = "*/5 * * * *"
}

variable "retry_service_iam" {
  description = "IAM bindings for the retry job."
  type = object({
    developers = list(string)
  })
  default = {
    developers = []
  }
}

# Artifacts Job Variables
variable "artifacts" {
  description = "Configuration for the artifacts job."
  type = object({
    enabled                 = bool
    bucket_name             = optional(string, null)
    bucket_location         = optional(string, null)
    job_name                = string
    table_id                = string
    job_additional_env_vars = map(string)
    scheduler_cron          = string
    job_iam = object({
      admins     = list(string)
      developers = list(string)
      invokers   = list(string)
    })
  })
}

variable "github_private_key_secret_id" {
  description = "The Secret Manager secret ID for the GitHub private key."
  type        = string
  default     = "github-private-key"
}

# Commit Review Status Variables
variable "commit_review_status" {
  description = "Configuration for the commit review status job."
  type = object({
    enabled                 = bool
    job_name                = string
    table_id                = string
    job_additional_env_vars = map(string)
    scheduler_cron          = string
    job_iam = object({
      admins     = list(string)
      developers = list(string)
      invokers   = list(string)
    })
  })
}
