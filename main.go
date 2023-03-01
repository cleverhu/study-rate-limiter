package main

import (
	"github.com/gin-gonic/gin"

	"my-lru/lib"
)

func main() {
	r := gin.Default()
	r.GET("/", lib.IPRateLimiter(1, 10)(func(c *gin.Context) {
		c.String(200, "hello world")
	}))
	r.Run()
}
