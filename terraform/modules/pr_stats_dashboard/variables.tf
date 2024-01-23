variable "project_id" {
  description = "The GCP project ID."
  type        = string
}

variable "dataset_id" {
  type        = string
  description = "The BigQuery dataset id."
}

variable "looker_report_id" {
  type = string
  description = "The template Looker Studio Report ID."
}

variable "viewers" {
  type = list(string)
  description = "The list of members who can view the PR Stats Looker Studio Dashboard."
  default = []
}
