package main

import (
	"expvar"
	"fmt"
	"log"
	"mdb/internal/data"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// func (app *application) rateLimiter() gin.HandlerFunc {
// 	// Any code here will run only once, when we wrap something with the middleware.

// 	// Initialize a new rate limiter which allows an average of 2 requests per second,
// 	// with a maximum of 4 requests in a single ‘burst’.

// 	limiter := rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)
// 	return func(ctx *gin.Context) {
// 		// Any code here will run for every request that the middleware handles.
// 		if !limiter.Allow() {
// 			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"msg": "rate limit exceeded"})
// 			return
// 		}
// 		ctx.Next()
// 	}
// }

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

func (app *application) authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader("Authorization")
		if authorizationHeader == "" {
			app.contextSetUser(ctx, data.AnonymousUser)
			ctx.Next()
			return
		}
		headerData := strings.Split(authorizationHeader, " ")
		if len(headerData) != 2 || headerData[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "Invalid Request"})
			return
		}
		token := headerData[1]
		if !data.ValidateTokenPlaintext(token) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "Invalid Token"})
			return
		}
		user, err := app.models.User.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "Invalid Token as no user found against it"})
			return
		}
		app.contextSetUser(ctx, user)
		ctx.Next()
	}
}

func (app *application) requireAuthenticatedUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := app.contextGetUser(ctx)
		if user.IsAnonymous() {
			app.authRequiredError(ctx, "authenticate to access this resource")
			return
		}
		ctx.Next()
	}
}
func (app *application) requireActivatedUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := app.contextGetUser(ctx)
		if !user.Activated {
			app.inactiveAccountError(ctx)
			return
		}
		ctx.Next()
	}
}

func (app *application) requirePermission(code string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := app.contextGetUser(ctx)
		perms, err := app.models.Permission.GetAllForUser(user.ID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "Not able to read permissions"})
			return
		}
		if !perms.Include(code) {
			app.noPermitError(ctx)
			return
		}
		ctx.Next()
	}
}

func (app *application) metrics() gin.HandlerFunc {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")
	// totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")
	return func(ctx *gin.Context) {
		fmt.Println("Metirc")
		start := time.Now()
		totalRequestsReceived.Add(1)
		ctx.Next()
		// Can't httpsnoop as gin context does not have http handler
		// httpsnoop.CaptureMetrics(ctx.Handler(), ctx.Writer, ctx.Request)
		totalResponsesSent.Add(1)
		diff := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(diff)
	}
}
