package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (l *IPRateLimiter) StartCleanUp() {
	go l.cleanUpVisitors()
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	r        rate.Limit
	b        int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		visitors: make(map[string]*visitor),
		r:        r,
		b:        b,
	}

}

func (l *IPRateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		limiter := l.getVisitor(ip)
		if limiter.Allow() == false {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *IPRateLimiter) getVisitor(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, exists := l.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(l.r, l.b)
		v = &visitor{limiter, time.Now()}
		l.visitors[ip] = v
	}
	v.lastSeen = time.Now()

	return v.limiter
}

func (l *IPRateLimiter) cleanup(maxAge time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, v := range l.visitors {
		if time.Since(v.lastSeen) > maxAge {
			delete(l.visitors, ip)
		}
	}
}

func (l *IPRateLimiter) cleanUpVisitors() {
	for {
		time.Sleep(time.Minute)
		l.cleanup(3 * time.Minute)
	}
}
