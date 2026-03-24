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

variable "dataset_id" {
  description = "The dataset ID."
  type        = string
}

variable "table_id" {
  description = "The table ID."
  type        = string
}

variable "schema" {
  description = "The schema for the table in JSON format."
  type        = string
}

variable "deletion_protection" {
  description = "Whether to protect the table from deletion."
  type        = bool
  default     = true
}

variable "time_partitioning" {
  description = "Optional time partitioning configuration."
  type = object({
    type  = string
    field = optional(string)
  })
  default = null
}

variable "clustering" {
  description = "Optional clustering configuration."
  type        = list(string)
  default     = []
}

variable "iam" {
  description = "IAM bindings for the table in {GROUP_TYPE => [MEMBERS]} format."
  type = object({
    owners  = optional(list(string), [])
    editors = optional(list(string), [])
    viewers = optional(list(string), [])
  })
  default = {}
}
