package main

import (
	"errors"
	"mdb/internal/data"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *application) registerUserHandler(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required,min=2,max=255"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6,max=255"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	if err := user.Password.Set(input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	err := a.models.User.Insert(user)
	if err != nil {
		var dupE *data.ErrDupEmail
		switch {
		case errors.As(err, &dupE):
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"err": "Duplicate Email"})
			return
		}
	}
	// msg := fmt.Sprintf("User %v created Successfully.", user)
	a.Background(func() {
		if err := a.mailer.Send(user.Email, "user_welcome.tmpl", user); err != nil {
			a.logger.PrintError(err, map[string]string{"user": user.Email, "msg": "Failed to send email"})
			// c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to send email", "user": user, "err": err.Error()})
			// return
		}
	})

	// log.Println("Send Mail")
	// a.mailer.SendRest()
	c.JSON(http.StatusAccepted, gin.H{"msg": "User Created Successfully", "user": user})
}
