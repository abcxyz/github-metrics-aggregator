# Copyright 2024 The Authors (see AUTHORS file)
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
  playbook_prefix   = "https://github.com/abcxyz/github-metrics-aggregator/blob/main/docs/playbooks/alerts"
  job_metric_prefix = "run.googleapis.com/job"

  second = 1
  minute = 60 * local.second
  hour   = 60 * local.minute
  day    = 24 * local.hour

  forward_progress_job_indicators = merge(
    {
      # artifacts job runs every 15m, alert after 3 failures + buffer
      "gma-artifacts" = { metric = "completed_execution_count", window = 45 * local.minute + 5 * local.minute },

      # review status job runs every 4h, alert after 2 failures + buffer
      "commit-review-status-job" = { metric = "completed_execution_count", window = 8 * local.hour + 10 * local.minute },
    },
    var.forward_progress_job_indicators,
  )
}

resource "google_monitoring_alert_policy" "job_alert_policy" {
  for_each = local.forward_progress_job_indicators

  project = var.project_id

  display_name = "ForwardProgress-${each.key}"
  combiner     = "OR"

  alert_strategy {
    auto_close = "${local.day}s"

    notification_channel_strategy {
      renotify_interval = "${local.day}s"
    }
  }

  conditions {
    display_name = "${each.key} failing"

    condition_threshold {
      filter   = "metric.type = \"${local.job_metric_prefix}/${each.value.metric}\" AND resource.type = \"cloud_run_job\" AND resource.label.\"job_name\"=\"${each.key}\""
      duration = "${each.value.window}s"

      comparison      = "COMPARISON_LT"
      threshold_value = 1

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_DELTA"
        group_by_fields      = ["resource.labels.job_name"]
        cross_series_reducer = "REDUCE_SUM"
      }

      trigger {
        count = 1
      }
    }
  }

  conditions {
    display_name = "${each.key} missing"

    condition_absent {
      filter   = "metric.type = \"${local.job_metric_prefix}/${each.value.metric}\" AND resource.type = \"cloud_run_job\" AND resource.label.\"job_name\"=\"${each.key}\""
      duration = "${each.value.window}s"

      aggregations {
        alignment_period     = "60s"
        per_series_aligner   = "ALIGN_DELTA"
        group_by_fields      = ["resource.labels.job_name"]
        cross_series_reducer = "REDUCE_SUM"
      }
    }
  }

  documentation {
    content   = "${local.playbook_prefix}/ForwardProgressFailed.md"
    mime_type = "text/markdown"
  }

  notification_channels = [for x in values(google_monitoring_notification_channel.non_paging) : x.id]
}
