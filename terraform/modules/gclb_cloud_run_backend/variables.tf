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
  type        = string
  description = "The Google Cloud project ID."
}

variable "region" {
  type        = string
  default     = "us-central1"
  description = "The default Google Cloud region to deploy resources in (defaults to 'us-central1')."
}

variable "name" {
  type        = string
  description = "The name of this project."
  validation {
    condition     = can(regex("^[a-z][0-9a-z-]+[0-9a-z]$", var.name))
    error_message = "Name can only contain lowercase letters, numbers, hyphens(-) and must start with letter. Name will be truncated and suffixed with at random string if it exceeds requirements for a given resource."
  }
}

variable "run_service_name" {
  type        = string
  description = "The name of the Cloud Run service to the compute backend serverless network endpoint group."
}

variable "domains" {
  type        = list(string)
  description = "Domain names to use for the HTTPS Global Load Balancer for the Cloud Run service (e.g. [\"my-project.e2e.tycho.joonix.net\"])."
}

variable "security_policy" {
  type        = string
  default     = null
  description = "Cloud Armor security policy for the load balancer."
}

variable "iap_config" {
  type = object({
    enable               = bool
    oauth2_client_id     = string
    oauth2_client_secret = string
  })
  default = {
    enable               = false
    oauth2_client_id     = ""
    oauth2_client_secret = ""
  }
  description = "Identity-Aware Proxy configuration for the load balancer."
}
