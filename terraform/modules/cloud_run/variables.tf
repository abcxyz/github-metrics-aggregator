variable "project_id" {
  type        = string
  description = "The GCP project ID."
}

variable "region" {
  type        = string
  description = "The GCP region."
}

variable "name" {
  description = "The name of this component."
  type        = string
  validation {
    condition     = can(regex("^[A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]$", var.name))
    error_message = "Name can only contain letters, numbers, hyphens(-) and must start with letter."
  }
}

variable "ingress" {
  type        = string
  description = "The Cloud Run ingress setting (e.g. all, internal, internal-and-cloud-load-balancing)."
  default     = "internal-and-cloud-load-balancing"
}

variable "ssl" {
  type        = bool
  description = "Enable SSL on the global load balancer in front of the Cloud Run service."
  default     = true
}

variable "domain" {
  type        = string
  description = "The managed SSL domain for the load balancer."
  default     = ""
}
