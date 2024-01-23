output "pr_stats_looker_studio_report_link" {
  description = "The Looker Studio Linking API link for connecting the data sources for the PR Stats dashboard."
  value = join("",
    [
      "https://lookerstudio.google.com/reporting/create",
      "?c.reportId=${var.looker_report_id}",
      "&r.reportName=PR%20Stats",
      "&ds.ds0.keepDatasourceName",
      "&ds.ds0.connector=bigQuery",
      "&ds.ds0.refreshFields",
      "&ds.ds0.projectId=${var.project_id}",
      "&ds.ds0.type=TABLE",
      "&ds.ds0.datasetId=${var.dataset_id}",
      "&ds.ds0.tableId=pull_requests",
      "&ds.ds2.keepDatasourceName",
      "&ds.ds2.connector=bigQuery",
      "&ds.ds2.refreshFields",
      "&ds.ds2.projectId=${var.project_id}",
      "&ds.ds2.type=TABLE",
      "&ds.ds2.datasetId=${var.dataset_id}",
      "&ds.ds2.tableId=pull_request_reviews",
    ]
  )
}
