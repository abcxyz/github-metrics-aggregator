locals { cloudscheduler_job_creator = "cloudschedulerJobCreator" }
resource "google_project_iam_custom_role" "cloudscheduler_job_creator" {
  project      = var.project_id

  role_id     = local.cloudscheduler_job_creator
  title       = "Cloud Scheduler Job Creator"
  description = "Access to create Cloud Scheduler jobs"
  permissions = [
    "cloudscheduler.jobs.create",
  ]
}

locals { cloudstorage_bucket_creator = "cloudstorageBucketCreator" }
resource "google_project_iam_custom_role" "cloudstorage_bucket_creator" {
  project      = var.project_id

  role_id     = local.cloudstorage_bucket_creator
  title       = "Cloud Storage Bucket Creator"
  description = "Access to create GCS buckets"
  permissions = [
    "storage.buckets.create",
  ]
}

locals { secretmanager_secret_creator = "secretmanagerSecretCreator" }
resource "google_project_iam_custom_role" "secretmanager_secret_creator" {
  project      = var.project_id

  role_id     = local.secretmanager_secret_creator
  title       = "Secret Manager Secret Creator"
  description = "Access to create secrets in Secret Manager"
  permissions = [
    "secretmanager.secrets.create",
  ]
}
