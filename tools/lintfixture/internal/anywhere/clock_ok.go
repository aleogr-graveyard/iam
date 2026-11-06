package anywhere

import "time"

// ClockOK calls time.Now outside a core package; the forbidigo rule must not fire here.
func ClockOK() time.Time {
	return time.Now()
}
