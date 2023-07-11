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
-- Filters events to those just pertaining to pull request reviews.
-- Relevant GitHub Docs:
-- https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request_review

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
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.id") AS INT64) pull_request_id,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.number") AS INT64) pull_request_number,
  JSON_VALUE(payload, "$.pull_request.state") pull_request_state,
  JSON_VALUE(payload, "$.pull_request.url") pull_request_url,
  SAFE_CAST(JSON_QUERY(payload, "$.pull_request.user.id") AS INT64) pull_request_author_id,
  JSON_VALUE(payload, "$.pull_request.user.login") pull_request_author,
  SAFE_CAST(JSON_QUERY(payload, "$.review.id") AS INT64) id,
  JSON_VALUE(payload, "$.review.body") body,
  JSON_VALUE(payload, "$.review.html_url") html_url,
  JSON_VALUE(payload, "$.review.state") state,
  TIMESTAMP(JSON_VALUE(payload, "$.review.submitted_at")) submitted_at,
  JSON_VALUE(payload, "$.review.user.login") reviewer,
  SAFE_CAST(JSON_QUERY(payload, "$.review.user.id") AS INT64) reviewer_id,
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "pull_request_review";
