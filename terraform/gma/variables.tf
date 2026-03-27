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
  default = {}
}

variable "retry_service_iam" {
  description = "IAM member bindings for the retry Cloud Run services."
  type = object({
    admins     = optional(list(string), [])
    developers = optional(list(string), [])
    invokers   = optional(list(string), [])
  })
  default = {}
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
  default = {}
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
  default = {}
}

variable "dead_letter_sub_iam" {
  description = "IAM member binding for the PubSub dead letter subscription."
  type = object({
    admins      = optional(list(string), [])
    editors     = optional(list(string), [])
    viewers     = optional(list(string), [])
    subscribers = optional(list(string), [])
  })
  default = {}
}

variable "compute_service_account_email" {
  description = "The email of an existing service account to use for the GMA compute services. If left blank, one will be created."
  type        = string
  default     = ""
}


variable "dataset_id" {
  type        = string
  description = "The BigQuery dataset id to create."
  default     = "github_metrics"
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


variable "checkpoint_table_id" {
  description = "The BigQuery checkpoint table id to create."
  type        = string
  default     = "checkpoint"
}


variable "failure_events_table_id" {
  description = "The BigQuery failure events table id to create."
  type        = string
  default     = "failure_events"
}




variable "bigquery_project_id" {
  description = "The project ID where the BigQuery instance exists."
  type        = string
}

variable "event_delivery_retry_limit" {
  description = "Number of attempts to delivery a failed event from GitHub."
  type        = string
  default     = "10"
}


variable "github_app_id" {
  description = "The GitHub App ID."
  type        = string
}

variable "region" {
  description = "The default Google Cloud region to deploy resources in (defaults to 'us-central1')."
  type        = string
  default     = "us-central1"
}


variable "artifacts" {
  description = "The configuration block for artifacts table"
  type = object({
    enabled  = bool
    table_id = optional(string, null)
    table_iam = optional(object({
      owners  = optional(list(string), [])
      editors = optional(list(string), [])
      viewers = optional(list(string), [])
      })
      default = {}
    bucket_name     = optional(string, null)
    bucket_location = optional(string, null)
    job_name        = optional(string, "artifacts-job")
    job_iam = optional(object({
      admins     = optional(list(string), [])
      developers = optional(list(string), [])
      invokers   = optional(list(string), [])
      })
      default = {}
    job_additional_env_vars = optional(map(string), {})
    scheduler_cron          = optional(string, "*/15 * * * *")
    alerts = optional(object({
      enabled = bool
      built_in_forward_progress_indicators = optional(map(object({
        metric = string
        window = number
      })))
      built_in_container_util_indicators = optional(map(object({
        metric    = string
        window    = number
        threshold = number
        p_value   = number
      })))
    }))
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
      })
      default = {}
    job_name = optional(string, "commit-review-status-job")
    job_iam = optional(object({
      admins     = optional(list(string), [])
      developers = optional(list(string), [])
      invokers   = optional(list(string), [])
      })
      default = {}
    job_additional_env_vars = optional(map(string), {})
    scheduler_cron          = optional(string, "0 */4 * * *")
    alerts = optional(object({
      enabled = bool
      built_in_forward_progress_indicators = optional(map(object({
        metric = string
        window = number
      })))
      built_in_container_util_indicators = optional(map(object({
        metric    = string
        window    = number
        threshold = number
        p_value   = number
      })))
    }))
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

variable "github_enterprise_server_url" {
  description = "The GitHub Enterprise server URL if available, format \"https://[hostname]\"."
  type        = string
  default     = ""
}



variable "default_log_bucket_configuration" {
  description = "The configuration for the _Default log bucket"
  type = object({
    retention_period = number
    location         = string
    enable_analytics = bool
  })
  default = {
    retention_period = 30
    location         = "global"
    enable_analytics = false
  }
}

variable "bigquery_event_views_override" {
  description = "BigQuery event view resources. To be used when the BigQuery infrastructure module is disabled."
  type        = map(string)
  default     = {}
}

variable "bigquery_resource_views_override" {
  description = "BigQuery resource view resources. To be used when the BigQuery infrastructure module is disabled."
  type        = map(string)
  default     = {}
}

variable "webhook_max_instances" {
  type        = string
  default     = "10"
  description = "The max number of instances for the Webhook Cloud Run service (defaults to '10')."

}

variable "retry_job_schedule" {
  type        = string
  default     = "0 * * * *"
  description = "Frequencey to run the retry job. Follows standard cron syntax. Defaults to every hour."
}

variable "retry_job_timeout" {
  description = "The task timeout setting see: https://cloud.google.com/run/docs/configuring/task-timeout#set_task_timeout. Defaults to 45 minutes"
  type        = string
  default     = "2700s"
}

# This variable controls which secrets are CREATED by the retry module
variable "secrets_to_create" {
  description = "A list of secret IDs to create in Secret Manager."
  type        = set(string)
  default     = ["github-private-key"]
}

variable "enable_relay_service" {
  description = "Enable the relay service. Defaults to false."
  type        = bool
  default     = false
}

variable "relay_service_iam" {
  description = "IAM member bindings for the relay Cloud Run service."
  type = object({
    admins     = optional(list(string), [])
    developers = optional(list(string), [])
    invokers   = optional(list(string), [])
  })
  default = {}
}

variable "relay_topic_id" {
  description = "The PubSub topic ID for the relay service to publish to."
  type        = string
  default     = ""
}

variable "relay_project_id" {
  description = "The project ID where the relay PubSub topic exists."
  type        = string
  default     = ""
}

variable "optimized_events_table_id" {
  description = "The ID of the BigQuery table for optimized events."
  type        = string
  default     = "optimized_events"
}



variable "bigquery_infra_deploy" {
  description = "Enable the deployment of the BigQuery infrastructure module."
  type        = bool
  default     = true
}
