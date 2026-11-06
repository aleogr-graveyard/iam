package core

import t "time"

// ClockAlias reads the wall clock through an aliased import.
func ClockAlias(start t.Time) t.Duration { return t.Since(start) }
