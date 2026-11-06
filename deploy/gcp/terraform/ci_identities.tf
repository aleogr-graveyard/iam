/* THE TWO IDENTITIES CI ACTS AS, and everything each one may do here.

   EVERY RESOURCE TYPE ADDED TO THIS CONFIGURATION NEEDS A MATCHING GRANT
   HERE, for both identities: the planner must read it and the deployer must
   write it. The bootstrap apply runs as the project owner, who can do
   everything, so a missing grant only shows up on the first CI run. Ask, for
   each new type: what does the deployer need to create it, including the
   right to act as any service account it runs as?

   The state bucket grants (objectViewer for the planner, objectAdmin for the
   deployer) are made by hand and deliberately not declared: a configuration
   managing its own access to its state would revoke it halfway through a
   destroy. See ../README.md. */

resource "google_service_account" "planner" {
  account_id   = "iam-planner"
  display_name = "CI: Terraform plan"
  description  = "Federated to ${var.github_repository} on any ref. Reads only."
}

resource "google_service_account" "deployer" {
  account_id   = "iam-deployer"
  display_name = "CI: Terraform apply"
  description  = "Federated to ${var.github_repository} on main only. Applies this configuration."
}

resource "google_service_account_iam_member" "planner_federation" {
  service_account_id = google_service_account.planner.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "${local.pool_principal_set}/attribute.repository_id/${var.github_repository_id}"
}

resource "google_service_account_iam_member" "deployer_federation" {
  service_account_id = google_service_account.deployer.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "${local.pool_principal_set}/attribute.ref_is_default/true"
}

# The planner reads every resource and IAM policy declared here, and needs
# `serviceusage.services.use` because the provider bills calls to this project.
resource "google_project_iam_member" "planner" {
  for_each = toset([
    "roles/viewer",
    "roles/iam.securityReviewer",
    "roles/serviceusage.serviceUsageConsumer",
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.planner.email}"
}

# `projectIamAdmin` can grant any role in this project, including to itself,
# so the deployer is effectively this project's owner. That is why it is
# reachable only from main, which only the owner merges into.
resource "google_project_iam_member" "deployer" {
  for_each = toset([
    "roles/iam.serviceAccountAdmin",
    "roles/iam.workloadIdentityPoolAdmin",
    "roles/resourcemanager.projectIamAdmin",
    "roles/serviceusage.serviceUsageAdmin",
  ])

  project = var.project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.deployer.email}"
}
