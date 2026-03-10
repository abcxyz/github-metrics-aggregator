resource "google_bigquery_data_transfer_config" "integration_events_schedule" {
  project = var.project_id

  display_name           = "gma_integration_events_query"
  location               = var.location
  data_source_id         = "scheduled_query"
  schedule               = var.integration_events_schedule
  service_account_name   = var.prstats_service_account_email
  destination_dataset_id = var.dataset_id
  params = {
    query = <<EOT
INSERT INTO `${var.project_id}.${var.dataset_id}.${var.integration_events_table_name}`
  (
    WITH new_events AS (
      SELECT
        LOWER(CAST(COALESCE(
          JSON_VALUE(payload, '$.sender.login'),
          JSON_VALUE(payload, '$.actor.login'),
          JSON_VALUE(payload, '$.pull_request.user.login'),
          JSON_VALUE(payload, '$.issue.user.login')
        ) AS STRING)) AS username,
        TIMESTAMP_TRUNC(CAST(received AS TIMESTAMP), SECOND) AS timestamp_seconds,
        'GITHUB' AS event_type,
        CASE
          WHEN event = 'create' AND JSON_VALUE(payload, '$.ref_type') = 'branch' THEN 'BRANCH'
          WHEN event = 'create' AND JSON_VALUE(payload, '$.ref_type') = 'tag' THEN 'TAG'
          WHEN event = 'deployment' AND JSON_VALUE(payload, '$.action') = 'created' AND JSON_VALUE(payload, '$.deployment_status.state') IS NULL THEN 'DEPLOYMENT_STARTED'
          WHEN event = 'deployment_status' AND JSON_VALUE(payload, '$.deployment_status.state') = 'success' THEN 'DEPLOYMENT_ENDED'
          WHEN event = 'deployment_status' AND JSON_VALUE(payload, '$.deployment_status.state') IN ('failure', 'error') THEN 'DEPLOYMENT_FAILED'
          WHEN event = 'pull_request' AND JSON_VALUE(payload, '$.action') = 'assigned' THEN 'PR_ASSIGNED'
          WHEN event = 'pull_request' AND JSON_VALUE(payload, '$.action') = 'closed' AND JSON_VALUE(payload, '$.pull_request.merged') = 'false' THEN 'PR_ABANDONED'
          WHEN event = 'pull_request' AND JSON_VALUE(payload, '$.action') = 'closed' AND JSON_VALUE(payload, '$.pull_request.merged') = 'true' THEN 'PR_CLOSED'
          WHEN event = 'pull_request' AND JSON_VALUE(payload, '$.action') = 'opened' THEN 'PR_OPENED'
          WHEN event = 'pull_request' AND JSON_VALUE(payload, '$.action') = 'ready_for_review' THEN 'PR_READY'
          WHEN event = 'pull_request' AND JSON_VALUE(payload, '$.action') = 'review_requested' THEN 'PR_REVIEW_REQUEST'
          WHEN event = 'pull_request_review' AND JSON_VALUE(payload, '$.action') = 'submitted' AND JSON_VALUE(payload, '$.review.state') IS NOT NULL THEN 'PR_SIGNOFF'
          WHEN event = 'pull_request_review_comment' AND JSON_VALUE(payload, '$.action') IN ('created', 'deleted', 'edited') THEN 'PR_COMMENT'
          WHEN event = 'push' THEN 'PUSH'
          WHEN event = 'repository' AND JSON_VALUE(payload, '$.action') = 'created' THEN 'REPO'
          WHEN event = 'workflow_job' AND JSON_VALUE(payload, '$.action') = 'queued' THEN 'JOB_CREATED'
          WHEN event = 'workflow_job' AND JSON_VALUE(payload, '$.action') = 'completed' THEN 'JOB_COMPLETED'
          WHEN event = 'workflow_run' AND JSON_VALUE(payload, '$.action') = 'requested' THEN 'RUN_CREATED'
          WHEN event = 'workflow_run' AND JSON_VALUE(payload, '$.action') = 'completed' THEN 'RUN_COMPLETED'
          ELSE 'GIT_HUB_EVENT_TYPE_UNSPECIFIED'
        END AS github_event_type,
        CASE
          WHEN event IN ('create', 'pull_request', 'pull_request_review', 'pull_request_review_comment', 'push', 'repository') THEN 'USER_INTERACTION'
          WHEN event IN ('deployment', 'deployment_status', 'workflow_job', 'workflow_run') THEN 'SYSTEM_ACTION'
          ELSE 'ACTION_TYPE_UNSPECIFIED'
        END AS action_type,
        IF(
          COALESCE(enterprise_name, '') LIKE 'Google-%',
          'GITHUB_SOURCE_GITHUB_ONPREM',
          'GITHUB_SOURCE_GITHUB_COM') AS source,
        delivery_id,
        CAST(COALESCE(
          JSON_VALUE(payload, '$.pull_request.id'),
          JSON_VALUE(payload, '$.issue.id'),
          JSON_VALUE(payload, '$.comment.id'),
          JSON_VALUE(payload, '$.workflow_run.id')
        ) AS STRING) AS event_id,
        CAST(JSON_VALUE(payload, '$.repository.html_url') AS STRING) AS repository_url,
        CAST(COALESCE(
          JSON_VALUE(payload, '$.pull_request.head.sha'),
          JSON_VALUE(payload, '$.check_suite.head_sha'),
          JSON_VALUE(payload, '$.workflow_run.head_sha')
        ) AS STRING) AS ref_sha,
        enterprise_name AS enterprise,
        organization_name AS organization,
        repository_name AS repository,
        CAST(COALESCE(
            JSON_VALUE(payload, '$.pull_request.head.ref'),
            JSON_VALUE(payload, '$.ref')
        ) AS STRING) AS ref_name
      FROM `${var.project_id}.${var.dataset_id}.${var.optimized_events_table_name}`
      WHERE received > (
          SELECT COALESCE(MAX(timestamp_seconds), TIMESTAMP('2015-01-01 00:00:00 UTC'))
          FROM `${var.project_id}.${var.dataset_id}.${var.integration_events_table_name}`
      )
    )
    SELECT * FROM new_events
    WHERE github_event_type != 'GIT_HUB_EVENT_TYPE_UNSPECIFIED'
  )
EOT
  }

  depends_on = [google_project_iam_member.bq_transfer_permission]
}
