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

output "service_url" {
  description = "Cloud Run service URL."
  value       = google_cloud_run_service.default.status[0].url
}

output "service_account_email" {
  description = "Cloud Run service account email."
  value       = google_service_account.default.email
}

output "service_account_iam_email" {
  description = "Cloud Run service account email iam string."
  value       = format("serviceAccount:%s", google_service_account.default.email)
}
