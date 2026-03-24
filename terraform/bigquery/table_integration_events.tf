module "integration_events_table" {
  source = "./modules/table"

  project_id = var.project_id

  dataset_id = google_bigquery_dataset.default.dataset_id

  table_id = var.integration_events_table_id
  schema = jsonencode([
    {
      "name" : "username",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The username of the user who triggered the event."
    },
    {
      "name" : "timestamp_seconds",
      "type" : "TIMESTAMP",
      "mode" : "NULLABLE",
      "description" : "The time the action took place."
    },
    {
      "name" : "event_type",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Describes the developer tool or system the event relates to."
    },
    {
      "name" : "action_type",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Describes the relationship between the event and the user's workflow."
    },
    {
      "name" : "source",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The source of the GitHub event."
    },
    {
      "name" : "github_event_type",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The type of GitHub event."
    },
    {
      "name" : "delivery_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "Delivery id of the event from GH."
    },
    {
      "name" : "event_id",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "This is the pull_request id, comment id, etc."
    },
    {
      "name" : "repository_url",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The URL of the repository."
    },
    {
      "name" : "ref_sha",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The SHA of the commit."
    },
    {
      "name" : "enterprise",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The enterprise."
    },
    {
      "name" : "organization",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The organization that the repository belongs to."
    },
    {
      "name" : "repository",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The repository name."
    },
    {
      "name" : "ref_name",
      "type" : "STRING",
      "mode" : "NULLABLE",
      "description" : "The ref name."
    }
  ])

  time_partitioning = {
    field = "timestamp_seconds"
    type  = var.bigquery_events_partition_granularity
  }

  clustering = ["event_type", "action_type"]
}
