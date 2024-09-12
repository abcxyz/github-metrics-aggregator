resource "google_monitoring_notification_channel" "non_paging" {
  for_each = var.alert_notification_channel_non_paging

  project = var.project_id

  display_name = "Non-paging Notification Channel"
  type         = each.key
  labels       = each.value.labels
}
