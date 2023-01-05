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

module "bigquery" {
  source     = "../bigquery"
  project_id = var.project_id
  dataset_id = "github_webhook"
  table_id   = "events"
}

module "pubsub" {
  source               = "../pubsub"
  project_id           = var.project_id
  name                 = var.name
  bigquery_destination = module.bigquery.bigquery_destination
  topic_iam = {
    "roles/pubsub.publisher" : [
      module.cloud_run.service_account_iam_email
    ]
  }
}

module "cloud_run" {
  source     = "../cloud_run"
  project_id = var.project_id
  region     = var.region
  name       = var.name
  ingress    = "internal-and-cloud-load-balancing"
  domain     = var.domain
}
