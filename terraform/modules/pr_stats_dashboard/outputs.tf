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

output "looker_studio_report_link" {
  description = "The Looker Studio Linking API link for connecting the data sources for the PR Stats dashboard."
  value = join("",
    [
      "https://lookerstudio.google.com/reporting/create",
      "?c.reportId=${var.looker_report_id}",
      "&r.reportName=PR%20Stats",
      "&ds.ds0.keepDatasourceName",
      "&ds.ds0.connector=bigQuery",
      "&ds.ds0.refreshFields",
      "&ds.ds0.projectId=${var.project_id}",
      "&ds.ds0.type=TABLE",
      "&ds.ds0.datasetId=${var.dataset_id}",
      "&ds.ds0.tableId=pull_requests",
      "&ds.ds2.keepDatasourceName",
      "&ds.ds2.connector=bigQuery",
      "&ds.ds2.refreshFields",
      "&ds.ds2.projectId=${var.project_id}",
      "&ds.ds2.type=TABLE",
      "&ds.ds2.datasetId=${var.dataset_id}",
      "&ds.ds2.tableId=pull_request_reviews",
    ]
  )
}
