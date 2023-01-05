/**
 * Copyright 2023 The Authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

resource "google_project_service" "services" {
  project = var.project_id
  for_each = toset([
    "cloudresourcemanager.googleapis.com",
    "compute.googleapis.com",
    "run.googleapis.com",
    "secretmanager.googleapis.com"
  ])
  service            = each.value
  disable_on_destroy = false
}

resource "google_service_account" "default" {
  project      = var.project_id
  account_id   = "${var.name}-sa"
  display_name = "${var.name} service account"
}

// GitHub webhook shared secret
resource "google_secret_manager_secret" "default" {
  project   = var.project_id
  secret_id = "${var.name}-secret"
  replication {
    automatic = true
  }
  depends_on = [
    google_project_service.services["secretmanager.googleapis.com"],
  ]
}

resource "google_secret_manager_secret_iam_member" "default" {
  project    = var.project_id
  secret_id  = google_secret_manager_secret.default.id
  role       = "roles/secretmanager.secretAccessor"
  member     = "serviceAccount:${google_service_account.default.email}"
  depends_on = [google_secret_manager_secret.default]
}

resource "google_project_iam_member" "observability" {
  for_each = toset([
    "roles/cloudtrace.agent",
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/stackdriver.resourceMetadata.writer",
  ])

  project = var.project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.default.email}"
}

resource "google_cloud_run_service" "default" {
  project  = var.project_id
  name     = var.name
  location = var.region

  autogenerate_revision_name = true

  metadata {
    annotations = {
      "run.googleapis.com/ingress" : var.ingress
      "run.googleapis.com/launch-stage" : "BETA"
    }
  }

  template {
    spec {
      service_account_name = google_service_account.default.email

      containers {
        image = "gcr.io/cloudrun/placeholder@sha256:45ec0903b0fe39eb9e4a1cca42f1719e4fe3fda2d5a0f49cc355d453a60f1ff5"

        resources {
          limits = {
            cpu    = "2000m"
            memory = "1G"
          }
        }
      }
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" : "10",
        "run.googleapis.com/sandbox" : "gvisor",
      }
    }
  }

  depends_on = [
    google_project_service.services["run.googleapis.com"],
    google_secret_manager_secret_iam_member.default,
  ]

  lifecycle {
    ignore_changes = [
      metadata[0].annotations["client.knative.dev/user-image"],
      metadata[0].annotations["run.googleapis.com/client-name"],
      metadata[0].annotations["run.googleapis.com/client-version"],
      metadata[0].annotations["run.googleapis.com/ingress-status"],
      metadata[0].annotations["run.googleapis.com/launch-stage"],
      metadata[0].annotations["serving.knative.dev/creator"],
      metadata[0].annotations["serving.knative.dev/lastModifier"],
      metadata[0].labels["cloud.googleapis.com/location"],
      template[0].metadata[0].annotations["client.knative.dev/user-image"],
      template[0].metadata[0].annotations["run.googleapis.com/client-name"],
      template[0].metadata[0].annotations["run.googleapis.com/client-version"],
      template[0].metadata[0].annotations["run.googleapis.com/sandbox"],
      template[0].metadata[0].annotations["serving.knative.dev/creator"],
      template[0].metadata[0].annotations["serving.knative.dev/lastModifier"],
      template[0].spec[0].containers[0].image,
      template[0].spec[0].containers[0].env,
    ]
  }
}

resource "google_cloud_run_service_iam_member" "default" {
  project  = google_cloud_run_service.default.project
  service  = google_cloud_run_service.default.name
  location = google_cloud_run_service.default.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}


module "lb_http" {
  source  = "GoogleCloudPlatform/lb-http/google//modules/serverless_negs"
  version = "~> 6.3"
  name    = "${var.name}-lb"
  project = var.project_id

  ssl                             = true
  managed_ssl_certificate_domains = [var.domain]
  https_redirect                  = true

  backends = {
    "${var.name}" = {
      description = null
      groups = [
        {
          group = google_compute_region_network_endpoint_group.default.id
        }
      ]
      enable_cdn              = false
      security_policy         = null
      custom_request_headers  = null
      custom_response_headers = null

      iap_config = {
        enable               = false
        oauth2_client_id     = ""
        oauth2_client_secret = ""
      }
      log_config = {
        enable      = false
        sample_rate = null
      }
    }
  }

  depends_on = [google_project_service.services["compute.googleapis.com"]]
}

resource "google_compute_region_network_endpoint_group" "default" {
  provider              = google-beta
  project               = var.project_id
  region                = var.region
  name                  = "${var.name}-neg"
  network_endpoint_type = "SERVERLESS"
  cloud_run {
    service = google_cloud_run_service.default.name
  }

  depends_on = [google_project_service.services["compute.googleapis.com"]]
}
