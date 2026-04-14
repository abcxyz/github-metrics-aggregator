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

variable "service_iam" {
  description = "IAM member bindings for the Cloud Run service."
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

variable "secrets" {
  type        = list(any)
  default     = []
  description = "Secret Manager secrets to be created with a value of 'DEFAULT_VALUE'."
}

variable "min_instances" {
  type        = string
  default     = "0"
  description = "The max number of instances for the Cloud Run service (defaults to '0')."

}

variable "max_instances" {
  type        = string
  default     = "10"
  description = "The max number of instances for the Cloud Run service (defaults to '10')."

}

variable "execution_environment" {
  type        = string
  default     = "gen1"
  description = "The Cloud Run execution environment, possible values are: gen1, gen2 (defaults to 'gen1')."
}

variable "resources" {
  type = object({
    requests = object({
      cpu    = string
      memory = string
    })
    limits = object({
      cpu    = string
      memory = string
    })
  })
  default = {
    requests = {
      cpu    = "1000m"
      memory = "512Mi" # minimum for gen2 runtime environment
    }
    limits = {
      cpu    = "1000m"
      memory = "512Mi"
    }
  }
  description = "The compute resource requests and limits for the Cloud Run service."
}

variable "ingress" {
  type        = string
  default     = "all"
  description = "The ingress settings for the Cloud Run service, possible values: all, internal, internal-and-cloud-load-balancing (defaults to 'all')."
}

variable "image" {
  type        = string
  description = "The container image for the Cloud Run service."
}

variable "service_account_email" {
  type        = string
  description = "The service account email for Cloud Run to run as."
}

variable "envvars" {
  type        = map(string)
  default     = {}
  description = "Environment variables for the Cloud Run service (plain text)."
}

variable "secret_envvars" {
  type = map(object({
    name    = string
    version = string
  }))
  default     = {}
  description = "Secret environment variables for the Cloud Run service (Secret Manager)."
}

variable "secret_volumes" {
  type = map(object({
    name    = string
    version = string
  }))
  default     = {}
  description = "Volume mounts for the Cloud Run service (Secret Manager)."
}

variable "args" {
  type        = list(string)
  default     = []
  description = "Arguments to the cloud run container's entrypoint."
}

variable "additional_revision_annotations" {
  type        = map(string)
  default     = {}
  description = "Annotations to add to the template.metadata.annotations field."
}

variable "startup_probe" {
  type = object({
    initial_delay_seconds = optional(number, 0)
    timeout_seconds       = optional(number, 1)
    period_seconds        = optional(number, 10)
    failure_threshold     = optional(number, 3)
    http_get = optional(object({
      http_headers = optional(map(string), {})
      path         = optional(string)
      port         = optional(number)
    }), null)
  })
  default     = null
  description = "Optional startup probe configuration"
}
