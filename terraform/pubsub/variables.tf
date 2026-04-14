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
  description = "The project ID."
  type        = string
}

variable "prefix_name" {
  description = "The prefix to apply to all resources."
  type        = string
  default     = "gma"
}


variable "relay_project_id" {
  description = "The project ID where the relay topic exists."
  type        = string
  default     = ""
}

variable "relay_topic_id" {
  description = "The topic ID for the relay service."
  type        = string
  default     = "gma-relay"
}

variable "relay_sub_service_account_email" {
  description = "The service account email for the relay subscription."
  type        = string
  default     = ""
}

variable "dead_letter_topic_id" {
  description = "The ID of the dead letter topic."
  type        = string
  default     = ""
}

variable "dataset_id" {
  description = "The bigquery dataset ID."
  type        = string
  default     = "github_metrics"
}

variable "optimized_events_table_id" {
  description = "The optimized events table ID."
  type        = string
  default     = "optimized_events"
}

variable "relay_schema_id" {
  description = "The ID of the enriched layout schema."
  type        = string
  default     = ""
}

variable "relay_publisher_member" {
  description = "The service account member allowed to publish to the relay topic."
  type        = string
  default     = ""
}
