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
--
-- Filters events to those just pertaining to issues.
-- Relevant GitHub Docs:
-- https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#issues

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
  JSON_VALUE(payload, "$.issue.active_lock_reason") active_lock_reason,
  JSON_VALUE(payload, "$.issue.assignee.login") assignee,
  JSON_VALUE(payload, "$.issue.user.login") author,
  SAFE_CAST(JSON_VALUE(payload, "$.issue.user.id") AS INT64) author_id,
  JSON_VALUE(payload, "$.issue.author_association") author_association,
  JSON_VALUE(payload, "$.issue.body") body,
  TIMESTAMP(JSON_VALUE(payload, "$.issue.closed_at")) closed_at,
  SAFE_CAST(JSON_VALUE(payload, "$.issue.comments") AS INT64) comments,
  TIMESTAMP(JSON_VALUE(payload, "$.issue.created_at")) created_at,
  SAFE_CAST(JSON_VALUE(payload, "$.issue.draft") AS BOOL) draft,
  JSON_VALUE(payload, "$.issue.html_url") html_url,
  SAFE_CAST(JSON_VALUE(payload, "$.issue.id") AS INT64) id,
  SAFE_CAST(JSON_VALUE(payload, "$.issue.locked") AS BOOL) locked,
  SAFE_CAST(JSON_VALUE(payload, "$.issue.number") AS INT64) number,
  TIMESTAMP_DIFF(TIMESTAMP(JSON_VALUE(payload, "$.issue.closed_at")), TIMESTAMP(JSON_VALUE(payload, "$.issue.created_at")), SECOND) open_duration_seconds,
  JSON_VALUE(payload, "$.issue.state") state,
  JSON_VALUE(payload, "$.issue.state_reason") state_reason,
  JSON_VALUE(payload, "$.issue.title") title,
  TIMESTAMP(JSON_VALUE(payload, "$.issue.updated_at")) updated_at,
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "issues";
