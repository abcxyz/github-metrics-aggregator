-- Copyright 2023 The Authors (see AUTHORS file)
--
-- Licensed under the Apache License, Version 2.0 (the 'License');
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an 'AS IS' BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

-- TVF queries cannot end in ; as it will be wrapped in a CREATE TABLE FUNCTION
-- and the end of the query is not the end of the compiled SQL statement.

SELECT
    received,
    event,
    delivery_id,
    organization,
    organization_id,
    repository_full_name,
    repository_id,
    repository,
    repository_visibility,
    sender,
    sender_id,
    LAX_STRING(payload.after) after_sha,
    LAX_STRING(payload.before) before_sha,
    LAX_STRING(payload.compare) compare_url,
    JSON_QUERY_ARRAY(payload.commits) commits,
    ARRAY_LENGTH(JSON_QUERY_ARRAY(payload.commits)) commit_count,
    SAFE.BOOL(payload.created) created,
    SAFE.BOOL(payload.deleted) deleted,
    SAFE.BOOL(payload.forced) forced,
    LAX_STRING(payload.pusher.name) pusher,
    LAX_STRING(payload.ref) ref,
    LAX_STRING(payload.repository.default_branch) repository_default_branch
FROM
  `${parent_project_id}.${parent_dataset_id}.${parent_routine_id}`(startTimestamp, endTimestamp, 'push')
