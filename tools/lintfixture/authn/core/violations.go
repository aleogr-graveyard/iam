// Package core is a lint fixture: it breaks the core boundary rules on purpose.
package core

import (
	"net/http"
	"time"
)

// Violations uses infrastructure and the wall clock inside a core package.
func Violations() (int, time.Time) {
	return http.StatusOK, time.Now()
}
