// Authentication via GOOGLE_APPLICATION_CREDENTIALS envvar

provider "google" {
  project     = var.gcp_project
  region      = var.gcp_region
}