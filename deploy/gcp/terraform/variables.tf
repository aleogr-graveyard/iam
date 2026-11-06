variable "project_id" {
  description = "This environment's project. Arrives from the repository variable GCP_PROJECT_ID, never from a versioned file."
  type        = string
}

variable "region" {
  description = "Where regional resources run. The same region as the shared Cloud SQL instance, to avoid cross-region connections."
  type        = string
  default     = "us-central1"
}

variable "github_repository" {
  description = "The repository allowed to federate, as owner/name. Informational: trust is granted by github_repository_id."
  type        = string
}

variable "github_repository_id" {
  description = "Numeric id of the repository allowed to federate. Unlike the name, it survives a transfer or rename and cannot be claimed by a new repository."
  type        = string

  validation {
    condition     = can(regex("^[0-9]+$", var.github_repository_id))
    error_message = "github_repository_id must be the numeric repository id."
  }
}
