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
  default     = "github_metrics"
}

variable "dataset_location" {
  description = "The location for the BigQuery dataset."
  type        = string
  default     = "US"
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





variable "bigquery_events_partition_granularity" {
  description = "The parition granularity for the raw_events_table, can be HOUR, DAY, MONTH, or YEAR."
  type        = string
  default     = "DAY"
}



variable "checkpoint_table_id" {
  description = "The ID of the BigQuery table for checkpoints."
  type        = string
  default     = "checkpoint"
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
  default     = "failure_events"
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

variable "job_users" {
  description = "A list of service accounts/groups that should be granted roles/bigquery.jobUser at the project level to run queries."
  type        = list(string)
  default     = []
}

variable "dataset_metadata_viewers" {
  description = "A list of service accounts/groups that should be granted roles/bigquery.metadataViewer for the dataset."
  type        = list(string)
  default     = []
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




variable "prefix_name" {
  description = "The prefix to apply to all resources."
  type        = string
  default     = "gma"
}

variable "optimized_events_table_id" {
  description = "The ID of the BigQuery table for optimized events."
  type        = string
  default     = "optimized_events"
}

variable "events_table_iam" {
  description = "IAM bindings for the optimized events table in {GROUP_TYPE => [MEMBERS]} format."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {}
}



variable "prstats_pull_requests_table_id" {
  description = "The ID of the BigQuery table for prstats pull requests."
  type        = string
  default     = "gma_prstats_pull_requests"
}

variable "prstats_pull_request_reviews_table_id" {
  description = "The ID of the BigQuery table for prstats pull request reviews."
  type        = string
  default     = "gma_prstats_pull_request_reviews"
}

variable "prstats_table_id" {
  description = "The ID of the BigQuery table for prstats."
  type        = string
  default     = "gma_prstats"
}




variable "bigquery_prstats_partition_granularity" {
  description = "The parition granularity for the prstats tables, can be HOUR, DAY, MONTH, or YEAR."
  type        = string
  default     = "DAY"
}

variable "integration_events_table_id" {
  description = "The ID of the BigQuery table for integration events."
  type        = string
  default     = "gma_integration_events"
}

variable "artifacts_table_id" {
  description = "The ID of the BigQuery table for artifacts status."
  type        = string
  default     = "artifacts_status"
}

variable "artifacts_table_iam" {
  description = "IAM bindings for the artifacts table."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {}
}

variable "commit_review_status_table_id" {
  description = "The ID of the BigQuery table for commit review status."
  type        = string
  default     = "commit_review_status"
}

variable "commit_review_status_table_iam" {
  description = "IAM bindings for the commit review status table."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {}
}
