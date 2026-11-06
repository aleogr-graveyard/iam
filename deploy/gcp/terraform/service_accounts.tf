/* THE RUNTIME IDENTITIES, created before anything they run exists, so that
   aleogr/lab can name them in its `tenants` map. Each gets its roles together
   with the resource it uses, not ahead of it. */

resource "google_service_account" "release" {
  account_id   = "iam-release"
  display_name = "Release"
  description  = "Publishes images and revisions and runs the migration job. Reads the shared instance's activationPolicy before starting."
}

resource "google_service_account" "run" {
  account_id   = "iam-run"
  display_name = "Service runtime"
  description  = "Runs the service. Connects to the shared instance as an IAM database user, with no password."
}

resource "google_service_account" "migrate" {
  account_id   = "iam-migrate"
  display_name = "Migration job runtime"
  description  = "Runs the migration job as the iam_migrator database user."
}
