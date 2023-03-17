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

output "gclb_external_ip_name" {
  description = "The external IPv4 name assigned to the global fowarding rule for the global load balancer."
  value       = module.webhook.gclb_external_ip_name
}

output "gclb_external_ip_address" {
  description = "The external IPv4 assigned to the global fowarding rule for the global load balancer."
  value       = module.webhook.gclb_external_ip_address
}

output "run_service_url" {
  description = "The Cloud Run service url."
  value       = module.webhook.run_service_url
}

output "run_service_id" {
  description = "The Cloud Run service id."
  value       = module.webhook.run_service_id
}

output "run_service_name" {
  description = "The Cloud Run service name."
  value       = module.webhook.run_service_name
}

output "run_service_account_name" {
  description = "Cloud Run service account name."
  value       = module.webhook.run_service_account_name
}

output "run_service_account_email" {
  description = "Cloud Run service account email."
  value       = module.webhook.run_service_account_email
}

output "run_service_account_member" {
  description = "Cloud Run service account email iam string."
  value       = module.webhook.run_service_account_member
}

output "bigquery_dataset_id" {
  description = "BigQuery dataset resource."
  value       = module.webhook.bigquery_dataset_id
}

output "bigquery_table_id" {
  description = "BigQuery table resource."
  value       = module.webhook.bigquery_table_id
}

output "bigquery_pubsub_destination" {
  description = "BigQuery PubSub destination"
  value       = module.webhook.bigquery_pubsub_destination
}


