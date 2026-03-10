resource "google_project_iam_member" "bq_transfer_permission" {
  project = var.project_id

  role = "roles/iam.serviceAccountShortTermTokenMinter"
  # This is the Google-managed Service Agent for BigQuery Data Transfer Service.
  # It requires this role to mint tokens for running scheduled queries.
  member = "serviceAccount:service-${var.project_number}@gcp-sa-bigquerydatatransfer.iam.gserviceaccount.com"
}
