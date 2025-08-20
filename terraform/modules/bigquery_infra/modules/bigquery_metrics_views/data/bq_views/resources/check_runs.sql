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

-- Extracts all distinct check_runs from the check_run_events view.
-- The most recent event for each distinct check_run id is used to extract
-- the check_run data. This ensures that the check_run data is up to date.

SELECT
  check_run_events.organization,
  check_run_events.organization_id,
  check_run_events.repository_id,
  check_run_events.repository,
  check_run_events.app,
  check_run_events.app_id,
  check_run_events.check_suite_id,
  check_run_events.completed_at,
  check_run_events.conclusion,
  check_run_events.delivery_id,
  check_run_events.deployment_id,
  check_run_events.details_url,
  check_run_events.external_id,
  check_run_events.head_sha,
  check_run_events.html_url,
  check_run_events.id,
  check_run_events.name,
  check_run_events.output_summary,
  check_run_events.output_text,
  check_run_events.output_title,
  check_run_events.started_at,
  check_run_events.status,
FROM
  `${dataset_id}.check_run_events` check_run_events
JOIN (
  SELECT
    id,
    MAX(received) received
  FROM
    `${dataset_id}.check_run_events`
  GROUP BY
    id ) unique_check_run_ids
ON
  check_run_events.id = unique_check_run_ids.id
  AND check_run_events.received = unique_check_run_ids.received;
