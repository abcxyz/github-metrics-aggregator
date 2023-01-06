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
  JSON_VALUE(payload, "$.ref") ref,
  JSON_VALUE(payload, "$.base_ref") base_ref,
  JSON_VALUE(payload, "$.compare") compare_url,
  JSON_QUERY_ARRAY(payload, "$.commits") commits,
  ARRAY_LENGTH(JSON_QUERY_ARRAY(payload, "$.commits")) commit_count,
  JSON_VALUE(payload, "$.pusher.email") pusher,
  PARSE_JSON(payload) payload
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "push";
