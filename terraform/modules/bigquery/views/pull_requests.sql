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
  JSON_VALUE(payload, "$.pull_request.title") title,
  JSON_VALUE(payload, "$.pull_request.state") state,
  JSON_VALUE(payload, "$.pull_request.html_url") html_url,
  JSON_VALUE(payload, "$.pull_request.base.ref") base_ref,
  JSON_VALUE(payload, "$.pull_request.head.ref") head_ref,
  JSON_VALUE(payload, "$.pull_request.user.login") submitter,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.created_at")) created_at,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.closed_at")) closed_at,
  JSON_VALUE(payload, "$.pull_request.merged") merged,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.merged_at")) merged_at,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.merged_by")) merged_by,
  TIMESTAMP_DIFF(TIMESTAMP(JSON_VALUE(payload, "$.pull_request.closed_at")), TIMESTAMP(JSON_VALUE(payload, "$.pull_request.created_at")), SECOND) open_duration_s,
  PARSE_JSON(payload) payload
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "pull_request";
