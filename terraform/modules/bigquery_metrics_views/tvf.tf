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

# Table Value Functions. Stored here so inputs can be modeled in HCL
# rather than raw SQL statements.

# Version of pull_request_events that uses TVF
resource "google_bigquery_routine" "pull_request_events_by_date" {
  project = var.project_id

  dataset_id      = var.dataset_id
  routine_id      = "unique_events_by_date_type"
  routine_type    = "TABLE_VALUED_FUNCTION"
  language        = "SQL"
  definition_body = templatefile("${path.module}/data/bq_tvf/events/pull_request_events_by_date.sql",
    {
      project    = var.project_id,
      dataset_id = var.dataset_id,
      routine_id = var.base_tvf_id,
    }
  )

  arguments {
    name      = "startTimestamp"
    data_type = jsonencode({ typeKind : "TIMESTAMP" })
  }
  arguments {
    name      = "endTimestamp"
    data_type = jsonencode({ typeKind : "TIMESTAMP" })
  }
}

# Version of pull_requests that uses TVF
resource "google_bigquery_routine" "pull_requests_by_date" {
  project = var.project_id

  dataset_id      = var.dataset_id
  routine_id      = "unique_events_by_date_type"
  routine_type    = "TABLE_VALUED_FUNCTION"
  language        = "SQL"
  definition_body = templatefile("${path.module}/data/bq_tvf/resources/pull_requests_by_date.sql",
    {
      project    = google_bigquery_routine.pull_request_events_by_date.project,
      dataset_id = google_bigquery_routine.pull_request_events_by_date.dataset_id,
      routine_id = google_bigquery_routine.pull_request_events_by_date.routine_id,
    }
  )

  arguments {
    name      = "startTimestamp"
    data_type = jsonencode({ typeKind : "TIMESTAMP" })
  }
  arguments {
    name      = "endTimestamp"
    data_type = jsonencode({ typeKind : "TIMESTAMP" })
  }
}
