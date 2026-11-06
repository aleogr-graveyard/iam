/* THIS PROJECT'S DATABASE INSIDE THE SHARED INSTANCE, and nothing else there.

   The instance belongs to aleogr/lab (docs/lab.md there). This file declares
   only the `iam` database and the `iam` users. `sqlTenant` lets this identity
   delete any tenant's users, so nothing here may name a database, user or
   role outside `iam` / `iam_*` / `iam-*`.

   What the Cloud SQL API cannot express (REVOKE CONNECT FROM PUBLIC, the
   connection limit, the `iam_app` role and its membership) is in
   ../sql/bootstrap.sql, run once by the owner. */

# ABANDON: with the default, removing or renaming this block drops the
# database. `sqlTenant` holds no `cloudsql.databases.delete` either.
resource "google_sql_database" "iam" {
  project         = var.shared_sql_project
  instance        = var.shared_sql_instance
  name            = "iam"
  deletion_policy = "ABANDON"

  depends_on = [google_project_service.enabled]
}

# Owns the schema and runs migrations. A built-in user, created WITHOUT a
# password: the owner sets it by hand from the secret below, so the value
# never passes through Terraform or its state.
resource "google_sql_user" "migrator" {
  project  = var.shared_sql_project
  instance = var.shared_sql_instance
  name     = "iam_migrator"
  type     = "BUILT_IN"

  depends_on = [google_sql_database.iam]
}

# The service logs in with an IAM token, no password. Cloud SQL names an IAM
# service-account user by its email without ".gserviceaccount.com". The name
# contains this project's id, so it cannot collide with another tenant.
resource "google_sql_user" "run" {
  project  = var.shared_sql_project
  instance = var.shared_sql_instance
  name     = trimsuffix(google_service_account.run.email, ".gserviceaccount.com")
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"

  depends_on = [google_sql_database.iam]
}

# The container for iam_migrator's password. The value is added by hand.
resource "google_secret_manager_secret" "migrator_password" {
  secret_id = "iam-migrator-password"

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }

  depends_on = [google_project_service.enabled]
}

# The migration job reads it at run time. Nobody else is granted access.
resource "google_secret_manager_secret_iam_member" "migrate_reads_password" {
  secret_id = google_secret_manager_secret.migrator_password.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.migrate.email}"
}
