#!/usr/bin/env bash
set -euo pipefail

GITHUB_METRICS_VERSION='REPLACE_GITHUB_METRICS_VERSION_TAG'
PLATFORM='amd64'
VERSION="${GITHUB_METRICS_VERSION}-${PLATFORM}"
UPSTREAM_IMAGE_NAME="us-docker.pkg.dev/abcxyz-artifacts/docker-images/github-metrics-aggregator:${VERSION}"
IMAGE_NAME="REPLACE_FULL_IMAGE_NAME:${GITHUB_METRICS_VERSION}-${GITHUB_SHA}"

echo "Copying ${UPSTREAM_IMAGE_NAME} to ${IMAGE_NAME}..."
crane copy "${UPSTREAM_IMAGE_NAME}" "${IMAGE_NAME}"

gcloud run services update "${WEBHOOK_SERVICE_NAME}" \
  --quiet \
  --project="${PROJECT_ID}" \
  --region="${REGION}" \
  --image="${IMAGE_NAME}"

gcloud run services update "${RETRY_SERVICE_NAME}" \
  --quiet \
  --project="${PROJECT_ID}" \
  --region="${REGION}" \
  --image="${IMAGE_NAME}"
