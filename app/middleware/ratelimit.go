package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"golang.org/x/time/rate"
)

type RateLimit struct {
	rps   float64
	burst int
}

func NewRateLimit(rps float64) *RateLimit {
	if rps <= 0 {
		return nil
	}

	burst := int(rps)
	if burst < 1 {
		burst = 1
	}

	return &RateLimit{
		rps:   rps,
		burst: burst,
	}
}

func (rl *RateLimit) Middleware(next http.Handler) http.Handler {
	if rl == nil {
		return next
	}

	var (
		mu       sync.Mutex
		limiters = make(map[string]*rate.Limiter)
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientIP(r)
		limiter := rl.limiterFor(&mu, limiters, key)

		if !limiter.Allow() {
			api.ErrorResponse(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimit) limiterFor(mu *sync.Mutex, limiters map[string]*rate.Limiter, key string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, ok := limiters[key]
	if !ok {
		limiter = rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
		limiters[key] = limiter
	}

	return limiter
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
