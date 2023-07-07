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
-- Extracts all distinct deployments from the deployment_events view.
-- The most recent event for each distinct deployment id is used to extract
-- the deployment data. This ensures that the deployment data is up to date.

SELECT
  deployment_events.organization,
  deployment_events.organization_id,
  deployment_events.repository_id,
  deployment_events.repository,
  deployment_events.created_at,
  deployment_events.creator,
  deployment_events.creator_id,
  deployment_events.description,
  deployment_events.environment,
  deployment_events.id,
  deployment_events.original_environment,
  deployment_events.deployment_payload,
  deployment_events.production_environment,
  deployment_events.ref,
  deployment_events.sha,
  deployment_events.task,
  deployment_events.transient_environment,
  deployment_events.updated_at,
FROM `${dataset_id}.deployment_events` deployment_events
JOIN (
  SELECT id, MAX(received) received
  FROM `${dataset_id}.deployment_events`
  GROUP BY id
) unique_deployment_ids
ON deployment_events.id = unique_deployment_ids.id
AND deployment_events.received = unique_deployment_ids.received;
