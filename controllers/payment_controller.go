package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"olra-v1/internal/database"
	"olra-v1/internal/structs"
	"olra-v1/middleware"
	"olra-v1/services"
)

func RequestFunds(c *gin.Context) {
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       "User not found",
			"data":          "",
		})
		return
	}
	userStruct, ok := user.(middleware.User)
	fmt.Println("userStruct", userStruct)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "User not a valid struct",
			"data":          "",
		})
		return
	}
	var (
		requestFundsRequest structs.RequestFundsRequest
		fundsResponse       structs.FundsResponse
	)
	if err := c.BindJSON(&requestFundsRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	validationErr := Validate.Struct(requestFundsRequest)
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
	// Check for requestee details with tag
	var existingUser database.User
	if err := database.DB.Where("tag = ?", requestFundsRequest.Requestee).First(&existingUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":         true,
			"response code": 404,
			"message":       "User with this tag does not exist.",
			"data":          "",
		})
		return
	}
	smsFundsResponse, err := services.SendRequestFundsSMS(
		existingUser.PhoneNumber,
		existingUser.FirstName,
		existingUser.LastName,
		existingUser.Tag,
		requestFundsRequest.Amount,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Could not send funds sms request",
			"data":          "",
		})
		return
	}
	e := json.Unmarshal([]byte(smsFundsResponse), &fundsResponse)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	transaction := database.Transaction{
		UserID:             *userStruct.UserId,
		TransactionEnviron: "withinOlra",
		TransactionType:    "request",
		Amount:             requestFundsRequest.Amount,
		Description:        requestFundsRequest.Description,
		Requestee:          existingUser.Tag,
		Status:             "completed",
		TransactionDate:    time.Now(),
	}
	if err := database.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to create transaction instance",
			"data":          "",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "Successfully sent request",
		"data":          fundsResponse,
	})

}

func PaymentRoutes(rg *gin.RouterGroup) {
	paymentroute := rg.Group("/payment")
	paymentroute.POST(
		"/request-funds",
		middleware.AuthMiddleware,
		RequestFunds,
	)
}
