# Copyright 2026 The Authors (see AUTHORS file)
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

locals {
  # time helpers
  second = 1
  minute = 60 * local.second
  hour   = 60 * local.minute
  day    = 24 * local.hour
}

resource "google_service_account" "backend" {
  project = var.project_id

  account_id = "${var.prefix_name}-backend"

  display_name = "GMA Backend Service Account"
}

locals {
  compute_service_account_email  = google_service_account.backend.email
  compute_service_account_member = "serviceAccount:${google_service_account.backend.email}"
}
