# Bad Requests

## Webhook Service Bad Requests High

This alert fires when the webhook Cloud Run service is experiencing a high level of bad requests. The alert policy monitors the `request-count` metric and checks for the response code class 4xx.

The service throws a 400 when the payload from GitHub is empty which is unexpected.

### Triage Steps

1. Navigate to Log Explorer and set the date range to match the range when the error was observed.
2. Query for `severity=ERROR jsonPayload.code=~"4[0-9][0-9]+" resource.type="cloud_run_revision resource.labels.service_name=~"github-metrics-webhook-....""`

## Retry Service Bad Requests High

This alert fires when the retry Cloud Run service is experiencing a high level of bad requests. The alert policy monitors the `request-count` metric and checks for the response code class 4xx.

### Triage Steps

1. Navigate to Log Explorer and set the date range to match the range when the error was observed.
2. Query for `severity=ERROR jsonPayload.code=~"4[0-9][0-9]+" resource.type="cloud_run_revision resource.labels.service_name=~"github-metrics-retry-....""`
