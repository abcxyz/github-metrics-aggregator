locals {
  default_view_labels = {
    "type" : "default"
  }
  default_views = {
    issues = {
      template_path  = "${path.module}/views/issues.sql"
      use_legacy_sql = false
    }
    pull_requests = {
      template_path  = "${path.module}/views/pull_requests.sql"
      use_legacy_sql = false
    }
    push = {
      template_path  = "${path.module}/views/push.sql"
      use_legacy_sql = false
    }
    workflow_runs = {
      template_path  = "${path.module}/views/workflow_runs.sql"
      use_legacy_sql = false
    }
  }
}

data "google_project" "project" {
  project_id = var.project_id
}

resource "google_bigquery_dataset" "default" {
  project    = var.project_id
  dataset_id = var.dataset_id
  location   = var.dataset_location
}

resource "google_bigquery_dataset_iam_binding" "bindings" {
  for_each   = var.dataset_iam
  project    = var.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  role       = each.key
  members    = each.value
}

resource "google_bigquery_table" "default" {
  project             = var.project_id
  deletion_protection = false
  table_id            = var.table_id
  dataset_id          = google_bigquery_dataset.default.dataset_id
  schema              = file("${path.module}/bq_schema.json")

  lifecycle {
    ignore_changes = [
      last_modified_time,
      num_rows,
      num_bytes,
      num_long_term_bytes,
    ]
  }

  depends_on = [
    google_bigquery_dataset.default
  ]
}

resource "google_bigquery_table" "default_views" {
  for_each = var.create_default_views ? local.default_views : {}

  project       = var.project_id
  dataset_id    = google_bigquery_dataset.default.dataset_id
  labels        = local.default_view_labels
  friendly_name = each.key
  table_id      = each.key

  view {
    query = templatefile(each.value.template_path, {
      dataset_id = google_bigquery_table.default.dataset_id,
      table_id   = google_bigquery_table.default.table_id
    })
    use_legacy_sql = each.value.use_legacy_sql
  }

  depends_on = [
    google_bigquery_dataset.default,
    google_bigquery_table.default
  ]
}

resource "google_bigquery_table" "views" {
  for_each = var.views

  project       = var.project_id
  dataset_id    = google_bigquery_dataset.default.dataset_id
  labels        = each.value["labels"]
  friendly_name = each.key
  table_id      = each.key

  view {
    query = each.value["query"] == null ? templatefile(each.value["template_path"], {
      dataset_id = google_bigquery_table.default.dataset_id,
      table_id   = google_bigquery_table.default.table_id
    }) : each.value["query"]
    use_legacy_sql = each.value["use_legacy_sql"]
  }

  depends_on = [
    google_bigquery_dataset.default,
    google_bigquery_table.default
  ]
}

resource "google_bigquery_table_iam_binding" "bindings" {
  for_each   = var.table_iam
  project    = var.project_id
  dataset_id = google_bigquery_dataset.default.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = each.key
  members    = each.value
}

resource "google_bigquery_table_iam_member" "default_editor" {
  project    = google_bigquery_table.default.project
  dataset_id = google_bigquery_table.default.dataset_id
  table_id   = google_bigquery_table.default.table_id
  role       = "roles/bigquery.dataEditor"
  member     = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-pubsub.iam.gserviceaccount.com"

  depends_on = [
    google_bigquery_table.default
  ]
}
