# Server Faults

## Webhook Service Faults High

This alert fires when the webhook Cloud Run service is experiencing server faults. The alert policy monitors the `request-count` metric and checks for the response code class 5xx.

Event failures emitted by the webhook service are later processed by the retry service. This means users do not have to manually reattempt event delivery. 

### Triage Steps

1. Navigate to Log Explorer and set the date range to match the range when the error was observed.
2. Query for `severity=ERROR jsonPayload.code=~"5[0-9][0-9]+" resource.type="cloud_run_revision resource.labels.service_name=~"github-metrics-webhook-....""`
3. Write failures to BigQuery will be retried, no further action to take.


## Retry Service Faults High

This alert fires when the retry Cloud Run service is experiencing server faults. The alert policy monitors the `request-count` metric and checks for the response code class 5xx.

### Triage Steps

1. Navigate to Log Explorer and set the date range to match the range when the error was observed.
2. Query for `severity=ERROR jsonPayload.code=~"4[0-9][0-9]+" resource.type="cloud_run_revision resource.labels.service_name=~"github-metrics-retry-....""`
3. Review logs for recent runs to identify the failure.
