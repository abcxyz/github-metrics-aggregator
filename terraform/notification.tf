resource "google_monitoring_notification_channel" "non_paging" {
  for_each = {
    for k, v in var.alert_notification_channel_non_paging : k => v
    if k == "email" && try(v.labels.email_address, "") != ""
  }

  project = var.project_id

  display_name = "Non-paging Notification Channel"
  type         = each.key
  labels       = each.value.labels
}
