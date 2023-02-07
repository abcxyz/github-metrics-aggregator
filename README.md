# GitHub Metrics Aggregator

Service to ingest GitHub webhook event payloads. This service will post all requests to a PubSub topic for ingestion and aggregation into BigQuery.

## Architecture

!["Architecture"](./assets/architecture.svg)

## Setup

### Deploy the service

You can use the provided terraform module to setup the basic infrastructure needed for this service. Otherwise you can refer to the provided module to see how to build your own terraform from scratch.

```terraform
module "github_metrics_aggregator" {
  source     = "git::https://github.com/abcxyz/github-metrics-aggregator.git//terraform?ref=main"
  project_id = "YOUR_PROJECT_ID"
  name       = "github-events-webhook"
  image      = "us-docker.pkg.dev/abcxyz-artifacts/docker-images/github-metrics-aggregator-server:v0.0.1-amd64"
  domain     = "github-events-webhook.domain.com"
  service_iam = {
    admins     = []
    developers = []
    invokers   = ["allUsers"] # public access, called by github webhook
  }
  dataset_id = "github_webhook"
  table_id   = "events"
  dataset_iam = {
    owners  = ["group:your-owner-group@domain.com"]
    editors = []
    viewers = ["group:your-viewer-group@domain.com"]
  }
}
```

Additionally, if you plan to use Looker Studio to visualize this data, you will need to add the users to run BigQuery Jobs with the following IAM role

```terraform
resource "google_project_iam_member" "bigquery_jobusers" {
  for_each = toset([
    "group:your-owner-group@domain.com",
    "group:your-viewer-group@domain.com"
  ])

  project = "YOUR_PROJECT_ID"
  role    = "roles/bigquery.jobUser"
  member  = each.value
}
```

**NOTE: You will also need to provide access to Looker Studio via sharing the dashboard as well**

## Create webhook secret

Run the following command to generate a random string to be use for the Github Webhook secret

```shell
openssl rand -base64 32
```

Save this value for the next step.

The terraform moule will create a Secret Manager secret in the project provided with the name `github-webhook-secret`. Navigate to the Google Cloud dashboard for Secret Manager and add a new revision with this generated value.

**NOTE: Before continuing, you may want to replace your CloudRun service to ensure it picks up the latest version of the secret**

```shell
terraform apply \
  -replace=module.github_metrics_aggregator.module.cloud_run.google_cloud_run_service.service
```

## Create organization webhook

- Navigate to your organization home page and click `Settings` in the top right
- In the left menu bar, click `Webhooks`
- Click `Add webhook` button in the top right
- For the Payload URL, enter your domain for the deployed service, with the suffix `/webhook`
- For Content type select `application/json`
- For secret, enter the value created above
- Select the `Let me select individual events.` option
  - Choose the events you want to recieve
- Leave the `Active` setting checked
  - You can uncheck this, but no events will be sent until this is checked

## Looker Studio

To make use of the events data, it is recommended to create views per event. This allows you to create Looker Studio data sources per event that can be used in dashboard.

### Example

```sql

SELECT
  received,
  event,
  JSON_VALUE(payload, "$.organization.login") owner,
  JSON_VALUE(payload, "$.organization.id") owner_id,
  JSON_VALUE(payload, "$.repository.name") repo,
  JSON_VALUE(payload, "$.repository.id") repo_id,
  JSON_VALUE(payload, "$.repository.full_name") repo_full_name,
  JSON_VALUE(payload, "$.repository.visibility") repo_visibility,
  JSON_VALUE(payload, "$.sender.login") sender,
  JSON_VALUE(payload, "$.action") action,
  JSON_VALUE(payload, "$.pull_request.id") id,
  JSON_VALUE(payload, "$.pull_request.title") title,
  JSON_VALUE(payload, "$.pull_request.state") state,
  JSON_VALUE(payload, "$.pull_request.url") url,
  JSON_VALUE(payload, "$.pull_request.html_url") html_url,
  JSON_VALUE(payload, "$.pull_request.base.ref") base_ref,
  JSON_VALUE(payload, "$.pull_request.head.ref") head_ref,
  JSON_VALUE(payload, "$.pull_request.user.login") author,
  JSON_VALUE(payload, "$.pull_request.user.id") author_id,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.created_at")) created_at,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.closed_at")) closed_at,
  JSON_VALUE(payload, "$.pull_request.merged") merged,
  JSON_VALUE(payload, "$.pull_request.merge_commit_sha") merge_commit,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.merged_at")) merged_at,
  TIMESTAMP(JSON_VALUE(payload, "$.pull_request.merged_by")) merged_by,
  TIMESTAMP_DIFF(TIMESTAMP(JSON_VALUE(payload, "$.pull_request.closed_at")), TIMESTAMP(JSON_VALUE(payload, "$.pull_request.created_at")), SECOND) open_duration_s,
  PARSE_JSON(payload) payload
FROM
  `YOUR_PROJECT_ID.github_webhook.events`
WHERE
  event = "pull_request";
```

## Testing Locally

### Creating GitHub HMAC Signature

```bash
echo -n `cat testdata/issues.json` | openssl sha256 -hmac "test-secret"

# Output:
08a88fe31f89ab81a944e51e51f55ebf9733cb958dd83276040fd496e5be396a
```

Use this value in the `X-Hub-Signature-256` request header as follows:

```bash
X-Hub-Signature-256: sha256=08a88fe31f89ab81a944e51e51f55ebf9733cb958dd83276040fd496e5be396a
```

### Example Request

```bash
PAYLOAD=$(echo -n `cat testdata/issues.json`)
WEBHOOK_SECRET="test-secret"

curl \
  -H "Content-Type: application/json" \
  -H "X-Github-Delivery: $(uuidgen)" \
  -H "X-Github-Event: issues" \
  -H "X-Hub-Signature-256: sha256=$(echo -n $PAYLOAD | openssl sha256 -hmac $WEBHOOK_SECRET)" \
  -d $PAYLOAD \
  http://localhost:8080/webhook

# Output
Ok
```
