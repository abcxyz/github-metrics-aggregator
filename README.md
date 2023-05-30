# GitHub Metrics Aggregator

GitHub Metrics Aggregator (GMA) is made up of two components, webhook service and retry service. The webhook service ingests GitHub webhook event payloads. This service will post all requests to a PubSub topic for ingestion and aggregation into BigQuery. The retry service will run on a configurable cadence and redeliver events that failed to process by the webhook service.

## Architecture

!["Architecture"](./assets/architecture.svg)

## Setup

## Create a GitHub App

Follow the directions from these [GitHub instructions](https://docs.github.com/en/apps/creating-github-apps/setting-up-a-github-app/creating-a-github-app#creating-a-github-app). Uncheck everything and provide all required fields that remain. Make sure to uncheck the Active checkbox within the Webhook section so you don't have to supply a webhook yet, it will be created when you deploy the terraform module in the next section. Create a private key and download it for an upcoming step. Once the GitHub App is created, take note of the GitHub App ID.

### Deploy the service

You can use the provided terraform module to setup the infrastructure needed for GMA. Otherwise you can refer to the provided module to see how to build your own terraform from scratch.

```terraform
module "github_metrics_aggregator" {
  source               = "git::https://github.com/abcxyz/github-metrics-aggregator.git//terraform?ref=main" # this should be pinned to the SHA desired
  project_id           = "YOUR_PROJECT_ID"
  image                = "us-docker.pkg.dev/abcxyz-artifacts/docker-images/github-metrics-aggregator:v0.0.1-amd64" # versions exist for releases for both *-amd64 and *-arm64
  big_query_project_id = "PROJECT_ID_FOR_BIG_QUERY_INFRA" # this can be the same as the project_id
  webhook_domains      = [<YOUR_WEBHOOK_DOMAIN> i.e. "github-events-webhook.domain.com"]
  github_app_id        = "<YOUR_GITHUB_APP_ID>"
  leech_bucket_name = "<YOUR_PROJECT_ID>_leech_logs"
  retry_service_iam = {
    owners  = []
    editors = []
    viewers = []
  }
  webhook_service_iam = {
    admins     = []
    developers = []
    invokers   = ["allUsers"] # public access, called by github webhook
  }
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

## Create a webhook within your GitHub App

Now that you have deployed your terraform module, your webhook endpoint is up and running. Grab the URL from Cloud Run (or the DNS name if you configured one) and edit your GitHub App. Go to the Webhook section and check the box next to Active and supply the webhook endpoint i.e. `<your_endpoint>/webhook` as the service is listening expecting traffic on the `/webhook` route.

## Create webhook secret

Run the following command to generate a random string to be use for the Github Webhook secret

```shell
openssl rand -base64 32
```

Save this value for the next step.

The terraform module will create a Secret Manager secret in the project provided with the name `github-webhook-secret`. Navigate to the Google Cloud dashboard for Secret Manager and add a new revision with this generated value. Note there will be another secret (`github-private-key`) in Secret Manager that will be addressed in the next section.

## Create github private key secret

The terraform module will create a Secret Manager secret in the project provided with the name `github-private-key`. Convert your downloaded key from when you created your GitHub App and convert it into a string using the following comman in your terminal. If you didn't create a key earlier when creating your GitHub App or lost your key, go to your GitHub App and create a new one and delete older keys.

```shell
cat location/to/private/key.private-key.pem | pbcopy
```

Navigate to the Google Cloud dashboard for Secret Manager and add a new revision with this generated value to `github-private-key`. 

**NOTE: Before continuing, you may want to replace your Cloud Run services to ensure it picks up the latest version of the secret**

```shell
terraform apply \
  -replace=module.github_metrics_aggregator.module.webhook_cloud_run.google_cloud_run_service.service
```

```shell
terraform apply \
  -replace=module.github_metrics_aggregator.module.retry_cloud_run.google_cloud_run_service.service
```

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

## Environment Variables

### Webhook Service

- `BIG_QUERY_PROJECT_ID`: (Optional) The project ID where your BigQuery instance exists in. Defaults to the `PROJECT_ID`. 
- `DATASET_ID`: (Required) The dataset ID within the BigQuery instance.
- `EVENTS_TABLE_ID`: (Required) The event table ID.
- `FAILURE_EVENTS_TABLE_ID`: (Required) The falure event table ID.
- `PORT`: (Optional) The port where the webhook service will run on. Defaults to 8080.
- `PROJECT_ID`: (Required) The project where the webhook service exists in.
- `RETRY_LIMIT`: (Required) The number of retry attempts to make for failed GitHub event before writing to the DLQ.
- `EVENTS_TOPIC_ID`: (Required) The topic ID for PubSub.
- `DLQ_EVENTS_TOPIC_ID`: : (Required) The topic ID for PubSub DLQ where exhausted events are written.
- `GITHUB_WEBHOOK_SECRET`: Used to decrypt the payload from the webhook events.

### Retry Service

- `GITHUB_APP_ID`: (Required) The provisioned GitHub App reference.
- `GITHUB_PRIVATE_KEY`: (Required) A PEM encoded string representation of the GitHub App's private key.
- `BIG_QUERY_PROJECT_ID`: (Optional) The project ID where your BigQuery instance exists in. Defaults to the `PROJECT_ID`.
- `BUCKET_NAME`: (Required) The name of the bucket that holds the lock to enforce synchronous processing of the retry service.
- `CHECKPOINT_TABLE_ID`: (Required) The checkpoint table ID.
- `EVENTS_TABLE_ID`: (Required) The event table ID.
- `DATASET_ID`: (Required) The dataset ID within the BigQuery instance.
- `LOCK_TTL_CLOCK_SKEW`: (Optional) Duration to account for clock drift when considering the `LOCK_TTL`. Defaults to 10s.
- `LOCK_TTL`: (Optional) Duration for a lock to be active until it is allowed to be taken. Defaults to 5m.
- `PROJECT_ID`: (Required) The project where the retry service exists in.
- `PORT`: (Optional) The port where the retry service will run on. Defaults to 8080.
- `LOG_MODE`: (Required) The mode for logs. Defaults to production.
- `LOG_LEVEL`: (Required) The level for logging. Defaults to warning.

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
GITHUB_WEBHOOK_SECRET="test-secret"

curl \
  -H "Content-Type: application/json" \
  -H "X-Github-Delivery: $(uuidgen)" \
  -H "X-Github-Event: issues" \
  -H "X-Hub-Signature-256: sha256=$(echo -n $PAYLOAD | openssl sha256 -hmac $GITHUB_WEBHOOK_SECRET)" \
  -d $PAYLOAD \
  http://localhost:8080/webhook

# Output
Ok
```
