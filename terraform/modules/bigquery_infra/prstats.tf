resource "google_bigquery_table" "prstats_pull_requests_table" {
  project = var.project_id

  deletion_protection = true
  table_id            = var.prstats_pull_requests_table_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  schema = jsonencode(
    [
      {
        "name" : "enterprise_id",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The ID of the enterprise."
      },
      {
        "name" : "enterprise_name",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The name of the enterprise."
      },
      {
        "name" : "organization_id",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The ID of the organization."
      },
      {
        "name" : "organization_name",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The name of the organization."
      },
      {
        "name" : "repository_id",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The ID of the repository."
      },
      {
        "name" : "repository_name",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The name of the repository."
      },
      {
        "name" : "pr_id",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The ID of the pull request."
      },
      {
        "name" : "url",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The URL of the pull request."
      },
      {
        "name" : "author",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The author of the pull request."
      },
      {
        "name" : "insertions",
        "type" : "INTEGER",
        "mode" : "NULLABLE",
        "description" : "The number of insertions in the pull request."
      },
      {
        "name" : "deletions",
        "type" : "INTEGER",
        "mode" : "NULLABLE",
        "description" : "The number of deletions in the pull request."
      },
      {
        "name" : "created",
        "type" : "TIMESTAMP",
        "mode" : "NULLABLE",
        "description" : "The timestamp when the pull request was created."
      },
      {
        "name" : "submitted",
        "type" : "TIMESTAMP",
        "mode" : "NULLABLE",
        "description" : "The timestamp when the pull request was submitted."
      },
      {
        "name" : "updated",
        "type" : "TIMESTAMP",
        "mode" : "NULLABLE",
        "description" : "The timestamp when the pull request was updated."
      },
      {
        "name" : "received",
        "type" : "TIMESTAMP",
        "mode" : "NULLABLE",
        "description" : "The timestamp when the pull request event was received."
      },
      {
        "name" : "reviews",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The reviewers of the pull request. This is a comma-separated list of reviewer names. Format: '<reviewer1_name>,<reviewer2_name>'."
      },
      {
        "name" : "comments",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The commenters of the pull request. This is a comma-separated list of commenter names with a pipe (|) separator for the timestamp of the comment. Format: '<commenter1_name>|<timestamp>,<commenter2_name>|<timestamp>'"
      }
    ]
  )

  time_partitioning {
    field = "updated"
    type  = var.bigquery_prstats_partition_granularity
  }

  clustering = ["author", "enterprise_name", "organization_name", "repository_name"]
}

resource "google_bigquery_table" "prstats_pull_request_reviews_table" {
  project = var.project_id

  deletion_protection = true
  table_id            = var.prstats_pull_request_reviews_table_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  schema = jsonencode(
    [
      {
        "name" : "enterprise_id",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The ID of the enterprise."
      },
      {
        "name" : "enterprise_name",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The name of the enterprise."
      },
      {
        "name" : "organization_id",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The ID of the organization."
      },
      {
        "name" : "organization_name",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The name of the organization."
      },
      {
        "name" : "repository_id",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The ID of the repository."
      },
      {
        "name" : "repository_name",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The name of the repository."
      },
      {
        "name" : "pr_id",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The ID of the pull request."
      },
      {
        "name" : "url",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The URL of the pull request."
      },
      {
        "name" : "author",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The author of the pull request."
      },
      {
        "name" : "insertions",
        "type" : "INTEGER",
        "mode" : "NULLABLE",
        "description" : "The number of insertions in the pull request."
      },
      {
        "name" : "deletions",
        "type" : "INTEGER",
        "mode" : "NULLABLE",
        "description" : "The number of deletions in the pull request."
      },
      {
        "name" : "created",
        "type" : "TIMESTAMP",
        "mode" : "NULLABLE",
        "description" : "The timestamp when the pull request was created."
      },
      {
        "name" : "submitted",
        "type" : "TIMESTAMP",
        "mode" : "NULLABLE",
        "description" : "The timestamp when the pull request was submitted."
      },
      {
        "name" : "updated",
        "type" : "TIMESTAMP",
        "mode" : "NULLABLE",
        "description" : "The timestamp when the pull request was updated."
      },
      {
        "name" : "received",
        "type" : "TIMESTAMP",
        "mode" : "NULLABLE",
        "description" : "The timestamp when the pull request event was received."
      },
      {
        "name" : "reviewer",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The reviewer of the pull request."
      },
      {
        "name" : "reviews",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The reviewers of the pull request. This is a comma-separated list of reviewer names. Format: '<reviewer1_name>,<reviewer2_name>'."
      },
      {
        "name" : "comments",
        "type" : "STRING",
        "mode" : "NULLABLE",
        "description" : "The commenters of the pull request. This is a comma-separated list of commenter names with a pipe (|) separator for the timestamp of the comment. Format: '<commenter1_name>|<timestamp>,<commenter2_name>|<timestamp>'."
      }
    ]
  )

  time_partitioning {
    field = "updated"
    type  = var.bigquery_prstats_partition_granularity
  }

  clustering = ["reviewer", "enterprise_name", "organization_name", "repository_name"]
}

resource "google_bigquery_table_iam_member" "prstats_pull_requests_service_account_iam" {
  for_each = toset(["roles/bigquery.dataEditor"])

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.prstats_pull_requests_table.table_id
  role       = each.key
  member     = var.prstats_service_account_member
}

resource "google_bigquery_table_iam_member" "prstats_pull_request_reviews_service_account_iam" {
  for_each = toset(["roles/bigquery.dataEditor"])

  project = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.prstats_pull_request_reviews_table.table_id
  role       = each.key
  member     = var.prstats_service_account_member
}
