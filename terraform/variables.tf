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

variable "automation_service_account_member" {
  description = "The service account member used for deploying new revisions"
  type        = string
}

# This current approach allows the end-user to disable the GCLB in favor of calling the Cloud Run service directly.
# This was done to use tagged revision URLs for integration testing on multiple pull requests.
# TODO(https://github.com/abcxyz/github-metrics-aggregator/issues/45)
variable "enable_webhook_gclb" {
  description = "Enable the use of a Google Cloud load balancer for the webhook Cloud Run service. By default this is true, this should only be used for integration environments where services will use tagged revision URLs for testing."
  type        = bool
  default     = true
}

variable "webhook_domains" {
  description = "Domain names for the Google Cloud Load Balancer used by the webhook."
  type        = list(string)
  default     = []
}

variable "image" {
  description = "Cloud Run service image for github-metrics-aggregator and server entrypoints."
  type        = string
}

variable "webhook_service_iam" {
  description = "IAM member bindings for the webhook Cloud Run services."
  type = object({
    admins     = optional(list(string), [])
    developers = optional(list(string), [])
    invokers   = optional(list(string), [])
  })
  default = {
    admins     = []
    developers = []
    invokers   = []
  }
}

variable "retry_service_iam" {
  description = "IAM member bindings for the retry Cloud Run services."
  type = object({
    admins     = optional(list(string), [])
    developers = optional(list(string), [])
    invokers   = optional(list(string), [])
  })
  default = {
    admins     = []
    developers = []
    invokers   = []
  }
}

variable "events_topic_iam" {
  description = "IAM member bindings for the events PubSub ingestion topic."
  type = object({
    admins      = optional(list(string), [])
    editors     = optional(list(string), [])
    viewers     = optional(list(string), [])
    publishers  = optional(list(string), [])
    subscribers = optional(list(string), [])
  })
  default = {
    admins      = []
    editors     = []
    viewers     = []
    publishers  = []
    subscribers = []
  }
}

variable "dlq_topic_iam" {
  description = "IAM member bindings for the events PubSub dead-letter topic."
  type = object({
    admins      = optional(list(string), [])
    editors     = optional(list(string), [])
    viewers     = optional(list(string), [])
    publishers  = optional(list(string), [])
    subscribers = optional(list(string), [])
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
    admins      = optional(list(string), [])
    editors     = optional(list(string), [])
    viewers     = optional(list(string), [])
    subscribers = optional(list(string), [])
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

variable "events_table_id" {
  description = "The BigQuery events table id to create."
  type        = string
  default     = "events"
}

variable "raw_events_table_id" {
  description = "The BigQuery raw_events table id to create."
  type        = string
  default     = "raw_events"
}

variable "events_table_iam" {
  description = "IAM member bindings for the BigQuery events table."
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

variable "checkpoint_table_id" {
  description = "The BigQuery checkpoint table id to create."
  type        = string
  default     = "checkpoint"
}

variable "checkpoint_table_iam" {
  description = "IAM member bindings for the BigQuery checkpoint table."
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

variable "failure_events_table_id" {
  description = "The BigQuery failure events table id to create."
  type        = string
  default     = "failure_events"
}

variable "failure_events_table_iam" {
  description = "IAM member bindings for the BigQuery failure events table."
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

variable "event_delivery_retry_limit" {
  description = "Number of attempts to delivery a failed event from GitHub."
  type        = string
  default     = "10"
}

variable "lock_ttl" {
  description = "Amount of time to append to the current time when calculating the lock TTL."
  type        = string
  default     = "5m"
}

variable "lock_ttl_clock_skew" {
  description = "A conservative time estimate to subtract from the current time to account for clock skew given the system can drift ahead."
  type        = string
  default     = "10s"
}

variable "cloud_scheduler_deadline_duration" {
  description = "The deadline for job attempts in seconds. If the request handler does not respond by this deadline then the request is cancelled and the attempt is marked as a DEADLINE_EXCEEDED failure. Defaults to 30 minutes."
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
  default     = "0 * * * *"
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

variable "bigquery_events_partition_granularity" {
  description = "How granular you want partition to be. Ideally partitions will be 1-10 GB. Our org had ~750 MB/month. Valid values are HOUR, DAY, MONTH, and YEAR."
  type        = string
  default     = "MONTH"
  validation {
    condition     = contains(["HOUR", "DAY", "MONTH", "YEAR"], var.bigquery_events_partition_granularity)
    error_message = "ERROR: bigquery_events_partition_granularity must be one of HOUR, DAY, MONTH, and YEAR"
  }
}

variable "github_app_id" {
  description = "The GitHub App ID."
  type        = string
}

variable "github_install_id" {
  description = "The GitHub App installation ID."
  type        = string
}

variable "region" {
  description = "The default Google Cloud region to deploy resources in (defaults to 'us-central1')."
  type        = string
  default     = "us-central1"
}

variable "log_mode" {
  description = "The mode for logging. Defaults to production."
  type        = string
  default     = "production"
}

variable "log_level" {
  description = "The level for logging. Defaults to warning."
  type        = string
  default     = "warning"
}

variable "leech" {
  description = "The configuration block for leech table"
  type = object({
    enabled  = bool
    table_id = optional(string, null)
    table_iam = optional(object({
      owners  = optional(list(string), [])
      editors = optional(list(string), [])
      viewers = optional(list(string), [])
    }), null)
    bucket_name     = optional(string, null)
    bucket_location = optional(string, null)
    job_name        = optional(string, "artifacts-job")
    scheduler_cron  = optional(string, "*/15 * * * *")
  })
  default = {
    enabled = false
  }
}

variable "commit_review_status" {
  description = "The configuration block for commit review status"
  type = object({
    enabled  = bool
    table_id = optional(string, null)
    table_iam = optional(object({
      owners  = optional(list(string), [])
      editors = optional(list(string), [])
      viewers = optional(list(string), [])
    }), null)
    job_name       = optional(string, "commit-review-status-job")
    scheduler_cron = optional(string, "0 */4 * * *")
  })
  default = {
    enabled = false
  }
}

variable "invocation_comment" {
  description = "The configuration block for invocation comment"
  type = object({
    enabled  = bool
    table_id = optional(string, null)
    table_iam = optional(object({
      owners  = optional(list(string), [])
      editors = optional(list(string), [])
      viewers = optional(list(string), [])
    }), null)
  })
  default = {
    enabled = false
  }
}

variable "github_metrics_dashboard" {
  description = "The configuration for the GitHub Metrics dashboard"
  type = object({
    enabled          = bool
    looker_report_id = optional(string, "de3a9011-f38b-4d9a-a48e-23fe58186589") # abcxyz-provided GitHub Metrics report template
    viewers          = optional(list(string), [])
  })
  default = {
    enabled = false
    viewers = []
  }
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
