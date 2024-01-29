terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.12.0"
    }
  }
}

provider "google" {
  credentials = file("credentials.json")

  project = "prime-bonbon-407317"
  region  = "europe-west10"
  zone    = "europe-west10-a"
}

module "startup-scripts" {
  source  = "terraform-google-modules/startup-scripts/google"
  version = "2.0.0"
}

module "cloud-nat" {
  source     = "terraform-google-modules/cloud-nat/google"
  version    = "~> 5.0"
  router     = google_compute_router.router.name
  project_id = "prime-bonbon-407317"
  region     = "europe-west10"
}