package main

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	reqRate  rate.Limit = 2
	reqBrust int        = 4
)

func (app *application) rateLimiter() gin.HandlerFunc {
	// Any code here will run only once, when we wrap something with the middleware.

	// Initialize a new rate limiter which allows an average of 2 requests per second,
	// with a maximum of 4 requests in a single ‘burst’.

	limter := rate.NewLimiter(reqRate, reqBrust)
	return func(ctx *gin.Context) {
		// Any code here will run for every request that the middleware handles.
		if !limter.Allow() {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "rate limit exceeded"})
			return
		}
		ctx.Next()
	}
}

func (app *application) rateLimterPerHost() gin.HandlerFunc {
	var clients map[string]*rate.Limiter
	var mu sync.Mutex
	
	return func(ctx *gin.Context) {
		rip := ctx.RemoteIP()
		mu.Lock()
		if _, found := clients[rip]; !found {
			clients[rip] = rate.NewLimiter(reqRate, reqBrust)
		}
		if !clients[rip].Allow() {
			mu.Unlock()
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "rate limit exceeded"})
			return
		}
		mu.Unlock()
		ctx.Next()
	}
}
