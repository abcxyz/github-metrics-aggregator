SELECT
  received,
  event,
  JSON_VALUE(payload, "$.organization.login") owner,
  JSON_VALUE(payload, "$.repository.name") repo,
  JSON_VALUE(payload, "$.repository.full_name") repo_full_name,
  JSON_VALUE(payload, "$.sender.login") sender,
  JSON_VALUE(payload, "$.ref") ref,
  JSON_VALUE(payload, "$.base_ref") base_ref,
  JSON_VALUE(payload, "$.compare") compare_url,
  JSON_QUERY_ARRAY(payload, "$.commits") commits,
  ARRAY_LENGTH(JSON_QUERY_ARRAY(payload, "$.commits")) commit_count,
  JSON_VALUE(payload, "$.pusher.email") pusher,
  PARSE_JSON(payload) payload
FROM
  `${dataset_id}.${table_id}`
WHERE
  event = "push";