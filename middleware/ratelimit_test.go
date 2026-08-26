package middleware

import (
	"testing"
	"time"
)

func TestIPRateLimiter_BurstExceeded(t *testing.T) {
	l := NewIPRateLimiter(1, 3) // 1/сек, burst 3

	allowed, denied := 0, 0
	for i := 0; i < 5; i++ {
		if l.getVisitor("1.1.1.1").Allow() {
			allowed++
		} else {
			denied++
		}
	}

	if allowed != 3 {
		t.Fatalf("allowed = %d, want 3", allowed)
	}
	if denied != 2 {
		t.Fatalf("denied = %d, want 2", denied)
	}
}

func TestIPRateLimiter_DifferentIPsAreIndependent(t *testing.T) {
	l := NewIPRateLimiter(1, 3)

	for i := 0; i < 3; i++ {
		if !l.getVisitor("1.1.1.1").Allow() {
			t.Fatalf("client A denied on request %d, expected within burst", i+1)
		}
	}
	if l.getVisitor("1.1.1.1").Allow() {
		t.Fatalf("client A should be over budget by now")
	}

	if !l.getVisitor("1.1.1.2").Allow() {
		t.Fatal("client B was denied even though it never made a request before")
	}
}

func TestIPRateLimiter_CleanupRemovesOnlyStale(t *testing.T) {
	l := NewIPRateLimiter(1, 3)

	_ = l.getVisitor("old-ip")
	_ = l.getVisitor("fresh-ip")

	l.mu.Lock()
	l.visitors["old-ip"].lastSeen = time.Now().Add(-10 * time.Minute)
	l.mu.Unlock()

	l.cleanup(3 * time.Minute)

	l.mu.Lock()
	_, oldStillThere := l.visitors["old-ip"]
	_, freshStillThere := l.visitors["fresh-ip"]
	l.mu.Unlock()

	if oldStillThere {
		t.Error("old-ip should have been removed by cleanup")
	}
	if !freshStillThere {
		t.Error("fresh-ip should NOT have been removed by cleanup")
	}
}
