package main

import (
	"errors"
	"fmt"
	"mdb/internal/data"
	"net/http"
	"time"

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

	err = a.models.Permission.AddForUser(user.ID, "movies:read")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": "Failed to add permission"})
		return
	}

	token, err := a.models.Token.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		a.logger.PrintError(err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"err": err})
		return
	}
	fmt.Printf("Token:%v\n", token)
	a.Background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}
		if err := a.mailer.Send(user.Email, "user_welcome.tmpl", data); err != nil {
			a.logger.PrintError(err, map[string]string{"user": user.Email, "msg": "Failed to send email"})
			// c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to send email", "user": user, "err": err.Error()})
			// return
		}
	})

	// log.Println("Send Mail")
	// a.mailer.SendRest()
	c.JSON(http.StatusAccepted, gin.H{"msg": "User Created Successfully", "user": user})
}

func (a *application) activateUserHandler(c *gin.Context) {
	var input struct {
		Token string `json:"token" binding:"required,len=26"`
		Scope string `json:"scope" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		a.logger.PrintError(err, map[string]string{"activateUserHandler": "error while binding user input"})
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	user, err := a.models.User.GetForToken(input.Scope, input.Token)
	if err != nil {
		a.logger.PrintError(err, map[string]string{"activateUserHandler": "error while getting user details against token"})
		c.JSON(http.StatusBadRequest, gin.H{"err": "Inactive or exipred token"})
		return
	}
	user.Activated = true
	err = a.models.User.Update(user)
	if err != nil {
		a.logger.PrintError(err, map[string]string{"activateUserHandler": "error while updating user details post activation"})
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	err = a.models.Token.Delete(user.ID, data.ScopeActivation)
	if err != nil {
		a.logger.PrintError(err, map[string]string{"activateUserHandler": "error while deleting token post activation"})
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User Activated", "user": user})
}

func (a *application) createAuthenticationTokenHandler(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Passowrd string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		a.logger.PrintError(err, map[string]string{"activateTokenAuth": "error while binding user input"})
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	user, err := a.models.User.GetByEmail(input.Email)
	if err != nil {
		a.logger.PrintError(err, map[string]string{"activateTokenAuth": "error while getting user details by email"})
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	match, err := user.Password.Matches(input.Passowrd)
	if err != nil {
		a.logger.PrintError(err, map[string]string{"activateTokenAuth": "error while matching password"})
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	if !match {
		a.logger.PrintInfo("activateTokenAuth:Password mismatch", nil)
		c.JSON(http.StatusBadRequest, gin.H{"err": "password mismatch"})
		return
	}
	token, err := a.models.Token.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		a.logger.PrintError(err, map[string]string{"activateTokenAuth": "error while generating token"})
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Token created successfully", "token": token})
}
