# GitHub Metrics Aggregator

GitHub Metrics Aggregator (GMA) is a GitHub app that ingests events from the GitHub API and creates dashboards about velocity and productivity.

It is made up of two components, webhook service and retry service. The webhook service ingests GitHub webhook event payloads. This service will post all requests to a PubSub topic for ingestion and aggregation into BigQuery. The retry service will run on a configurable cadence and redeliver events that failed to process by the webhook service.

## Architecture

!["Architecture"](./assets/architecture.svg)

## Setup

### What to expect

We recommend using the abc CLI to render templates for setting up GMA. The setup
is split into three parts:

1. Provision infrastructure with Terraform
2. Add secret values via Secret Manager
3. Build and deploy the service with GitHub workflows

```
.github/
  workflows/
    deploy-github-metrics.yaml
github-metrics/
  infra/
    main.tf
    outputs.tf
    terraform.tf
  deployments/
    deploy.sh
```

### Pre-requisites
- abc CLI version >= v0.5.0
- a Google Artifact Registry repository (for uploading GMA images)
- a custom domain with access to update DNS
- access to create and manage a GitHub App
- a WIF service account for deploying to the Cloud Run services
- a Google Cloud Storage bucket for storing the Terraform state remotely

### Create a GitHub App

Follow the directions from these [GitHub instructions](https://docs.github.com/en/apps/creating-github-apps/setting-up-a-github-app/creating-a-github-app#creating-a-github-app). Uncheck everything and provide all required fields that remain. Make sure to uncheck the Active checkbox within the Webhook section so you don't have to supply a webhook yet, it will be created when you deploy the Terraform module in the next section. Create a private key and download it for an upcoming step. Once the GitHub App is created, take note of the GitHub App ID.

#### Grant GitHub App permissions
Grant any of the following permissions (or more) according to your requirements:
- **Repository Permissions**
  - Pull Requests - Read Only
- **Subscribe to Events**
  - Check the Pull Request box
  - Check the Pull Request Review box

## Provision the infrastructure

Run the following command after replacing the input values.

```shell
abc templates render \
  -input=custom_name=GMA-CUSTOM-NAME \
  -input=project_id=GMA-PROJECT-ID \
  -input=automation_service_account_email=CI-SERVICE-ACCOUNT \
  -input=domain=GMA-DOMAIN \
  -input=terraform_state_bucket=TERRAFORM-BUCKET-NAME \
  -input=github_app_id=GMA-GITHUB-APP-ID \
  github.com/abcxyz/github-metrics-aggregator/abc.templates/infra@v0.0.24
```

This should render the following Terraform files.

```
GMA-CUSTOM-NAME/
  infra/
    main.tf
    outputs.tf
    terraform.tf
```

Run Terraform init:

```shell
terraform -chdir=GMA-CUSTOM-NAME/infra init -backend=false
```

Then apply the Terraform. Take note of the following values from the generated
output:
- `webhook_run_service.service_name`
- `retry_run_service.service_name`
- `gclb_external_ip_address`

### Update DNS
Create an `A` record pointing your custom domain to the `gclb_external_ip_address`.

### Update the GitHub App
In the GitHub App settings,
1. Check the Active checkbox in the Webhook section
2. Set the webhook URL: `https://GMA-DOMAIN/webhook`
3. Save changes

## Add Secret Values to Secret Manager

### Create webhook secret

Run the following command to generate a random string to be use for the Github Webhook secret

```shell
openssl rand -base64 32
```

Save this value for the ["Upload the secrets"](#upload-the-secrets) step.

The Terraform module will have created a Secret Manager secret in the project provided with the name `github-webhook-secret`.

### Create Github private key secret

The Terraform module will have created a Secret Manager secret in the project provided with the name `github-private-key`.

In the GitHub App settings, under the Private Keys section,
1. Click Generate a private key. This will add a `.pem` file to your Downloads.
The contents of the .pem file is the private_key, including the BEGIN and END.
It should look something like,

    ```
    -----BEGIN RSA PRIVATE KEY-----
    SOME-SUPER-SECRET-
    SHHHHHHHHHHHHHH-
    KEEP-THIS-A-SECRET
    -----END RSA PRIVATE KEY-----
    ```

Copy the contents of the file:
```shell
cat location/to/private/key.private-key.pem | pbcopy
```

### Upload the secrets
Navigate to the Google Cloud dashboard for Secret Manager and add a new revision
with the generated values to their corresponding secret ID's.
- `github-webhook-secret`
- `github-private-key`.

## Deploy the service

**NOTE:** Before going through the following steps, ensure that your GMA Cloud
Run service agent can read from your GAR repository. The service agent is in the
form of `service-GMA-PROJECT-NUM@serverless-robot-prod.iam.gserviceaccount.com`

Run the following command after replacing the input values.

```shell
abc templates render \
  -input=wif_provider=CI-WIF-PROVIDER \
  -input=wif_service_account=CI-SERVICE-ACCOUNT \
  -input=project_id=GMA-PROJECT-ID \
  -input=full_image_name=us-docker.pkg.dev/GAR-PROJECT-ID/GAR-REPOSITORY/gma-server \
  -input=region=REGION \
  -input=webhook_service_name=GMA-WEBHOOK-SERVICE-NAME \
  -input=retry_service_name=GMA-RETRY-SERVICE-NAME \
  -input=custom_name=GMA-CUSTOM-NAME \
  github.com/abcxyz/github-metrics-aggregator/abc.templates/deployments@v0.0.24
```

This should generate the following files:

```
.github/
  workflows/
    deploy-GMA-CUSTOM-NAME.yaml
GMA-CUSTOM-NAME/
  deployments/
    deploy.sh
```

Merge these files into the `main` branch of your repository. This should trigger
the `deploy-GMA-CUSTOM-NAME.yaml` workflow to build and upload your GMA image to
GAR and then deploy to the Cloud Run services.

You can alternatively manually run the workflow, if necessary.

## Looker Studio

### Template Dashboard
abcxyz provides a template Looker Studio Dashboard. To utilize this, add the
following config in the `GMA-CUSTOM-NAME/infra/main.tf` file.

```terraform
module "GMA_CUSTOM_NAME" {
  # ...hidden properties
  # ...

  github_metrics_dashboard = {
      enabled = true # set this to true (defaults to false)
      viewers = [] # add viewers, such as "group:<group-email>",
  }
}
```

After applying these changes with Terraform, copy the value of
`github_metrics_looker_studio_report_link` from the output values and navigate
to the link in your browser.

This will give you a preview of the dashboard. On the top right, click Edit and Share.
Verify the data, then proceed to save. This will complete the process to link your datasource to the Looker Studio report template.

### Custom Dashboard
To make use of the events data, it is recommended to create views per event. This allows you to create Looker Studio data sources per event that can be used in dashboard.

#### Example

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
- `FAILURE_EVENTS_TABLE_ID`: (Required) The failure event table ID.
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
- `GITHUB_ENTERPRISE_SERVER_URL`: (Optional) The GitHub Enterprise server URL if available, format \"https://[hostname]\".

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
