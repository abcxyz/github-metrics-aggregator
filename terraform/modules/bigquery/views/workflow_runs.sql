SELECT
  received,
  event,
  JSON_VALUE(payload, "$.organization.login") owner,
  JSON_VALUE(payload, "$.repository.name") repo,
  JSON_VALUE(payload, "$.repository.full_name") repo_full_name,
  JSON_VALUE(payload, "$.sender.login") sender,
  JSON_VALUE(payload, "$.action") action,
  TIMESTAMP(JSON_VALUE(payload, "$.workflow_run.run_started_at")) started_at,
  TIMESTAMP(JSON_VALUE(payload, "$.workflow_run.updated_at")) updated_at,
  TIMESTAMP_DIFF(TIMESTAMP(JSON_VALUE(payload, "$.workflow_run.updated_at")), TIMESTAMP(JSON_VALUE(payload, "$.workflow_run.run_started_at")), SECOND) duration_s,
  JSON_VALUE(payload, "$.workflow_run.status") status,
  JSON_VALUE(payload, "$.workflow_run.conclusion") conclusion,
  JSON_VALUE(payload, "$.workflow_run.name") name,
  JSON_VALUE(payload, "$.workflow_run.display_title") title,
  JSON_VALUE(payload, "$.workflow_run.event") workflow_event,
  JSON_VALUE(payload, "$.workflow_run.head_branch") head_branch,
  JSON_VALUE(payload, "$.workflow_run.html_url") html_url,
  PARSE_JSON(payload) payload
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "workflow_run";