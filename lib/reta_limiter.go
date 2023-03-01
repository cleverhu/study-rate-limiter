package lib

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiterInterface interface {
	Accept() bool
}

type rateLimiter struct {
	cap           int64
	token         int64
	rate          int64
	lastTimestamp int64
	lock          sync.Mutex
}

func (r *rateLimiter) Accept() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	now := time.Now().Unix()
	delta := (now - r.lastTimestamp) * r.rate
	r.lastTimestamp = now
	token := r.token + delta
	if token > r.cap {
		token = r.cap
	}
	r.token = token

	if r.token > 0 {
		r.token--
		return true
	}
	return false
}

var _ RateLimiterInterface = new(rateLimiter)

func NewRateLimiter(rate, cap int64) RateLimiterInterface {
	return &rateLimiter{token: rate, cap: cap, rate: rate, lastTimestamp: time.Now().Unix()}
}

func RateLimiter(rate, cap int64) func(gin.HandlerFunc) gin.HandlerFunc {
	r := NewRateLimiter(rate, cap)
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			if !r.Accept() {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"message": "to many request",
				})
				return
			}
			handler(c)
		}
	}
}
