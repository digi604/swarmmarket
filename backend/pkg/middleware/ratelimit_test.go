package middleware

import (
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	// Create limiter with 10 RPS and burst of 10
	limiter := NewRateLimiter(10, 10)

	key := "test-key"

	// First 10 requests should succeed (burst)
	for i := 0; i < 10; i++ {
		if !limiter.Allow(key) {
			t.Errorf("request %d should be allowed", i)
		}
	}

	// 11th request should be denied
	if limiter.Allow(key) {
		t.Error("11th request should be denied")
	}
}

func TestRateLimiterRefill(t *testing.T) {
	// Create limiter with high RPS for quick refill
	limiter := NewRateLimiter(100, 1)

	key := "test-key"

	// Use up the token
	if !limiter.Allow(key) {
		t.Error("first request should be allowed")
	}

	// Should be denied immediately
	if limiter.Allow(key) {
		t.Error("second immediate request should be denied")
	}

	// Wait for refill (10ms = 1 token at 100 RPS)
	time.Sleep(15 * time.Millisecond)

	// Should be allowed after refill
	if !limiter.Allow(key) {
		t.Error("request after refill should be allowed")
	}
}

func TestRateLimiterDifferentKeys(t *testing.T) {
	limiter := NewRateLimiter(10, 1)

	// Different keys should have independent limits
	if !limiter.Allow("key1") {
		t.Error("key1 first request should be allowed")
	}
	if !limiter.Allow("key2") {
		t.Error("key2 first request should be allowed")
	}

	// key1 exhausted
	if limiter.Allow("key1") {
		t.Error("key1 second request should be denied")
	}
	// key2 also exhausted
	if limiter.Allow("key2") {
		t.Error("key2 second request should be denied")
	}
}
