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

-- Filters events to those just pertaining to releases.
-- Relevant GitHub Docs:
-- https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#release

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
  JSON_VALUE(payload, "$.release.author.login") author,
  SAFE_CAST(JSON_VALUE(payload, "$.release.author.id") AS INT64) author_id,
  JSON_QUERY(payload, "$.release.body") body,
  TIMESTAMP(JSON_VALUE(payload, "$.release.created_at")) created_at,
  SAFE_CAST(JSON_VALUE(payload, "$.release.draft") AS BOOL) draft,
  JSON_VALUE(payload, "$.release.html_url") html_url,
  SAFE_CAST(JSON_VALUE(payload, "$.release.id") AS INT64) id,
  JSON_VALUE(payload, "$.release.name") name,
  SAFE_CAST(JSON_VALUE(payload, "$.release.prerelease") AS BOOL) prerelease,
  TIMESTAMP(JSON_VALUE(payload, "$.release.published_at")) published_at,
  JSON_VALUE(payload, "$.release.tag_name") tag_name,
  JSON_VALUE(payload, "$.release.tarball_url") tarball_url,
  JSON_VALUE(payload, "$.release.target_commitish") target_commitish,
  JSON_VALUE(payload, "$.release.upload_url") upload_url,
  JSON_VALUE(payload, "$.release.zipball_url") zipball_url,
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "release";
