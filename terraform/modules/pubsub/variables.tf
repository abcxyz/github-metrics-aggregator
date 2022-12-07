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

variable "topic_iam" {
  description = "IAM bindings in {ROLE => [MEMBERS]} format for the PubSub topic."
  type        = map(list(string))
  default     = {}
}

variable "bigquery_destination" {
  description = "BigQuery subscription destination (e.g. project:dataset.table)"
  type        = string
}
