# Copyright 2023 The Authors (see AUTHORS file)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

locals {
  run_service_name = "${substr(var.name, 0, 58)}-${random_id.default.hex}" # 63 character limit

  default_run_revision_annotations = {
    "autoscaling.knative.dev/minScale" : var.min_instances,
    "autoscaling.knative.dev/maxScale" : var.max_instances,
    "run.googleapis.com/sandbox" : "gvisor",
    "run.googleapis.com/execution-environment" : var.execution_environment
  }

  default_run_service_annotations = {
    "run.googleapis.com/ingress" : var.ingress
  }

  default_run_envvars = {}

  default_run_secret_envvars = {}

  default_run_secret_volumes = {}
}

resource "random_id" "default" {
  byte_length = 2
}

resource "google_project_service" "services" {
  for_each = toset([
    "cloudresourcemanager.googleapis.com",
    "monitoring.googleapis.com",
    "run.googleapis.com",
    "secretmanager.googleapis.com",
    "serviceusage.googleapis.com",
    "stackdriver.googleapis.com",
  ])

  project = var.project_id

  service                    = each.value
  disable_on_destroy         = false
  disable_dependent_services = false
}

resource "google_cloud_run_service" "service" {
  project = var.project_id

  name                       = local.run_service_name
  location                   = var.region
  autogenerate_revision_name = true

  metadata {
    annotations = local.default_run_service_annotations
  }


  template {
    spec {
      service_account_name = var.service_account_email
      containers {
        image = var.image
        args  = var.args

        # TODO: Implement tcp_socket and grpc configuration blocks
        dynamic "startup_probe" {
          for_each = var.startup_probe == null ? [] : [""]

          content {
            initial_delay_seconds = var.startup_probe.initial_delay_seconds
            timeout_seconds       = var.startup_probe.timeout_seconds
            period_seconds        = var.startup_probe.period_seconds
            failure_threshold     = var.startup_probe.failure_threshold

            dynamic "http_get" {
              for_each = var.startup_probe.http_get == null ? [] : [""]

              content {
                path = var.startup_probe.http_get.path
                port = var.startup_probe.http_get.port

                dynamic "http_headers" {
                  for_each = (
                    var.startup_probe.http_get.http_headers
                  )
                  content {
                    name  = http_headers.key
                    value = http_headers.value
                  }
                }
              }
            }
          }
        }

        resources {
          requests = var.resources.requests
          limits   = var.resources.limits
        }

        dynamic "env" {
          for_each = merge(local.default_run_envvars, var.envvars)

          content {
            name  = env.key
            value = env.value
          }
        }

        dynamic "env" {
          for_each = merge(local.default_run_secret_envvars, var.secret_envvars)

          content {
            name = env.key
            value_from {
              secret_key_ref {
                key  = env.value.version
                name = env.value.name
              }
            }
          }
        }

        dynamic "volume_mounts" {
          for_each = merge(local.default_run_secret_volumes, var.secret_volumes)
          content {
            mount_path = volume_mounts.key
            name       = volume_mounts.value.name
          }
        }
      }

      dynamic "volumes" {
        for_each = merge(local.default_run_secret_volumes, var.secret_volumes)
        content {
          name = volumes.value.name
          secret {
            secret_name = volumes.value.name
            items {
              key  = volumes.value.version
              path = volumes.value.name
            }
          }
        }
      }
    }

    metadata {
      annotations = merge(local.default_run_revision_annotations,
        var.additional_revision_annotations,
      )
    }
  }

  depends_on = [
    google_project_service.services["run.googleapis.com"],
    google_secret_manager_secret_iam_member.secrets_accessors_iam,
  ]
  lifecycle {
    ignore_changes = [
      metadata[0].annotations["client.knative.dev/user-image"],
      metadata[0].annotations["run.googleapis.com/client-name"],
      metadata[0].annotations["run.googleapis.com/client-version"],
      metadata[0].annotations["run.googleapis.com/ingress-status"],
      metadata[0].annotations["run.googleapis.com/launch-stage"],
      metadata[0].annotations["run.googleapis.com/operation-id"],
      metadata[0].annotations["serving.knative.dev/creator"],
      metadata[0].annotations["serving.knative.dev/lastModifier"],
      metadata[0].labels["cloud.googleapis.com/location"],
      metadata[0].labels["commit-sha"],
      metadata[0].labels["managed-by"],
      template[0].metadata[0].annotations["client.knative.dev/user-image"],
      template[0].metadata[0].annotations["run.googleapis.com/client-name"],
      template[0].metadata[0].annotations["run.googleapis.com/client-version"],
      template[0].metadata[0].annotations["run.googleapis.com/sandbox"],
      template[0].metadata[0].annotations["serving.knative.dev/creator"],
      template[0].metadata[0].annotations["serving.knative.dev/lastModifier"],
      template[0].metadata[0].labels["run.googleapis.com/startupProbeType"],
      template[0].metadata[0].labels["client.knative.dev/nonce"],
      template[0].metadata[0].labels["serving.knative.dev/nonce"],
      template[0].metadata[0].labels["sha"],
      template[0].spec[0].containers[0].image,
    ]
  }
}

resource "google_cloud_run_service_iam_binding" "admins" {
  project = google_cloud_run_service.service.project

  location = google_cloud_run_service.service.location
  service  = google_cloud_run_service.service.name
  role     = "roles/run.admin"
  members  = toset(var.service_iam.admins)
}

resource "google_cloud_run_service_iam_binding" "invokers" {
  project = google_cloud_run_service.service.project

  location = google_cloud_run_service.service.location
  service  = google_cloud_run_service.service.name
  role     = "roles/run.invoker"
  members  = toset(var.service_iam.invokers)
}

resource "google_cloud_run_service_iam_binding" "developers" {
  project = google_cloud_run_service.service.project

  location = google_cloud_run_service.service.location
  service  = google_cloud_run_service.service.name
  role     = "roles/run.developer"
  members  = toset(var.service_iam.developers)
}

resource "google_project_iam_member" "run_observability_iam" {
  for_each = toset([
    "roles/cloudtrace.agent",
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/stackdriver.resourceMetadata.writer",
  ])

  project = var.project_id

  role   = each.key
  member = "serviceAccount:${var.service_account_email}"

  depends_on = [
    google_project_service.services["iam.googleapis.com"],
  ]
}

# Secret Manager secrets for the Cloud Run service to use
resource "google_secret_manager_secret" "secrets" {
  for_each = toset(var.secrets)

  project = var.project_id

  secret_id = each.value
  replication {
    auto {}
  }

  depends_on = [
    google_project_service.services["secretmanager.googleapis.com"]
  ]
}

resource "google_secret_manager_secret_version" "secrets_default_version" {
  for_each = toset(var.secrets)

  secret = google_secret_manager_secret.secrets[each.key].id
  # default value used for initial revision to allow cloud run to map the secret
  # to manage this value and versions, use the google cloud web application
  secret_data = "DEFAULT_VALUE"

  lifecycle {
    ignore_changes = [
      enabled
    ]
  }
}

resource "google_secret_manager_secret_iam_member" "secrets_accessors_iam" {
  for_each = toset(var.secrets)

  project = var.project_id

  secret_id = google_secret_manager_secret.secrets[each.key].id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.service_account_email}"
}
