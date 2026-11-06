output "workload_identity_provider" {
  description = "Value for the repository variable GCP_WIF_PROVIDER."
  value       = google_iam_workload_identity_pool_provider.github.name
}

output "planner_email" {
  description = "Value for the repository variable GCP_PLANNER_SA."
  value       = google_service_account.planner.email
}

output "deployer_email" {
  description = "Value for the repository variable GCP_DEPLOYER_SA."
  value       = google_service_account.deployer.email
}

output "runtime_service_accounts" {
  description = "The identities aleogr/lab names in its tenants map."
  value = {
    release = google_service_account.release.email
    run     = google_service_account.run.email
    migrate = google_service_account.migrate.email
  }
}
