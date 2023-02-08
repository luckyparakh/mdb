package main

import (
	"mdb/internal/data"

	"github.com/gin-gonic/gin"
)

type contextKey string

const userContextKey = contextKey("user")

func (a *application) contextSetUser(c *gin.Context, user *data.User) {
	c.Set(string(userContextKey), user)
}

func (a *application) contextGetUser(c *gin.Context) *data.User {
	val, found := c.Get(string(userContextKey))
	if found {
		if user, ok := val.(*data.User); !ok {
			return nil
		} else {
			return user
		}
	}
	return nil
}
