variable "project_id" {
  type        = string
  description = "The GCP project ID."
}

variable "name" {
  description = "The name of this component."
  type        = string
  validation {
    condition     = can(regex("^[A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]$", var.name))
    error_message = "Name can only contain letters, numbers, hyphens(-) and must start with letter."
  }
}

variable "dataset_id" {
  type        = string
  description = "The BigQuery dataset id."
}

variable "dataset_location" {
  type        = string
  description = "The BigQuery dataset location."
  default     = "US"
}

variable "dataset_iam" {
  description = "IAM bindings in {ROLE => [MEMBERS]} format for the BigQuery github_webhook dataset."
  type        = map(list(string))
  default     = {}
}

variable "table_id" {
  type        = string
  description = "The BigQuery table id."
}

variable "table_iam" {
  description = "IAM bindings in {ROLE => [MEMBERS]} format for the BigQuery github_webhook.events table."
  type        = map(list(string))
  default     = {}
}

variable "create_default_views" {
  description = "Create the default curated set of views"
  type        = bool
  default     = true
}

variable "views" {
  description = "A list of custom views to create"
  default     = {}
  type = map(object({
    template_path  = string,
    query          = string,
    use_legacy_sql = bool,
    labels         = map(string),
  }))
}
