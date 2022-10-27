package util

/* References */
/* https://gobyexample.com/timers */
/* https://gobyexample.com/tickers */

import (
	"time" // https://golang.org/pkg/time/
)

// Timer is an implementation of Time based event dispatching
type Timer struct {
	callable      func()
	dur           time.Duration
	internalTimer *time.Timer
	cancel        chan struct{}
}

// NewTimer returns a new timer with the given callback function
func NewTimer(call func()) *Timer {
	t := &Timer{callable: call}
	t.cancel = make(chan struct{})
	return t
}

// SetTime sets the waiting time for the timer
func (t *Timer) SetTime(dur time.Duration) {
	t.dur = dur
}

func (t *Timer) SetCallAndCancel(call func()) {
	t.callable = call
	t.cancel = make(chan struct{})
}

// Start begins the countdown for the timer
func (t *Timer) Start() {
	t.internalTimer = time.NewTimer(t.dur)
	// Start a goroutine that waits for the timer to finish
	go func() {
		// Wait for timer to finish or if cancelled
		select {
		case <-t.internalTimer.C:
			t.callable()
		case _ = <-t.cancel:
			t.internalTimer.Stop()
			return
		}
	}()
}

// Reset restarts the timer for the same duration
func (t Timer) Reset() {
	t.internalTimer.Reset(t.dur)
}

// Cancel cancels the timer
func (t Timer) Cancel() {
	var a struct{}
	t.cancel <- a
}
