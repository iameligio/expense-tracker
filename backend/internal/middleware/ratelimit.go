package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// visitor tracks a per-key rate limiter and last-seen time for cleanup.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter throttles requests per client key (e.g. IP), with periodic GC.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     rate.Limit
	burst    int
}

// NewRateLimiter builds a limiter allowing `perMinute` events/min with the given burst.
func NewRateLimiter(perMinute float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(perMinute / 60.0),
		burst:    burst,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) get(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, ok := rl.visitors[key]
	if !ok {
		lim := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[key] = &visitor{limiter: lim, lastSeen: time.Now()}
		return lim
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupLoop() {
	for {
		time.Sleep(3 * time.Minute)
		rl.mu.Lock()
		for k, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow reports whether an event is permitted for an arbitrary key (e.g. an
// email address), consuming one token. Used for per-account throttling.
func (rl *RateLimiter) Allow(key string) bool {
	return rl.get(key).Allow()
}

// Middleware limits requests keyed by client IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.get(clientIP(r)).Allow() {
			writeError(w, http.StatusTooManyRequests, "too many requests, please slow down")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
