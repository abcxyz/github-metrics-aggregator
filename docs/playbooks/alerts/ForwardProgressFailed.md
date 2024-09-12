# ForwardProgressFailed

## Services

This alert fires when cloud run services have not made forward progress in an acceptable amount of time. 

- `github-metrics-retry-XXXX` - Checks for failed event deliveries since the last run and retries delivery. The default cadence for this service is to run hourly and is configured by Cloud Scheduler.
- `github-metrics-webhook-XXXX` - Triggered by GitHub webhook events.


### Retry Service Triage Steps

When retry service does not execute within a configured interval, this alert will fire.

To begin triage, find the root cause by doing the following:

1. Go to the Cloud Run page in your GCP project.
2. Confirm the selected tab is Services (this is the default).
3. Select the retry service and select the Revisions tab.
4. Ensure the latest deployment is successful, if not review the logs under the Logs tab and review for errors.
6. If the latest service revision is successful, navigate to Cloud Scheduler by searching for it Cloud Scheduler in the search bar and confirm the cadence is aligned with the expected alert policy interval. If it is not, adjust the alert policy in terraform titled `revision_alert_policy` to better align to the new cadence.

### Webhook Service Triage Steps

The webhook service execution rate is tightly coupled to how active a team is on GitHub. This alert will fire anytime there are no events within a configured interval.

1. Go to the Cloud Run page in your GCP project.
2. Confirm the selected tab is Services (this is the default).
3. Select the retry service and select the Revisions tab.
4. Ensure the latest deployment is successful, if not review the logs under the Logs tab and review for errors.
5. If the latest service revision is successful, or there are no errors, then consider increasing the time window. Alternatively write a workflow in a GitHub repo that executes on an hourly basis. This event should get processed by the webhook service and bridge these gaps of missing data.

## Jobs

This alert fires when background jobs have not made forward progress in an acceptable amount of time. The alert will include the name of the job that is failing to make forward progress. The jobs are invoked in the background.

- `gma-artifacts` - Writes workflow logs from GitHub to Cloud Storage. 

- `commit-review-status-job` - Audits pull requests and writes the results to BigQuery

Each job runs on a different interval. Check your Terraform configuration to see how frequently a specific job runs.

### Cloud Run Job Triage Steps

When one of the jobs does not return success within a configured interval, this alert will fire. For most cases, this means the job has already failed 2+ times.

To begin triage, identify the offending job and its root cause by doing the following:

1. Go to the Cloud Run page in your GCP project.
2. Click on the Jobs tab.
3. Under the History tab, identify the failing execution(s) of the job and select one.
4. Once selected, a second screen of the execution details will appear.
5. Under the Tasks tab, select View Logs.
6. Review logs for errors.
