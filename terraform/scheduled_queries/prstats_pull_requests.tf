resource "google_bigquery_data_transfer_config" "prstats_pull_requests_schedule" {
  project = var.project_id

  display_name           = "gma_prstats_pull_requests_query"
  location               = var.location
  data_source_id         = "scheduled_query"
  destination_dataset_id = var.dataset_id
  schedule               = var.prstats_pull_requests_schedule
  service_account_name   = var.prstats_service_account_email
  params = {
    query = <<EOT
INSERT INTO `${var.project_id}.${var.dataset_id}.${var.prstats_pull_requests_table_name}`
  (
    WITH
      pull_requests
      AS (
        SELECT
          CAST(JSON_VALUE(payload, '$.enterprise.id') AS STRING) AS enterprise_id,
          CAST(JSON_VALUE(payload, '$.enterprise.name') AS STRING) AS enterprise_name,
          CAST(JSON_VALUE(payload, '$.organization.id') AS STRING) AS organization_id,
          CAST(JSON_VALUE(payload, '$.organization.login') AS STRING) AS organization_name,
          CAST(JSON_VALUE(payload, '$.repository.id') AS STRING) AS repository_id,
          CAST(JSON_VALUE(payload, '$.repository.name') AS STRING) AS repository_name,
          CAST(JSON_VALUE(payload, '$.pull_request.id') AS STRING) AS pr_id,
          CAST(JSON_VALUE(payload, '$.pull_request.html_url') AS STRING) AS url,
          CAST(JSON_VALUE(payload, '$.pull_request.user.login') AS STRING) AS author,
          CAST(JSON_VALUE(payload, '$.pull_request.additions') AS INT64)
            AS insertions,
          CAST(JSON_VALUE(payload, '$.pull_request.deletions') AS INT64)
            AS deletions,
          CAST(JSON_VALUE(payload, '$.pull_request.created_at') AS TIMESTAMP)
            AS created,
          CAST(JSON_VALUE(payload, '$.pull_request.merged_at') AS TIMESTAMP)
            AS submitted,
          CAST(JSON_VALUE(payload, '$.pull_request.updated_at') AS TIMESTAMP)
            AS updated,
          CAST(received AS TIMESTAMP) AS received
        FROM `${var.project_id}.${var.dataset_id}.${var.prstats_source_table_name}`
        WHERE
          event = 'pull_request'
          AND CAST(JSON_VALUE(payload, '$.action') AS STRING) = 'closed'
          AND JSON_VALUE(payload, '$.pull_request.merged_at') IS NOT NULL
      ),
      pr_reviews AS (
        SELECT
          CAST(JSON_VALUE(payload, '$.pull_request.id') AS STRING) AS pr_id,
          CAST(JSON_VALUE(payload, '$.review.user.login') AS STRING) AS reviewer,
          CAST(JSON_VALUE(payload, '$.review.updated_at') AS STRING) AS updated_at,
        FROM `${var.project_id}.${var.dataset_id}.${var.prstats_source_table_name}`
        WHERE
          event = 'pull_request_review'
      ),
      pr_comments AS (
        SELECT
          CAST(JSON_VALUE(payload, '$.pull_request.id') AS STRING) AS pr_id,
          CAST(JSON_VALUE(payload, '$.comment.user.login') AS STRING) AS commenter,
          CAST(JSON_VALUE(payload, '$.comment.updated_at') AS STRING) AS updated_at,
        FROM `${var.project_id}.${var.dataset_id}.${var.prstats_source_table_name}`
        WHERE
          event = 'pull_request_review_comment'
      )
    SELECT
      pull_requests.*,
      (
        SELECT ARRAY_TO_STRING(array_agg(DISTINCT pr_reviews.reviewer), ',')
        FROM pr_reviews
        WHERE
          pr_reviews.pr_id = pull_requests.pr_id
          AND pr_reviews.reviewer != pull_requests.author
      ) AS reviews,
      (
        SELECT
          ARRAY_TO_STRING(
            array_agg(CONCAT(pr_comments.commenter, '|', pr_comments.updated_at)),
            ',')
        FROM pr_comments
        WHERE
          pr_comments.pr_id = pull_requests.pr_id
          AND pr_comments.commenter != pull_requests.author
      ) AS comments
    FROM pull_requests
    WHERE
      pull_requests.received > (
        SELECT COALESCE(MAX(received), TIMESTAMP('2015-01-01 00:00:00 UTC')) FROM `${var.project_id}.${var.dataset_id}.${var.prstats_pull_requests_table_name}`
      )
  )
EOT
  }

  depends_on = [google_project_iam_member.bq_transfer_permission]
}
