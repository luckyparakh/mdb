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
	r.Use(a.rateLimiterPerHost())

	// Custom Validations
	v := binding.Validator.Engine().(*validator.Validate)
	v.RegisterValidation("yearrange", validation.YearRange)
	v.RegisterValidation("genre", validation.Genres)
	v.RegisterValidation("oneof", validation.OneOf)

	// r.Use(a.bodyValidationMW)
	r.GET("/v1/healthcheck", a.healthcheckHandler)
	r.POST("/v1/movies", a.createMovieHandler)
	r.GET("/v1/movies/:id", a.showMovieHandler)
	r.GET("/v1/movies", a.listMoviesHandler)
	r.PATCH("/v1/movies/:id", a.updateMovieHandler)
	r.DELETE("/v1/movies/:id", a.deleteMovieHandler)
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
