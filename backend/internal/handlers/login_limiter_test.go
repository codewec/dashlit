package handlers

import (
	"testing"
	"time"
)

func TestLoginLimiterBlocksAfterFailuresUntilWindowEnds(t *testing.T) {
	now := time.Unix(0, 0)
	l := newLoginLimiter()
	l.now = func() time.Time { return now }
	for i := 0; i < loginMaxFailures; i++ {
		if l.blocked("Alice") {
			t.Fatalf("blocked after %d failures", i)
		}
		l.fail("alice")
	}
	if !l.blocked("ALICE") {
		t.Fatal("not blocked after max failures")
	}
	if l.blocked("bob") {
		t.Fatal("other user blocked")
	}
	now = now.Add(loginWindow)
	if l.blocked("alice") {
		t.Fatal("still blocked after window")
	}
	l.fail("alice")
	l.succeed("alice")
	if _, ok := l.failures["alice"]; ok {
		t.Fatal("success did not clear failures")
	}
}
