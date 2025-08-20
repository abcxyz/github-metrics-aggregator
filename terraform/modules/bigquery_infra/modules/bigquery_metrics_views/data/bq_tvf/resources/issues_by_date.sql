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

-- TVF queries cannot end in ; as it will be wrapped in a CREATE TABLE FUNCTION
-- and the end of the query is not the end of the compiled SQL statement.

SELECT
  issue_events.active_lock_reason,
  issue_events.assignee,
  issue_events.author,
  issue_events.author_association,
  issue_events.body,
  issue_events.closed_at,
  issue_events.comments,
  issue_events.created_at,
  issue_events.delivery_id,
  issue_events.draft,
  issue_events.html_url,
  issue_events.id,
  issue_events.locked,
  issue_events.number,
  issue_events.open_duration_seconds,
  issue_events.organization,
  issue_events.organization_id,
  issue_events.repository,
  issue_events.repository_id,
  issue_events.state,
  issue_events.state_reason,
  issue_events.title,
  issue_events.updated_at,
FROM
  `${parent_project_id}.${parent_dataset_id}.${parent_routine_id}`(startTimestamp,
    endTimestamp) issue_events
INNER JOIN (
  SELECT
    id,
    MAX(received) received
  FROM
    `${parent_project_id}.${parent_dataset_id}.${parent_routine_id}`(startTimestamp,
      endTimestamp)
  GROUP BY
    id ) unique_issue_ids
ON
  issue_events.id = unique_issue_ids.id
  AND issue_events.received = unique_issue_ids.received
