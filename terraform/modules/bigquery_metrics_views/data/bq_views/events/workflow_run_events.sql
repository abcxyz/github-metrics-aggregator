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
-- Filters events to those just pertaining to workflow runs.
-- Relevant GitHub Docs:
-- https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#workflow_run

SELECT
  received,
  event,
  delivery_id,
  JSON_VALUE(payload, "$.action") action,
  organization,
  organization_id,
  repository_full_name,
  repository_id,
  repository,
  repository_visibility,
  sender,
  sender_id,
  JSON_VALUE(payload, "$.workflow_run.actor.login") actor,
  JSON_VALUE(payload, "$.workflow_run.conclusion") conclusion,
  TIMESTAMP(JSON_VALUE(payload, "$.workflow_run.created_at")) created_at,
  JSON_VALUE(payload, "$.workflow_run.display_title") display_title,
  TIMESTAMP_DIFF(TIMESTAMP(JSON_VALUE(payload, "$.workflow_run.updated_at")), TIMESTAMP(JSON_VALUE(payload, "$.workflow_run.run_started_at")), SECOND) duration_seconds,
  JSON_VALUE(payload, "$.workflow_run.event") workflow_event,
  JSON_VALUE(payload, "$.workflow_run.head_branch") head_branch,
  JSON_VALUE(payload, "$.workflow_run.head_sha") head_sha,
  JSON_VALUE(payload, "$.workflow_run.html_url") html_url,
  SAFE_CAST(JSON_VALUE(payload, "$.workflow_run.id") AS INT64) id,
  JSON_VALUE(payload, "$.workflow_run.path") path,
  SAFE_CAST(JSON_VALUE(payload, "$.workflow_run.run_attempt") AS INT64) run_attempt,
  SAFE_CAST(JSON_VALUE(payload, "$.workflow_run.run_number") AS INT64) run_number,
  TIMESTAMP(JSON_VALUE(payload, "$.workflow_run.run_started_at")) run_started_at,
  JSON_VALUE(payload, "$.workflow_run.status") status,
  TIMESTAMP(JSON_VALUE(payload, "$.workflow_run.updated_at")) updated_at,
  JSON_VALUE(payload, "$.workflow_run.name") workflow_name,
  SAFE_CAST(JSON_VALUE(payload, "$.workflow_run.workflow_id") AS INT64) workflow_id,
  JSON_VALUE(payload, "$.workflow.html_url") workflow_html_url,
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "workflow_run";
