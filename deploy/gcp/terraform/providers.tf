# `billing_project` is stated rather than inherited: left out, the provider
# bills whatever project `gcloud config` last pointed at.
provider "google" {
  project               = var.project_id
  region                = var.region
  billing_project       = var.project_id
  user_project_override = true
}
