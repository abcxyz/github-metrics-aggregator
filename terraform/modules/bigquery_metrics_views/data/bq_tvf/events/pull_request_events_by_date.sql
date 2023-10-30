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

-- TVF queries cannot end in ; as it will be wrapped in a CREATE TABLE FUNCTION
-- and the end of the query is not the end of the compiled SQL statement.

SELECT
  received,
  event,
  JSON_VALUE(payload, "$.action") action,
  organization,
  organization_id,
  repository_full_name,
  repository_id,
  repository,
  repository_visibility,
  sender,
  sender_id,
  LAX_STRING(payload.pull_request.active_lock_reason) active_lock_reason,
  SAFE.INT64(payload.pull_request.additions) additions,
  LAX_STRING(payload.pull_request.base.ref) base_ref,
  SAFE.INT64(payload.pull_request.changed_files) changed_files,
  TIMESTAMP(LAX_STRING(payload.pull_request.closed_at)) closed_at,
  SAFE.INT64(payload.pull_request.comments) comments,
  SAFE.INT64(payload.pull_request.commits) commits,
  TIMESTAMP(LAX_STRING(payload.pull_request.created_at)) created_at,
  SAFE.INT64(payload.pull_request.deletions) deletions,
  SAFE.BOOL(payload.pull_request.draft) draft,
  LAX_STRING(payload.pull_request.head.ref) head_ref,
  LAX_STRING(payload.pull_request.html_url) html_url,
  SAFE.INT64(payload.pull_request.id) id,
  SAFE.BOOL(payload.pull_request.locked) locked,
  SAFE.BOOL(payload.pull_request.maintainer_can_modify) maintainer_can_modify,
  LAX_STRING(payload.pull_request.merge_commit_sha) merge_commit_sha,
  LAX_STRING(payload.pull_request.mergeable_state) mergeable_state,
  SAFE.BOOL(payload.pull_request.merged) merged,
  TIMESTAMP(LAX_STRING(payload.pull_request.merged_at)) merged_at,
  LAX_STRING(payload.pull_request.merged_by.login) merged_by,
  SAFE.INT64(payload.pull_request.number) number,
  SAFE.INT64(payload.pull_request.review_comments) review_comments,
  LAX_STRING(payload.pull_request.state) state,
  LAX_STRING(payload.pull_request.title) title,
  TIMESTAMP(LAX_STRING(payload.pull_request.updated_at)) updated_at,
  SAFE.INT64(payload.pull_request.user.id) author_id,
  LAX_STRING(payload.pull_request.user.login) author,
  TIMESTAMP_DIFF( TIMESTAMP(LAX_STRING(payload.pull_request.closed_at)), TIMESTAMP(LAX_STRING(payload.pull_request.created_at)), SECOND) open_duration_seconds
FROM
  `${parent_project_id}.${parent_dataset_id}.${parent_routine_id}`(startTimestamp, endTimestamp, "pull_request")
