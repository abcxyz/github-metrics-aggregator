output "gclb_external_ip_name" {
  description = "The external IPv4 name assigned to the global fowarding rule for the global load balancer."
  value       = module.REPLACE_MODULE_NAME.gclb_external_ip_name
}

output "gclb_external_ip_address" {
  description = "The external IPv4 assigned to the global fowarding rule for the global load balancer."
  value       = module.REPLACE_MODULE_NAME.gclb_external_ip_address
}

output "webhook_run_service_name" {
  description = "The webhook Cloud Run service name."
  value       = module.REPLACE_MODULE_NAME.webhook_run_service.service_name
}

output "webhook_run_service_name_url" {
  description = "The webhook Cloud Run service url."
  value       = module.REPLACE_MODULE_NAME.webhook_run_service.service_url
}

output "retry_run_service_name" {
  description = "The retry Cloud Run service name."
  value       = module.REPLACE_MODULE_NAME.retry_run_service.service_name
}

output "retry_run_service_url" {
  description = "The retry Cloud Run service url."
  value       = module.REPLACE_MODULE_NAME.retry_run_service.service_url
}

output "github_metrics_looker_studio_report_link" {
  description = "The Looker Studio Linking API link for connecting the data sources for the GitHub Metrics dashboard."
  value       = module.REPLACE_MODULE_NAME.github_metrics_looker_studio_report_link
}
