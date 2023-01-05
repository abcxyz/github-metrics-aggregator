-- Copyright 2023 The Authors (see AUTHORS file)
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--      http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

SELECT
  received,
  event,
  JSON_VALUE(payload, "$.organization.login") owner,
  JSON_VALUE(payload, "$.repository.name") repo,
  JSON_VALUE(payload, "$.repository.full_name") repo_full_name,
  JSON_VALUE(payload, "$.sender.login") sender,
  JSON_VALUE(payload, "$.action") action,
  JSON_VALUE(payload, "$.issue.title") title,
  JSON_VALUE(payload, "$.issue.comments") comments,
  JSON_VALUE(payload, "$.issue.state") state,
  JSON_VALUE(payload, "$.issue.state_reason") state_reason,
  JSON_VALUE(payload, "$.issue.html_url") html_url,
  TIMESTAMP(JSON_VALUE(payload, "$.issue.created_at")) created_at,
  TIMESTAMP(JSON_VALUE(payload, "$.issue.closed_at")) closed_at,
  TIMESTAMP_DIFF(TIMESTAMP(JSON_VALUE(payload, "$.issue.closed_at")), TIMESTAMP(JSON_VALUE(payload, "$.issue.created_at")), SECOND) open_duration_s,
  PARSE_JSON(payload) payload
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "issues";
