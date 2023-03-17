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

data "google_project" "default" {
  project_id = var.project_id
}

resource "google_project_service" "default" {
  for_each = toset([
    "cloudresourcemanager.googleapis.com",
    "bigquery.googleapis.com",
    "pubsub.googleapis.com",
  ])
  
  project = var.project_id

  service            = each.value
  disable_on_destroy = false
}

module "gclb" {
  source           = "git::https://github.com/abcxyz/terraform-modules.git//modules/gclb_cloud_run_backend?ref=1d5d7f3f166679b02cd3f1ec615d287d6b7002dc"
  project_id       = data.google_project.default.project_id
  name             = var.name
  run_service_name = module.cloud_run.service_name
  domain           = var.domain
}

resource "google_service_account" "run_service_account" {
  project      = data.google_project.default.project_id
  account_id   = "${var.name}-sa"
  display_name = "${var.name}-sa Cloud Run Service Account"
}

module "cloud_run" {
  source                = "git::https://github.com/abcxyz/terraform-modules.git//modules/cloud_run?ref=1d5d7f3f166679b02cd3f1ec615d287d6b7002dc"
  project_id            = data.google_project.default.project_id
  name                  = var.name
  image                 = var.image
  ingress               = "internal-and-cloud-load-balancing"
  secrets               = ["github-webhook-secret"]
  service_account_email = google_service_account.run_service_account.email
  service_iam           = var.service_iam
  envvars = {
    "PROJECT_ID" : data.google_project.default.project_id,
    "TOPIC_ID" : google_pubsub_topic.default.name
  }
  secret_envvars = {
    "WEBHOOK_SECRET" : {
      name : "github-webhook-secret",
      version : "latest",
    }
  }
}
