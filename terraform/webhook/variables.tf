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
  description = "The GCP project ID."
  type        = string
}

variable "prefix_name" {
  description = "The prefix applied to all components."
  type        = string
  default     = "github-metrics"
}

variable "enable_webhook_gclb" {
  description = "Enable the use of a Google Cloud load balancer for the webhook Cloud Run service."
  type        = bool
  default     = true
}

variable "webhook_domains" {
  description = "Domain names for the Google Cloud Load Balancer used by the webhook."
  type        = list(string)
  default     = []
}

variable "region" {
  description = "The default Google Cloud region to deploy resources in."
  type        = string
  default     = "us-central1"
}

variable "image" {
  description = "Cloud Run service image for github-metrics-aggregator and server entrypoints."
  type        = string
}

variable "webhook_service_iam" {
  description = "IAM member bindings for the webhook Cloud Run services."
  type = object({
    admins     = optional(list(string), [])
    developers = optional(list(string), [])
    invokers   = optional(list(string), [])
  })
  default = {}
}

variable "automation_service_account_member" {
  description = "The service account member used for deploying new revisions"
  type        = string
}

variable "bigquery_project_id" {
  description = "The project ID where the BigQuery instance exists."
  type        = string
}

variable "dataset_id" {
  type        = string
  description = "The BigQuery dataset id."
  default     = "github_metrics"
}

variable "optimized_events_table_id" {
  description = "The ID of the BigQuery table for optimized events."
  type        = string
  default     = "optimized_events"
}

variable "failure_events_table_id" {
  description = "The BigQuery failure events table id to create."
  type        = string
  default     = "failure_events"
}

variable "event_delivery_retry_limit" {
  description = "Number of attempts to delivery a failed event from GitHub."
  type        = string
  default     = "10"
}

variable "events_topic_id" {
  description = "The PubSub topic ID for events."
  type        = string
}

variable "dlq_events_topic_id" {
  description = "The PubSub topic ID for dead-letter events."
  type        = string
}

variable "webhook_max_instances" {
  type        = string
  default     = "10"
  description = "The max number of instances for the Webhook Cloud Run service."
}
