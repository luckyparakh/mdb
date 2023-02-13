package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type errType struct {
	Err string `json:"err"`
	Msg string `json:"msg,omitempty"`
}

func (a *application) createError(e error, msg string) errType {
	et := errType{
		Err: e.Error(),
		Msg: msg,
	}
	a.logger.PrintError(e, map[string]string{"msg": msg})
	return et
}

func (a *application) noRouteHandler(c *gin.Context) {
	e := a.createError(fmt.Errorf("no route"), "")
	c.JSON(http.StatusNotFound, e)
}

func (a *application) noMethodHandler(c *gin.Context) {
	e := a.createError(fmt.Errorf("no method found"), "")
	c.JSON(http.StatusNotFound, e)
}

func (a *application) authRequiredError(c *gin.Context, msg ...string) {
	m := "you must authenticate to access this resource"
	if len(msg) != 0 {
		m = msg[0]
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": m})
}

func (a *application) inactiveAccountError(c *gin.Context, msg ...string) {
	m := "user account must be activated to access this resource"
	if len(msg) != 0 {
		m = msg[0]
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": m})
}

func (a *application) noPermitError(c *gin.Context, msg ...string) {
	m := "user account does not have enough permissions to access this resource"
	if len(msg) != 0 {
		m = msg[0]
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": m})
}
