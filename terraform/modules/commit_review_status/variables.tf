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
  description = "The GCP project ID."
  type        = string
}

variable "dataset_id" {
  type        = string
  description = "The BigQuery dataset id to create."
}

variable "commit_review_status_table_id" {
  description = "The BigQuery commit review status table id to create."
  type        = string
  default     = "commit_review_status"
}

variable "commit_review_status_iam" {
  description = "IAM member bindings for the BigQuery commit review status table."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {
    owners  = []
    editors = []
    viewers = []
  }
}
