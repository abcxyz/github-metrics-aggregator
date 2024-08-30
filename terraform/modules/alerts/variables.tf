# Copyright 2024 The Authors (see AUTHORS file)
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

variable "alert_notification_channel_non_paging" {
  description = "Non-paging notification channels"
  type        = map(any)
  default = {
    email = {
      labels = {
        email_address = ""
      }
    }
  }
}

variable "forward_progress_job_indicators" {
  description = "Map of overrides for forward progress Cloud Run Job indicators. These are merged with the default variables. The window must be in seconds."
  type = map(object({
    metric = string
    window = number
  }))
  default = {}
}
