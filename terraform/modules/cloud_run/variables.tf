/**
 * Copyright 2023 The Authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
