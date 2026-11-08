// Package rootctx is a lint fixture: code outside entry points must not create a root context.
package rootctx

import "context"

// Background creates a root context where a caller's ctx should be passed.
func Background() context.Context { return context.Background() }

// Todo creates a root context where a caller's ctx should be passed.
func Todo() context.Context { return context.TODO() }
