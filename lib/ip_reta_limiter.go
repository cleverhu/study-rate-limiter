package lib

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func IPRateLimiter(rate, cap int64) func(gin.HandlerFunc) gin.HandlerFunc {
	cache := NewCache(10)
	return func(handler gin.HandlerFunc) gin.HandlerFunc {
		return func(c *gin.Context) {
			ip := c.Request.RemoteAddr
			limiter := cache.Get(ip)
			var l RateLimiterInterface
			if limiter != nil {
				l = limiter.(RateLimiterInterface)
			} else {
				l = NewRateLimiter(rate, cap)
				cache.Set(ip, l, time.Second*5)
			}

			if !l.Accept() {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"message": "to many request",
				})
				return
			}

			handler(c)
		}
	}
}
