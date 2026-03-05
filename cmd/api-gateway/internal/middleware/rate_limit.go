package middleware

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
}

type limiter struct {
	mu     sync.Mutex
	rps    float64
	burst  float64
	bkts   map[string]*bucket
	ttl    time.Duration
	maxBkt int
}

func NewRateLimit() func(http.Handler) http.Handler {
	rps := envFloat("RATE_LIMIT_RPS", 20)
	burst := envFloat("RATE_LIMIT_BURST", 40)

	l := &limiter{
		rps:    rps,
		burst:  burst,
		bkts:   make(map[string]*bucket),
		ttl:    10 * time.Minute,
		maxBkt: 10000,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r.RemoteAddr)
			if !l.allow(ip, time.Now()) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (l *limiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.bkts) > l.maxBkt {
		l.evictStale(now)
	}

	b, ok := l.bkts[key]
	if !ok {
		b = &bucket{tokens: l.burst, last: now}
		l.bkts[key] = b
	}

	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * l.rps
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now

	if b.tokens < 1 {
		return false
	}
	b.tokens -= 1
	return true
}

func (l *limiter) evictStale(now time.Time) {
	for k, b := range l.bkts {
		if now.Sub(b.last) > l.ttl {
			delete(l.bkts, k)
		}
	}
}

func envFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
