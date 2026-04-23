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

output "gclb_external_ip_name" {
  description = "The external IPv4 name assigned to the global fowarding rule for the global load balancer fronting the webhook."
  value       = try(module.gclb[0].external_ip_name, null)
}

output "gclb_external_ip_address" {
  description = "The external IPv4 assigned to the global fowarding rule for the global load balancer fronting the webhook."
  value       = try(module.gclb[0].external_ip_address, null)
}

output "webhook_run_service" {
  description = "The Cloud Run webhook service data."
  value = {
    service_id             = module.webhook_cloud_run.service_id
    service_url            = module.webhook_cloud_run.url
    service_name           = module.webhook_cloud_run.service_name
    service_account_name   = google_service_account.webhook.name
    service_account_email  = google_service_account.webhook.email
    service_account_member = "serviceAccount:${google_service_account.webhook.email}"
  }
}
