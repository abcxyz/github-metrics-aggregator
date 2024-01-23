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

# add all groups who need to view through lookerstudio to jobUser role
resource "google_project_iam_member" "pr_stats_dashboard_job_users" {
  for_each = toset(var.viewers)

  project = var.project_id

  role   = "roles/bigquery.jobUser"
  member = each.value
}

# grant users access to the dataset used for displaying PR stats
resource "google_bigquery_dataset_iam_member" "pr_stats_dashboard_data_viewers" {
  for_each = toset(var.viewers)

  project = var.project_id

  dataset_id = var.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}
