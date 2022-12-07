#!/bin/bash

PROJECT_ID="github-metrics-dev"
REGION="us-central1"
SERVICE="github-webhook"
TOPIC_ID="github-webhook"
WEBHOOK_SECRET="github-webhook-secret:latest"

gcloud run deploy $SERVICE --source . \
    --project=$PROJECT_ID \
    --region=$REGION \
    --set-env-vars="TOPIC_ID=${TOPIC_ID}" \
    --set-secrets="WEBHOOK_SECRET=${WEBHOOK_SECRET}" \
    --service-account="github-webhook-sa@github-metrics-dev.iam.gserviceaccount.com" \
    --allow-unauthenticated