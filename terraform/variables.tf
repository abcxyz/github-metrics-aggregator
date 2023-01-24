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

variable "name" {
  description = "The name of this component."
  type        = string
  validation {
    condition     = can(regex("^[A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]$", var.name))
    error_message = "Name can only contain letters, numbers, hyphens(-) and must start with letter."
  }
}

variable "domain" {
  description = "Domain name for the Google Cloud Load Balancer."
  type        = string
}

variable "image" {
  description = "Cloud Run service image name to deploy."
  type        = string
  default     = "gcr.io/cloudrun/hello:latest"
}

variable "service_iam" {
  description = "IAM bindings in {ROLE => [MEMBERS]} format for the Cloud Run service."
  type        = map(list(string))
  default     = {}
}

variable "topic_iam" {
  description = "IAM bindings in {ROLE => [MEMBERS]} format for the PubSub topic."
  type        = map(list(string))
  default     = {}
}

variable "dataset_location" {
  type        = string
  description = "The BigQuery dataset location."
  default     = "US"
}

variable "dataset_id" {
  type        = string
  description = "The BigQuery dataset id to create."
}

variable "dataset_iam" {
  description = "IAM bindings in {ROLE => [MEMBERS]} format for the BigQuery github_webhook dataset."
  type        = map(list(string))
  default     = {}
}

variable "table_id" {
  type        = string
  description = "The BigQuery table id to create."
}

variable "table_iam" {
  description = "IAM bindings in {ROLE => [MEMBERS]} format for the BigQuery github_webhook.events table."
  type        = map(list(string))
  default     = {}
}
