package handlers

import (
	"strings"
	"sync"
	"time"
)

const (
	loginMaxFailures = 10
	loginWindow      = 15 * time.Minute
)

// loginLimiter throttles password guessing per username. It is keyed by
// username rather than client IP because the IP comes from proxy headers a
// client can forge. A success clears the counter; a lockout ends with the window.
type loginLimiter struct {
	mu       sync.Mutex
	now      func() time.Time
	failures map[string]*loginFailures
}

type loginFailures struct {
	count int
	start time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{now: time.Now, failures: map[string]*loginFailures{}}
}

func (l *loginLimiter) blocked(username string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	f := l.current(strings.ToLower(username))
	return f != nil && f.count >= loginMaxFailures
}

func (l *loginLimiter) fail(username string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := strings.ToLower(username)
	f := l.current(key)
	if f == nil {
		// Drop expired entries so made-up usernames cannot grow the map forever.
		for k, v := range l.failures {
			if l.now().Sub(v.start) >= loginWindow {
				delete(l.failures, k)
			}
		}
		f = &loginFailures{start: l.now()}
		l.failures[key] = f
	}
	f.count++
}

func (l *loginLimiter) succeed(username string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, strings.ToLower(username))
}

// current returns the live entry for key, or nil when none or expired. Caller holds mu.
func (l *loginLimiter) current(key string) *loginFailures {
	f := l.failures[key]
	if f == nil || l.now().Sub(f.start) >= loginWindow {
		return nil
	}
	return f
}
