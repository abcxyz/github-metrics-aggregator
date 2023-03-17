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

module "webhook" {
  source = "./modules/webhook"

  project_id          = var.project_id
  name                = var.name
  domain              = var.domain
  image               = var.image
  service_iam         = var.service_iam
  topic_iam           = var.topic_iam
  dead_letter_sub_iam = var.dead_letter_sub_iam
  dataset_location    = var.dataset_location
  dataset_id          = var.dataset_id
  dataset_iam         = var.dataset_iam
  table_id            = var.table_id
  table_iam           = var.table_iam
}
