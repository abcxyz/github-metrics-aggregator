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

variable "prefix_name" {
  description = "The prefix applied to all components."
  type        = string
  default     = "github-metrics"
  validation {
    condition     = can(regex("^[A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]$", var.prefix_name))
    error_message = "Name can only contain letters, numbers, hyphens(-) and must start with letter."
  }
}

variable "component_names" {
  description = "The name of each component."
  type = object({
    webhook_name = string
    retry_name   = string
  })
  default = {
    webhook_name = "webhook"
    retry_name   = "retry"
  }

  validation {
    condition     = can(regex("^[A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]$", var.component_names.webhook_name))
    error_message = "webhook_name can only contain letters, numbers, hyphens(-) and must start with letter."
  }

  validation {
    condition     = can(regex("^[A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]$", var.component_names.retry_name))
    error_message = "retry_name can only contain letters, numbers, hyphens(-) and must start with letter."
  }
}

variable "webhook_domain" {
  description = "Domain name for the Google Cloud Load Balancer used by the webhook."
  type        = string
}

variable "webhook_image" {
  description = "Cloud Run webhook service image name to deploy."
  type        = string
  default     = "gcr.io/cloudrun/hello:latest"
}

variable "retry_image" {
  description = "Cloud Run retry service image name to deploy."
  type        = string
  default     = "gcr.io/cloudrun/hello:latest"
}

variable "service_iam" {
  description = "IAM member bindings for the Cloud Run services."
  type = object({
    webhook = object({
      admins     = list(string)
      developers = list(string)
      invokers   = list(string)
    }),
    retry = object({
      admins     = list(string)
      developers = list(string)
      invokers   = list(string)
    })
  })
  default = {
    webhook = {
      admins     = []
      developers = []
      invokers   = []
    },
    retry = {
      admins     = []
      developers = []
      invokers   = []
    }
  }
}

variable "topic_iam" {
  description = "IAM member bindings for the PubSub ingestion topic."
  type = object({
    admins      = list(string)
    editors     = list(string)
    viewers     = list(string)
    publishers  = list(string)
    subscribers = list(string)
  })
  default = {
    admins      = []
    editors     = []
    viewers     = []
    publishers  = []
    subscribers = []
  }
}

variable "dead_letter_sub_iam" {
  description = "IAM member binding for the PubSub dead letter subscription."
  type = object({
    admins      = list(string)
    editors     = list(string)
    viewers     = list(string)
    subscribers = list(string)
  })
  default = {
    admins      = []
    editors     = []
    viewers     = []
    subscribers = []
  }
}

variable "dataset_location" {
  type        = string
  description = "The BigQuery dataset location."
  default     = "US"
}

variable "dataset_id" {
  type        = string
  description = "The BigQuery dataset id to create."
  default     = "github_metrics"
}

variable "dataset_iam" {
  description = "IAM member bindings for the BigQuery dataset."
  type = object({
    owners  = list(string)
    editors = list(string)
    viewers = list(string)
  })
  default = {
    owners  = []
    editors = []
    viewers = []
  }
}

variable "events_table_id" {
  description = "The BigQuery events table id to create."
  type        = string
  default     = "events"
}

variable "events_table_iam" {
  description = "IAM member bindings for the BigQuery events table."
  type = object({
    owners  = list(string)
    editors = list(string)
    viewers = list(string)
  })
  default = {
    owners  = []
    editors = []
    viewers = []
  }
}

variable "checkpoint_table_id" {
  description = "The BigQuery checkpoint table id to create."
  type        = string
  default     = "checkpoint"
}

variable "checkpoint_table_iam" {
  description = "IAM member bindings for the BigQuery checkpoint table."
  type = object({
    owners  = list(string)
    editors = list(string)
    viewers = list(string)
  })
  default = {
    owners  = []
    editors = []
    viewers = []
  }
}

variable "failure_events_table_id" {
  description = "The BigQuery failure events table id to create."
  type        = string
  default     = "failure_events"
}

variable "failure_events_table_iam" {
  description = "IAM member bindings for the BigQuery failure events table."
  type = object({
    owners  = list(string)
    editors = list(string)
    viewers = list(string)
  })
  default = {
    owners  = []
    editors = []
    viewers = []
  }
}

variable "event_delivery_retry_limit" {
  description = "Number of attempts to delivery a failed event from GitHub."
  type        = string
  default     = "10"
}

variable "execution_interval_minutes" {
  description = "Amount of time in minutes to append to the current time when calculating the lock TTL."
  type        = string
  default     = "10"
}

variable "execution_interval_clock_skew_ms" {
  description = "A conservative time estimate in ms to subtract from the current time to account for clock skew given the system can drift ahead."
  type        = string
  default     = "5000"
}

variable "cloud_scheduler_deadline_duration" {
  description = "The deadline for job attempts in seconds. If the request handler does not respond by this deadline then the request is cancelled and the attempt is marked as a DEADLINE_EXCEEDED failure. Defaults to 30 minutes (max)"
  type        = string
  default     = "1800s"
}

variable "cloud_scheduler_timezone" {
  description = "Specifies the time zone to be used in interpreting schedule."
  type        = string
  default     = "Etc/UTC"
}

variable "cloud_scheduler_schedule_cron" {
  description = "Cron expression that represents the schedule of the job. Default is every hour."
  type        = string
  default     = "*/1 * * * *"
}

variable "cloud_scheduler_retry_limit" {
  description = "Number of times Cloud Scheduler will retry the job when it "
  type        = string
  default     = "1"
}

variable "bigquery_project_id" {
  description = "The project ID where the BigQuery instance exists."
  type        = string
}

variable "github_app_id" {
  description = "The GitHub App ID."
  type        = string
}

variable "github_webhook_id" {
  description = "The GitHub webhook ID created in the GitHub App."
  type        = string
}

variable "region" {
  description = "The default Google Cloud region to deploy resources in (defaults to 'us-central1')."
  type        = string
  default     = "us-central1"
}
