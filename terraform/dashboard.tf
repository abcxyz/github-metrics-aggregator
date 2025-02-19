resource "google_monitoring_dashboard" "default" {
  project = var.project_id

  dashboard_json = file("${path.module}/dashboards/default.json")

  depends_on = [
    module.webhook_alerts,
    module.retry_alerts,
    module.leech,
    module.commit_review_status,
  ]
}
