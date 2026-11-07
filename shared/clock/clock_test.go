package clock

import (
	"testing"
	"time"
)

func TestFixed_NowAndAdvance(t *testing.T) {
	start := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	c := &Fixed{T: start}
	var _ Clock = c
	if got := c.Now(); !got.Equal(start) {
		t.Fatalf("Now = %v, want %v", got, start)
	}
	c.Advance(90 * time.Second)
	if got := c.Now(); !got.Equal(start.Add(90 * time.Second)) {
		t.Fatalf("after Advance: %v", got)
	}
}

func TestFixed_ConcurrentUse(t *testing.T) {
	c := &Fixed{T: time.Unix(0, 0)}
	done := make(chan struct{})
	go func() {
		for range 100 {
			c.Advance(time.Millisecond)
		}
		close(done)
	}()
	for range 100 {
		_ = c.Now()
	}
	<-done
	if got := c.Now(); !got.Equal(time.Unix(0, 0).Add(100 * time.Millisecond)) {
		t.Fatalf("Now = %v", got)
	}
}

func TestSystem_Now(t *testing.T) {
	var c Clock = System{}
	a := c.Now()
	b := c.Now()
	if a.IsZero() || b.Before(a) {
		t.Fatalf("system clock not monotonic: %v then %v", a, b)
	}
}

func TestFixed_Set(t *testing.T) {
	c := &Fixed{T: time.Unix(100, 0)}
	c.Set(time.Unix(5, 0))
	if got := c.Now(); !got.Equal(time.Unix(5, 0)) {
		t.Fatalf("after Set: %v", got)
	}
}
