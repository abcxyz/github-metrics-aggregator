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

variable "project_id" {
  description = "The project ID to deploy to."
  type        = string
}

variable "dataset_id" {
  description = "The ID of the BigQuery dataset."
  type        = string
}

variable "dataset_location" {
  description = "The location for the BigQuery dataset."
  type        = string
}

variable "dataset_iam" {
  description = "IAM bindings for the dataset in {GROUP_TYPE => [MEMBERS]} format."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {}
}

variable "events_table_id" {
  description = "The ID of the BigQuery table for events."
  type        = string
}

variable "raw_events_table_id" {
  description = "The ID of the BigQuery table for raw events."
  type        = string
}

variable "bigquery_events_partition_granularity" {
  description = "The parition granularity for the raw_events_table, can be HOUR, DAY, MONTH, or YEAR."
  type        = string
  default     = "DAY"
}

variable "events_table_iam" {
  description = "IAM bindings for the events table in {GROUP_TYPE => [MEMBERS]} format."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {}
}

variable "checkpoint_table_id" {
  description = "The ID of the BigQuery table for checkpoints."
  type        = string
}

variable "checkpoint_table_iam" {
  description = "IAM bindings for the checkpoint table in {GROUP_TYPE => [MEMBERS]} format."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {}
}

variable "failure_events_table_id" {
  description = "The ID of the BigQuery table for failure events."
  type        = string
}

variable "failure_events_table_iam" {
  description = "IAM bindings for the failure events table in {GROUP_TYPE => [MEMBERS]} format."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {}
}

variable "webhook_run_service_account_member" {
  description = "The service account member for the webhook service."
  type        = string
}

variable "retry_run_service_account_member" {
  description = "The service account member for the retry service."
  type        = string
}

variable "invocation_comment" {
  description = "The invocation comment table configuration."
  type = object({
    enabled  = optional(bool, false)
    table_id = optional(string)
    table_iam = optional(object({
      owners  = optional(list(string), [])
      editors = optional(list(string), [])
      viewers = optional(list(string), [])
    }), {})
  })
  default = {}
}

variable "github_metrics_dashboard" {
  description = "The github metrics dashboard configuration."
  type = object({
    enabled = optional(bool, false)
    viewers = optional(list(string), [])
  })
  default = {
    enabled = false
    viewers = []
  }
}

variable "relay_sub_service_account_email" {
  description = "The service account email for the relay subscription."
  type        = string
  default     = ""
}

variable "enable_relay_service" {
  description = "Whether to enable relay service resources."
  type        = bool
  default     = false
}

variable "relay_project_id" {
  description = "The project ID where the relay topic exists."
  type        = string
  default     = ""
}

variable "relay_topic_id" {
  description = "The topic ID for the relay service."
  type        = string
  default     = ""
}

variable "prefix_name" {
  description = "The prefix to apply to all resources."
  type        = string
}

variable "optimized_events_table_id" {
  description = "The ID of the BigQuery table for optimized events."
  type        = string
}

variable "optimized_events_table_iam" {
  description = "IAM bindings for the optimized events table in {GROUP_TYPE => [MEMBERS]} format."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {}
}

variable "dead_letter_topic_id" {
  description = "The ID of the dead letter topic."
  type        = string
  default     = ""
}
