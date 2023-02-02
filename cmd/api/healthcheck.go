package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *application) healthcheckHandler(c *gin.Context) {
	msg := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": a.config.env,
			"version":     version},
	}
	// msg := fmt.Sprintf(`{"status": "available", "environment": %s, "version": %s"`, a.config.env, version)
	c.JSON(http.StatusOK, &msg)
}
