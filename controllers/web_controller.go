package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"olra-v1/internal/database"
	"olra-v1/internal/structs"
)

func AddWhitelist(c *gin.Context) {
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var (
		waitlistRequest structs.WaitlistRequest
	)
	if err := c.BindJSON(&waitlistRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	validationErr := Validate.Struct(waitlistRequest)
	if validationErr != nil {
		fmt.Println(validationErr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       validationErr.Error(),
			"data":          "",
		})
		return
	}

	var existingUser database.Waitlist
	if err := database.DB.Where("email = ?", waitlistRequest.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "User with this email address exists in waitlist",
			"data":          "",
		})
		return
	}
	user := database.Waitlist{
		FullName: waitlistRequest.FullName,
		Email:    waitlistRequest.Email,
		Phone:    waitlistRequest.Phone,
	}
	result := database.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Couldn't add user",
			"data":          result.Error,
		})
		return
	}
	c.JSON(http.StatusFound, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "User added to waitlist successfully",
		"data":          user,
	})
}

func WebRoutes(rg *gin.RouterGroup) {
	webRoutes := rg.Group("/web")
	webRoutes.POST("/add-waitlist", AddWhitelist)
}
