variable "project_id" {
  description = "The project ID."
  type        = string
}

variable "project_number" {
  description = "The project number."
  type        = string
}

variable "dataset_id" {
  description = "The dataset ID."
  type        = string
}

variable "location" {
  description = "The location of the dataset."
  type        = string
  default     = "US"
}

variable "prstats_pull_requests_table_name" {
  description = "The name of the PRStats pull requests table."
  type        = string
  default     = "gma_prstats_pull_requests"
}

variable "prstats_pull_request_reviews_table_name" {
  description = "The name of the PRStats pull request reviews table."
  type        = string
  default     = "gma_prstats_pull_request_reviews"
}

variable "prstats_source_table_name" {
  description = "The name of the source table for PRStats."
  type        = string
  default     = "events"
}

variable "prstats_pull_requests_schedule" {
  description = "The schedule for the gma_prstats_pull_requests query."
  type        = string
  default     = "every 30 mins"
}

variable "prstats_pull_request_reviews_schedule" {
  description = "The schedule for the gma_prstats_pull_request_reviews query."
  type        = string
  default     = "every 30 mins"
}

variable "prstats_service_account_email" {
  description = "The service account email for PRStats."
  type        = string
}
