SELECT
  received,
  event,
  JSON_VALUE(payload, "$.organization.login") owner,
  JSON_VALUE(payload, "$.repository.name") repo,
  JSON_VALUE(payload, "$.repository.full_name") repo_full_name,
  JSON_VALUE(payload, "$.sender.login") sender,
  JSON_VALUE(payload, "$.action") action,
  JSON_VALUE(payload, "$.issue.title") title,
  JSON_VALUE(payload, "$.issue.comments") comments,
  JSON_VALUE(payload, "$.issue.state") state,
  JSON_VALUE(payload, "$.issue.state_reason") state_reason,
  JSON_VALUE(payload, "$.issue.html_url") html_url,
  TIMESTAMP(JSON_VALUE(payload, "$.issue.created_at")) created_at,
  TIMESTAMP(JSON_VALUE(payload, "$.issue.closed_at")) closed_at,
  TIMESTAMP_DIFF(TIMESTAMP(JSON_VALUE(payload, "$.issue.closed_at")), TIMESTAMP(JSON_VALUE(payload, "$.issue.created_at")), SECOND) open_duration_s,
  PARSE_JSON(payload) payload
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "issues";