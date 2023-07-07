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

-- Filters events to those just pertaining to deployments.
-- Relevant GitHub Docs:
-- https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#deployment_status

SELECT
  received,
  event,
  JSON_VALUE(payload, "$.action") action,
  JSON_VALUE(payload, "$.organization.login") ORGANIZATION,
  SAFE_CAST(JSON_VALUE(payload, "$.organization.id") AS INT64) organization_id,
  JSON_VALUE(payload, "$.repository.full_name") repository_full_name,
  SAFE_CAST(JSON_VALUE(payload, "$.repository.id") AS INT64) repository_id,
  JSON_VALUE(payload, "$.repository.name") repository,
  JSON_VALUE(payload, "$.repository.visibility") repository_visibility,
  JSON_VALUE(payload, "$.sender.login") sender,
  SAFE_CAST(JSON_VALUE(payload, "$.sender.id") AS INT64) sender_id,
  SAFE_CAST(JSON_VALUE(payload, "$.check_run.id") AS INT64) check_run_id,
  TIMESTAMP(JSON_VALUE(payload, "$.deployment_status.created_at")) created_at,
  JSON_VALUE(payload, "$.deployment_status.creator.login") creator,
  SAFE_CAST(JSON_VALUE(payload, "$.deployment_status.creator.id") AS INT64) creator_id,
  SAFE_CAST(JSON_VALUE(payload, "$.deployment.id") AS INT64) deployment_id,
  JSON_VALUE(payload, "$.deployment_status.description") description,
  JSON_VALUE(payload, "$.deployment_status.environment") environment,
  SAFE_CAST(JSON_VALUE(payload, "$.deployment_status.id") AS INT64) id,
  JSON_VALUE(payload, "$.deployment_status.state") state,
  JSON_VALUE(payload, "$.deployment_status.target_url") target_url,
  SAFE_CAST(JSON_VALUE(payload, "$.workflow.id") AS INT64) workflow_id,
  SAFE_CAST(JSON_VALUE(payload, "$.workflow_run.id") AS INT64) workflow_run_id,
  TIMESTAMP(JSON_VALUE(payload, "$.deployment_status.updated_at")) updated_at,
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "deployment_status";
