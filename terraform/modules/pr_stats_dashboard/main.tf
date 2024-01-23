# add all groups who need to view through lookerstudio to jobUser role
resource "google_project_iam_member" "pr_stats_dashboard_job_users" {
  for_each = toset(var.viewers)

  project = var.project_id

  role = "roles/bigquery.jobUser"
  member = each.value
}

resource "google_bigquery_dataset_iam_member" "pr_stats_dashboard_data_viewers" {
  for_each = toset(var.viewers)

  project = var.project_id

  dataset_id = var.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = each.value
}
