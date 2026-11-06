# The state bucket is created by hand before the first apply (see
# ../README.md). Its name arrives through -backend-config from the repository
# variable TF_STATE_BUCKET; the prefix comes from lab/backend.hcl.
terraform {
  backend "gcs" {}
}
