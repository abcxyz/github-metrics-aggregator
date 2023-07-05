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
-- Filters events to those just pertaining to issue comments.
-- Relevant GitHub Docs:
-- https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#issue_comment

SELECT
  received,
  event,
  JSON_VALUE(payload, "$.action") action,
  JSON_VALUE(payload, "$.organization.login") organization,
  SAFE_CAST(JSON_VALUE(payload, "$.organization.id") AS INT64) organization_id,
  JSON_VALUE(payload, "$.repository.full_name") repository_full_name,
  SAFE_CAST(JSON_VALUE(payload, "$.repository.id") AS INT64) repository_id,
  JSON_VALUE(payload, "$.repository.name") repository,
  JSON_VALUE(payload, "$.repository.visibility") repository_visibility,
  JSON_VALUE(payload, "$.sender.login") sender,
  SAFE_CAST(JSON_VALUE(payload, "$.sender.id") AS INT64) sender_id,
  JSON_VALUE(payload, "$.comment.body") body,
  SAFE_CAST(JSON_VALUE(payload, "$.comment.user.id") AS INT64) commenter_id,
  JSON_VALUE(payload, "$.comment.user.login") commenter,
  TIMESTAMP(JSON_VALUE(payload, "$.comment.created_at")) created_at,
  JSON_VALUE(payload, "$.comment.html_url") html_url,
  SAFE_CAST(JSON_VALUE(payload, "$.comment.id") AS INT64) id,
  SAFE_CAST(JSON_VALUE(payload, "$.issue.id") AS INT64) issue_id,
  SAFE_CAST(JSON_VALUE(payload, "$.comment.line") AS INT64) line,
  JSON_VALUE(payload, "$.comment.path") path,
  TIMESTAMP(JSON_VALUE(payload, "$.comment.updated_at")) updated_at,
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "issue_comment";