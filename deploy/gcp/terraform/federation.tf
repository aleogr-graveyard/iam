/* GITHUB ACTIONS AUTHENTICATES WITH NO KEY. GitHub mints an OIDC token, the
   pool exchanges it for a short-lived credential, and nothing long-lived
   exists to leak.

   TRUST IS BY REPOSITORY ID, NOT NAME. The repository is expected to move to
   its own organization once the brand is chosen; the id survives that move,
   and a name can be re-created by someone else after it.

   TWO IDENTITIES, SPLIT BY BRANCH. Any workflow run of the repository may act
   as the planner, which only reads. Only a run on refs/heads/main may act as
   the deployer, which writes. A pull request branch cannot apply. */
resource "google_iam_workload_identity_pool" "github" {
  workload_identity_pool_id = "github"
  display_name              = "GitHub Actions"

  depends_on = [google_project_service.enabled]
}

resource "google_iam_workload_identity_pool_provider" "github" {
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  display_name                       = "GitHub Actions"

  attribute_mapping = {
    "google.subject"           = "assertion.sub"
    "attribute.repository"     = "assertion.repository"
    "attribute.repository_id"  = "assertion.repository_id"
    "attribute.ref_is_default" = "assertion.ref == 'refs/heads/main' ? 'true' : 'false'"
  }

  # Without a condition, any repository on GitHub could exchange a token.
  attribute_condition = "assertion.repository_id == '${var.github_repository_id}'"

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

locals {
  pool_principal_set = "principalSet://iam.googleapis.com/projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/${google_iam_workload_identity_pool.github.workload_identity_pool_id}"
}

data "google_project" "current" {
  project_id = var.project_id
}
