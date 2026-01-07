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

resource "google_bigquery_table" "events_dashboard_mv" {
  project     = var.project_id

  dataset_id  = google_bigquery_dataset.default.dataset_id
  table_id    = "events_dashboard_mv"
  description = "Materialized view for CL Stats Dashboard optimization"

  time_partitioning {
    type  = "DAY"
    field = "updated_at"
  }

  clustering = [
    "enterprise_name",
    "organization_login",
    "repository_full_name",
    "sender_login",
    "event"
  ]

  materialized_view {
    query               = <<EOF
SELECT
  delivery_id,
  TIMESTAMP(JSON_VALUE(payload, '$.pull_request.updated_at')) as updated_at,
  TIMESTAMP(JSON_VALUE(payload, '$.pull_request.created_at')) as pr_created_at,
  TIMESTAMP(JSON_VALUE(payload, '$.pull_request.closed_at')) as pr_closed_at,
  received,
  event,
  
  JSON_VALUE(payload, '$.repository.full_name') as repository_full_name,
  SAFE_CAST(JSON_VALUE(payload, '$.repository.id') as INT64) as repository_id,
  JSON_VALUE(payload, '$.sender.login') as sender_login,
  JSON_VALUE(payload, '$.organization.login') as organization_login,
  
  JSON_VALUE(payload, '$.enterprise.name') as enterprise_name,
  JSON_VALUE(payload, '$.pull_request.user.login') as pr_author,
  JSON_VALUE(payload, '$.pull_request.html_url') as pull_request_url,
  JSON_VALUE(payload, '$.pull_request.title') as pull_request_title,
  JSON_VALUE(payload, '$.pull_request.number') as pull_request_number,
  JSON_VALUE(payload, '$.pull_request.state') as pull_request_state,
  JSON_VALUE(payload, '$.pull_request.merged') as is_merged,
  SAFE_CAST(JSON_VALUE(payload, '$.pull_request.additions') as INT64) as pr_additions,
  SAFE_CAST(JSON_VALUE(payload, '$.pull_request.deletions') as INT64) as pr_deletions,
  SAFE_CAST(JSON_VALUE(payload, '$.pull_request.changed_files') as INT64) as pr_changed_files,
  
  TIMESTAMP_DIFF(
    TIMESTAMP(JSON_VALUE(payload, '$.pull_request.closed_at')), 
    TIMESTAMP(JSON_VALUE(payload, '$.pull_request.created_at')), 
    SECOND
  ) as open_duration_seconds,
  
  (SAFE_CAST(JSON_VALUE(payload, '$.pull_request.additions') as INT64) + 
   SAFE_CAST(JSON_VALUE(payload, '$.pull_request.deletions') as INT64)) as pr_lines_changed,
   
  CASE
    WHEN (SAFE_CAST(JSON_VALUE(payload, '$.pull_request.additions') as INT64) + SAFE_CAST(JSON_VALUE(payload, '$.pull_request.deletions') as INT64)) IS NULL THEN 'U'
    WHEN (SAFE_CAST(JSON_VALUE(payload, '$.pull_request.additions') as INT64) + SAFE_CAST(JSON_VALUE(payload, '$.pull_request.deletions') as INT64)) < 9 THEN 'XS'
    WHEN (SAFE_CAST(JSON_VALUE(payload, '$.pull_request.additions') as INT64) + SAFE_CAST(JSON_VALUE(payload, '$.pull_request.deletions') as INT64)) < 49 THEN 'S'
    WHEN (SAFE_CAST(JSON_VALUE(payload, '$.pull_request.additions') as INT64) + SAFE_CAST(JSON_VALUE(payload, '$.pull_request.deletions') as INT64)) < 249 THEN 'M'
    WHEN (SAFE_CAST(JSON_VALUE(payload, '$.pull_request.additions') as INT64) + SAFE_CAST(JSON_VALUE(payload, '$.pull_request.deletions') as INT64)) < 999 THEN 'L'
    ELSE 'XL'
  END as tshirt_size,

  JSON_VALUE(payload, '$.review.user.login') as reviewer_login,
  JSON_VALUE(payload, '$.review.state') as review_state,
  JSON_VALUE(payload, '$.review.pull_request.html_url') as review_pr_url,
  JSON_VALUE(payload, '$.action') as action,
  JSON_VALUE(payload, '$.issue.pull_request') as issue_is_pr,
  JSON_VALUE(payload, '$.comment.id') as comment_id,
  JSON_VALUE(payload, '$.thread.id') as thread_id
FROM `${var.project_id}.${google_bigquery_dataset.default.dataset_id}.${google_bigquery_table.raw_events_table.table_id}`
EOF
    enable_refresh      = true
    refresh_interval_ms = 1800000 # 30 minutes
  }

  deletion_protection = false # MVs can be recreated easily
}
