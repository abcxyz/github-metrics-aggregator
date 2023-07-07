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

-- Extracts all distinct deployment_statuses from the deployment_status_events view.
-- The most recent event for each distinct deployment_status id is used to extract
-- the deployment_status data. This ensures that the deployment_status data is up to date.

SELECT
  deployment_status_events.organization,
  deployment_status_events.organization_id,
  deployment_status_events.repository_id,
  deployment_status_events.repository,
  deployment_status_events.check_run_id,
  deployment_status_events.created_at,
  deployment_status_events.creator,
  deployment_status_events.creator_id,
  deployment_status_events.deployment_id,
  deployment_status_events.description,
  deployment_status_events.environment,
  deployment_status_events.id,
  deployment_status_events.state,
  deployment_status_events.target_url,
  deployment_status_events.workflow_id,
  deployment_status_events.workflow_run_id,
  deployment_status_events.updated_at,
FROM
  `${dataset_id}.deployment_status_events` deployment_status_events
JOIN (
  SELECT
    id,
    MAX(received) received
  FROM
    `${dataset_id}.deployment_status_events`
  GROUP BY
    id ) unique_deployment_status_ids
ON
  deployment_status_events.id = unique_deployment_status_ids.id
  AND deployment_status_events.received = unique_deployment_status_ids.received
