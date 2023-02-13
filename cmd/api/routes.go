package main

import (
	"context"
	"encoding/json"
	"mdb/internal/data"
	"mdb/internal/validation"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func (a *application) routes() *gin.Engine {
	r := gin.Default()
	r.Use(a.metrics(), a.rateLimiterPerHost())
	r.Use(a.authenticate())

	// Custom Validations
	v := binding.Validator.Engine().(*validator.Validate)
	v.RegisterValidation("yearrange", validation.YearRange)
	v.RegisterValidation("genre", validation.Genres)
	v.RegisterValidation("oneof", validation.OneOf)

	// r.Use(a.bodyValidationMW)
	r.GET("/v1/healthcheck", a.healthcheckHandler)

	//Movies API
	movieGroup := r.Group("/v1/movies")
	movieGroup.Use(a.requireAuthenticatedUser(), a.requireActivatedUser())
	movieGroupRead := movieGroup.Group("")
	movieGroupRead.Use(a.requirePermission("movies:read"))
	movieGroupRead.GET("/:id", a.showMovieHandler)
	movieGroupRead.GET("", a.listMoviesHandler)
	movieGroupWrite := movieGroup.Group("")
	movieGroupWrite.Use(a.requirePermission("movies:write"))
	movieGroupWrite.POST("", a.createMovieHandler)
	movieGroupWrite.PATCH("/:id", a.updateMovieHandler)
	movieGroupWrite.DELETE("/:id", a.deleteMovieHandler)

	r.POST("/v1/users", a.registerUserHandler)
	r.PUT("/v1/users/activated", a.activateUserHandler)
	r.POST("/v1/tokens/authentication", a.createAuthenticationTokenHandler)
	r.GET("/debug/vars", expVarHandler(map[string]any{"memstats": nil, "cmdline": nil}))
	r.NoMethod(a.noMethodHandler)
	r.NoRoute(a.noRouteHandler)
	return r
}

func (a *application) bodyValidationMW(c *gin.Context) {
	var mv data.Movie
	maxBytes := 1_048_576
	body := c.Request.Body
	body = http.MaxBytesReader(c.Writer, body, int64(maxBytes))
	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&mv)
	if err != nil {
		c.JSON(http.StatusBadRequest, a.createError(err, "Bad req at MW"))
		c.Abort()
		return
	}
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "body", body))
	c.Next()
}
