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
-- Extracts all distinct workflow_runs from the workflow_run_events view.
-- The most recent event for each distinct workflow_run id is used to extract
-- the workflow_run data. This ensures that the workflow_run data is up to date.

SELECT
  workflow_run_events.actor,
  workflow_run_events.conclusion,
  workflow_run_events.created_at,
  workflow_run_events.delivery_id,
  workflow_run_events.display_title,
  workflow_run_events.duration_seconds,
  workflow_run_events.workflow_event,
  workflow_run_events.head_branch,
  workflow_run_events.head_sha,
  workflow_run_events.html_url,
  workflow_run_events.id,
  workflow_run_events.organization,
  workflow_run_events.organization_id,
  workflow_run_events.path,
  workflow_run_events.repository,
  workflow_run_events.repository_id,
  workflow_run_events.run_attempt,
  workflow_run_events.run_number,
  workflow_run_events.run_started_at,
  workflow_run_events.status,
  workflow_run_events.updated_at,
  workflow_run_events.workflow_name,
  workflow_run_events.workflow_id,
  workflow_run_events.workflow_html_url,
FROM
  `${dataset_id}.workflow_run_events` workflow_run_events
INNER JOIN (
  SELECT
    id,
    MAX(received) received
  FROM
    `${dataset_id}.workflow_run_events`
  GROUP BY
    id ) unique_workflow_run_ids
ON
  workflow_run_events.id = unique_workflow_run_ids.id
  AND workflow_run_events.received = unique_workflow_run_ids.received;
