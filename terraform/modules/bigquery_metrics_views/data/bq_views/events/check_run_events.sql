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

-- Filters events to those just pertaining to check_runs.
-- Relevant GitHub Docs:
-- https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#check_run

SELECT
  received,
  event,
  JSON_VALUE(payload, "$.action") action,
  JSON_VALUE(payload, "$.organization.login") organization,
  SAFE_CAST(JSON_VALUE(payload, "$.organization.id") AS INT64) organization_id,
  JSON_VALUE(payload, "$.repository.full_name") repository_full_name,
  SAFE_CAST(JSON_VALUE(payload, "$.repository.id") AS INT64) repository_id,
  JSON_VALUE(payload, "$.repository.name") repository,
  JSON_VALUE(payload, "$.repository.visibility") repository_visibility,
  JSON_VALUE(payload, "$.sender.login") sender,
  SAFE_CAST(JSON_VALUE(payload, "$.sender.id") AS INT64) sender_id,
  JSON_VALUE(payload, "$.check_run.app.name") app,
  SAFE_CAST(JSON_VALUE(payload, "$.check_run.app.id") AS INT64) app_id,
  SAFE_CAST(JSON_VALUE(payload, "$.check_run.check_suite.id") AS INT64) check_suite_id,
  TIMESTAMP(JSON_VALUE(payload, "$.check_run.completed_at")) completed_at,
  JSON_VALUE(payload, "$.check_run.conclusion") conclusion,
  SAFE_CAST(JSON_VALUE(payload, "$.check_run.deployment.id") AS INT64) deployment_id,
  JSON_VALUE(payload, "$.check_run.details_url") details_url,
  JSON_VALUE(payload, "$.check_run.external_id") external_id,
  JSON_VALUE(payload, "$.check_run.head_sha") head_sha,
  JSON_VALUE(payload, "$.check_run.html_url") html_url,
  SAFE_CAST(JSON_VALUE(payload, "$.check_run.id") AS INT64) id,
  JSON_VALUE(payload, "$.check_run.name") name,
  JSON_VALUE(payload, "$.check_run.output.summary") output_summary,
  JSON_VALUE(payload, "$.check_run.output.text") output_text,
  JSON_VALUE(payload, "$.check_run.output.title") output_title,
  TIMESTAMP(JSON_VALUE(payload, "$.check_run.started_at")) started_at,
  JSON_VALUE(payload, "$.check_run.status") status,
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "check_run";
