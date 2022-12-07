provider "google" {
  project = var.project_id
}

provider "google-beta" {
  project = var.project_id
}

terraform {
  required_version = ">=1.0.0"
  required_providers {
    google = {
      version = "~> 4.0"
    }
    google-beta = {
      version = "~> 4.0"
    }
  }
}
