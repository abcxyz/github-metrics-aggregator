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

variable "job_name" {
  type        = string
  description = "The name of the cloud run job"
  validation {
    condition     = can(regex("^[a-z](?:[-a-z0-9]{4,28}[a-z0-9])$", var.job_name))
    error_message = "job_name must match the regex: ^[a-z](?:[-a-z0-9]{4,28}[a-z0-9])$"
  }
}

variable "region" {
  type        = string
  description = "The cloud region to deploy this job in defaults to us-central1"
  default     = "us-central1"
}

variable "leech_table_id" {
  description = "The BigQuery leech table id to create."
  type        = string
  default     = "leech_status"
  nullable    = false
}

variable "leech_table_iam" {
  description = "IAM member bindings for the BigQuery leech table."
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

variable "artifacts_job_iam" {
  description = "IAM member bindings for the BigQuery Artifacts Cloud Run Job."
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

variable "leech_bucket_name" {
  description = "The name of the cloud storage bucket to store logs ingested by the leech pipeline."
  type        = string
  default     = "leech-df-store"
  nullable    = false
}

variable "leech_bucket_location" {
  description = "The location of the cloud storage bucket to store logs ingested by the leech pipeline."
  type        = string
  default     = "US"
  nullable    = false
}

variable "github_app_id" {
  description = "The GitHub App id of the application"
  type        = string
}

variable "github_install_id" {
  description = "The GitHub App installation id of the application"
  type        = string
}

variable "github_private_key_secret_id" {
  description = "The secret id containing the private key for the GitHub app. name"
  type        = string
}

variable "github_private_key_secret_version" {
  description = "The version of the secret containing the private key for the GitHub app"
  type        = string
  default     = "latest"
}

variable "events_table_id" {
  description = "The BigQuery events table id to create."
  type        = string
  nullable    = false
}

variable "scheduler_deadline_duration" {
  description = "The deadline for job attempts in seconds. If the request handler does not respond by this deadline then the request is cancelled and the attempt is marked as a DEADLINE_EXCEEDED failure. Defaults to 30 minutes."
  type        = string
  default     = "180s"
}

variable "scheduler_timezone" {
  description = "Specifies the time zone to be used in interpreting schedule."
  type        = string
  default     = "Etc/UTC"
}

variable "scheduler_cron" {
  description = "Cron expression that represents the schedule of the job. Default is every hour."
  type        = string
  nullable    = false
}

variable "scheduler_retry_limit" {
  description = "Number of times Cloud Scheduler will retry the job when it "
  type        = string
  default     = "0"
}

variable "image" {
  description = "Cloud Run service image for github-metrics-aggregator and server entrypoints."
  type        = string
}

variable "additional_env_vars" {
  description = "User supplied environment variables"
  type        = map(string)
}
