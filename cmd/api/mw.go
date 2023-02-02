package main

import (
	"log"
	"net/http"
	"sync"
	"time"

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

	limiter := rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)
	return func(ctx *gin.Context) {
		// Any code here will run for every request that the middleware handles.
		if !limiter.Allow() {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "rate limit exceeded"})
			return
		}
		ctx.Next()
	}
}

func (app *application) rateLimiterPerHost() gin.HandlerFunc {
	if app.config.limiter.enable {
		type client struct {
			limiter  *rate.Limiter
			lastseen time.Time
		}
		var clients = make(map[string]*client)
		var mu sync.Mutex

		go func() {
			log.Printf("Cleaning clients: %v\n", clients)
			time.Sleep(15 * time.Second)
			mu.Lock()
			for k, v := range clients {
				if time.Since(v.lastseen) > 3*time.Minute {
					delete(clients, k)
				}
			}
			mu.Unlock()
		}()

		return func(ctx *gin.Context) {
			rip := ctx.RemoteIP()
			mu.Lock()
			if _, found := clients[rip]; !found {
				clients[rip] = &client{
					limiter:  rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
					lastseen: time.Now(),
				}
			}
			if !clients[rip].limiter.Allow() {
				mu.Unlock()
				ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "rate limit exceeded"})
				return
			}
			mu.Unlock()
			ctx.Next()
		}
	}
	return func(ctx *gin.Context) {}
}
