resource "google_project_service" "enabled" {
  for_each = toset([
    "cloudresourcemanager.googleapis.com",
    "iam.googleapis.com",
    "iamcredentials.googleapis.com",
    "serviceusage.googleapis.com",
    "storage.googleapis.com",
    "sts.googleapis.com",
  ])

  service = each.value

  # Disabling an API breaks what it created instead of removing it. Both
  # settings are needed: `deletion_policy` defaults to DELETE, which for this
  # resource means disabling the service.
  disable_on_destroy = false
  deletion_policy    = "ABANDON"
}
