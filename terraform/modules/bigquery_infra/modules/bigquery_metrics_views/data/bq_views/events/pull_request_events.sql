-- Copyright 2023 The Authors (see AUTHORS file)
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

SELECT
  received,
  event,
  delivery_id,
  JSON_VALUE(payload, "$.action") action,
  organization,
  organization_id,
  repository_full_name,
  repository_id,
  repository,
  repository_visibility,
  sender,
  sender_id,
  JSON_VALUE(payload, "$.pull_request.active_lock_reason") active_lock_reason,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.additions") AS INT64) additions,
  JSON_VALUE(payload, "$.pull_request.base.ref") base_ref,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.changed_files") AS INT64) changed_files,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.closed_at")) closed_at,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.comments") AS INT64) comments,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.commits") AS INT64) commits,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.created_at")) created_at,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.deletions") AS INT64) deletions,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.draft") AS BOOL) draft,
  JSON_VALUE(payload, "$.pull_request.head.ref") head_ref,
  JSON_VALUE(payload, "$.pull_request.html_url") html_url,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.id") AS INT64) id,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.locked") AS BOOL) locked,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.maintainer_can_modify") AS BOOL) maintainer_can_modify,
  JSON_VALUE(payload, "$.pull_request.merge_commit_sha") merge_commit_sha,
  JSON_VALUE(payload, "$.pull_request.mergeable_state") mergeable_state,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.merged") AS BOOL) merged,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.merged_at")) merged_at,
  JSON_VALUE(payload, "$.pull_request.merged_by.login") merged_by,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.number") AS INT64) number,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.review_comments") AS INT64) review_comments,
  JSON_VALUE(payload, "$.pull_request.state") state,
  JSON_VALUE(payload, "$.pull_request.title") title,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.updated_at")) updated_at,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.user.id") AS INT64) author_id,
  JSON_VALUE(payload, "$.pull_request.user.login") author,
  TIMESTAMP_DIFF( TIMESTAMP(JSON_VALUE(payload, "$.pull_request.closed_at")), TIMESTAMP(JSON_VALUE(payload, "$.pull_request.created_at")), SECOND) open_duration_seconds
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "pull_request";
