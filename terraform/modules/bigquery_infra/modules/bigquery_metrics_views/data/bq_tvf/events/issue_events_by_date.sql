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
  delivery_id,
  LAX_STRING(payload.action) action,
  ORGANIZATION,
  organization_id,
  repository_full_name,
  repository_id,
  repository,
  repository_visibility,
  sender,
  sender_id,
  LAX_STRING(payload.issue.active_lock_reason) active_lock_reason,
  LAX_STRING(payload.issue.assignee.login) assignee,
  LAX_STRING(payload.issue.user.login) author,
  SAFE.INT64(payload.issue.user.id) author_id,
  LAX_STRING(payload.issue.author_association) author_association,
  LAX_STRING(payload.issue.body) body,
  TIMESTAMP(LAX_STRING(payload.issue.closed_at)) closed_at,
  SAFE.INT64(payload.issue.comments) comments,
  TIMESTAMP(LAX_STRING(payload.issue.created_at)) created_at,
  SAFE.BOOL(payload.issue.draft) draft,
  LAX_STRING(payload.issue.html_url) html_url,
  SAFE.INT64(payload.issue.id) id,
  SAFE.BOOL(payload.issue.locked) locked,
  SAFE.INT64(payload.issue.number) number,
  TIMESTAMP_DIFF(TIMESTAMP(LAX_STRING(payload.issue.closed_at)), TIMESTAMP(LAX_STRING(payload.issue.created_at)), SECOND) open_duration_seconds,
  LAX_STRING(payload.issue.state) state,
  LAX_STRING(payload.issue.state_reason) state_reason,
  LAX_STRING(payload.issue.title) title,
  TIMESTAMP(LAX_STRING(payload.issue.updated_at)) updated_at,
FROM
  `${parent_project_id}.${parent_dataset_id}.${parent_routine_id}`(startTimestamp, endTimestamp, 'issues')
