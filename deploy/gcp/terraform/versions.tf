terraform {
  # The same floors as aleogr/lab, which owns the shared instance this
  # project's database will live in. Two configurations on different provider
  # majors addressing one instance is a difference nobody chose.
  required_version = "~> 1.16.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 8.3"
    }
  }
}
