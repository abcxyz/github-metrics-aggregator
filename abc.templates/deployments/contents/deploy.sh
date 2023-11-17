#!/usr/bin/env bash
set -euo pipefail

GITHUB_METRICS_VERSION='v0.0.2'
PLATFORM='amd64'
IMAGE_NAME="REPLACE_FULL_IMAGE_NAME:${GITHUB_METRICS_VERSION}-${GITHUB_SHA}"

docker build -t "${IMAGE_NAME}" \
  --build-arg="VERSION=${GITHUB_METRICS_VERSION}-${PLATFORM}" \
  REPLACE_SUBDIRECTORY

docker push "${IMAGE_NAME}"

gcloud run services update "${WEBHOOK_SERVICE_NAME}" \
  --project="${PROJECT_ID}" \
  --region="${REGION}" \
  --image="${IMAGE_NAME}"

gcloud run services update "${RETRY_SERVICE_NAME}" \
  --project="${PROJECT_ID}" \
  --region="${REGION}" \
  --image="${IMAGE_NAME}"
