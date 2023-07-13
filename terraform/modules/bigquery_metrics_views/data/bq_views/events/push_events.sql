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

-- Filters events to those just pertaining to pushes.
-- Relevant GitHub Docs:
-- https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#push

SELECT
  received,
  event,
  organization,
  organization_id,
  repository_full_name,
  repository_id,
  repository,
  repository_visibility,
  sender,
  sender_id,
  JSON_VALUE(payload, "$.after") after_sha,
  JSON_VALUE(payload, "$.before") before_sha,
  JSON_VALUE(payload, "$.compare") compare_url,
  SAFE_CAST(JSON_VALUE(payload, "$.created") AS BOOL) created,
  SAFE_CAST(JSON_VALUE(payload, "$.deleted") AS BOOL) deleted,
  SAFE_CAST(JSON_VALUE(payload, "$.forced") AS BOOL) forced,
  JSON_VALUE(payload, "$.pusher.name") pusher,
  JSON_VALUE(payload, "$.ref") ref,
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = 'push'
