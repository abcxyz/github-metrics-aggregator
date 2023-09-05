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
-- Extracts all distinct issues from the issue_events view.
-- The most recent event for each distinct issue id is used to extract
-- the issue data. This ensures that the issue data is up to date.

SELECT
  issue_events.active_lock_reason,
  issue_events.assignee,
  issue_events.author,
  issue_events.author_association,
  issue_events.body,
  issue_events.closed_at,
  issue_events.comments,
  issue_events.created_at,
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
  `${dataset_id}.issue_events` issue_events
INNER JOIN (
  SELECT
    id,
    MAX(received) received
  FROM
    `${dataset_id}.issue_events`
  GROUP BY
    id ) unique_issue_ids
ON
  issue_events.id = unique_issue_ids.id
  AND issue_events.received = unique_issue_ids.received;
