package core

import "github.com/aleogr/iam/internal/database"

// InfraImport reaches internal infrastructure, which is not on the core allow-list.
func InfraImport() string { return database.Name }
