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

resource "google_project" "dev" {
  name       = "github-metrics-dev"
  project_id = "github-metrics-dev"
  # folder id for "tycho.joonix.net > github-metrics-envs"
  folder_id = "folders/758171742657"

  billing_account = "016242-61A3FB-F92462"
}

module "e2e" {
  source     = "../modules/e2e"
  project_id = google_project.dev.project_id
  region     = "us-central1"
  name       = "github-webhook"
  domain     = "github-webhook-dev.tycho.joonix.net"
}
