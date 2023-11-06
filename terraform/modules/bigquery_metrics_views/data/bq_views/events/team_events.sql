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
-- Filters events to those just pertaining to teams.
-- Relevant GitHub Docs:
-- https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#team

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
  SAFE_CAST(JSON_VALUE(payload, "$.team.deleted") AS BOOL) deleted,
  JSON_VALUE(payload, "$.team.description") description,
  JSON_VALUE(payload, "$.team.html_url") html_url,
  SAFE_CAST(JSON_VALUE(payload, "$.team.id") AS INT64) id,
  JSON_VALUE(payload, "$.team.members_url") members_url,
  JSON_VALUE(payload, "$.team.name") name,
  SAFE_CAST(JSON_VALUE(payload, "$.team.parent.id") AS INT64) parent_id,
  JSON_VALUE(payload, "$.team.parent.name") parent_name,
  JSON_VALUE(payload, "$.team.permission") permission,
  JSON_VALUE(payload, "$.team.privacy") privacy,
  JSON_VALUE(payload, "$.team.notification_setting") notification_setting,
  JSON_VALUE(payload, "$.team.slug") slug,
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "team";
