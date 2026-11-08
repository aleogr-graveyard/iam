package rootctx

import (
	"context"
	"testing"
)

// Tests may create a root context; no rule must fire here.
func TestRoot(t *testing.T) {
	_ = context.Background()
}
