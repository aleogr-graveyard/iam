package testdb

import "context"

// Root is an entry point creating a root context; no rule must fire here.
func Root() context.Context { return context.Background() }
