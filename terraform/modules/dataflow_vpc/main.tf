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

data "github_ip_ranges" "gh_ranges" {}

# add a vpc for dataflow to use
resource "google_compute_network" "dataflow_vpc" {
  project = var.project_id

  name                    = "gma-df-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "dataflow_vpc_sub" {
  project = var.project_id

  name = "gma-df-${var.region}"

  ip_cidr_range = "10.1.0.0/26"
  region        = var.region
  network       = google_compute_network.dataflow_vpc.id
}

resource "google_compute_firewall" "dataflow_vpc_rule_node" {
  project = var.project_id

  name        = "gma-df-fw-rule-node"
  network     = google_compute_network.dataflow_vpc.name
  description = "Allow traffic between 'dataflow' nodes"

  allow {
    protocol = "tcp"
    ports    = ["12345-12346"]
  }

  source_tags = ["dataflow"]
  target_tags = ["dataflow"]
}

resource "google_compute_firewall" "dataflow_vpc_outbound" {
  project = var.project_id

  name        = "gma-df-fw-rule-outbound"
  network     = google_compute_network.dataflow_vpc.name
  description = "Allow 'dataflow' outbound traffic for GitHub access"

  direction          = "EGRESS"
  destination_ranges = data.github_ip_ranges.gh_ranges.api_ipv4

  allow {
    protocol = "tcp"
    ports    = ["443"]
  }
}

resource "google_compute_router" "dataflow_router" {
  project = var.project_id

  name    = "gma-df-router"
  network = google_compute_network.dataflow_vpc.name
  region  = var.region

  bgp {
    asn = 64514
  }
}

resource "google_compute_router_nat" "dataflow_nat" {
  project = var.project_id

  name                               = "gma-df-router-nat"
  router                             = google_compute_router.dataflow_router.name
  region                             = google_compute_router.dataflow_router.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}
