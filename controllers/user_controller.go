package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"olra-v1/internal/database"
	"olra-v1/services"
	helpers "olra-v1/utils"
)

var Validate = validator.New()

type VerifyPhoneOTPRequestBody struct {
	PinId string `json:"pinId" validate:"required"`
	Pin   string `json:"pin" validate:"required"`
	Phone string `json:"phone" validate:"required"`
}

type VerifyPhoneOTPBody struct {
	PinId    string `json:"pinId"`
	Verified bool   `json:"verified"`
	// Verifiedd string `json:"verified"`
	Msisdn string `json:"msisdn"`
	Status int    `json:"status"`
}

type PhoneOTPRequestBody struct {
	Mobile string `json:"mobile" validate:"required"`
}

type PhoneOTPResponse struct {
	PinId     string `json:"pinId"`
	To        string `json:"to"`
	SmsStatus string `json:"smsStatus"`
	Status    int    `json:"status"`
}

type UserRequestBody struct {
	FirstName   string `json:"firstName" validate:"required"`
	LastName    string `json:"lastName" validate:"required"`
	Email       string `json:"email" validate:"required"`
	PhoneNumber string `json:"phoneNumber" validate:"required"`
}

type EmailVerificationCodeRequest struct {
	Code  string `json:"code"`
	Email string `json:"email" validate:"required"`
}

func RequestPhoneOTP(c *gin.Context) {
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var (
		phoneOTPRequestBody PhoneOTPRequestBody
		phoneOTPResponse    PhoneOTPResponse
	)
	if err := c.BindJSON(&phoneOTPRequestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	validationErr := Validate.Struct(phoneOTPRequestBody)
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

	// Check if phone number already exists in the database
	var existingUser database.User
	if err := database.DB.Where("phone_number = ?", phoneOTPRequestBody.Mobile).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "User with this number exists.",
			"data":          "",
		})
		return
	}

	phoneOTP, err := services.SendPhoneOTP(phoneOTPRequestBody.Mobile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Could not send OTP request",
			"data":          "",
		})
		return
	}

	e := json.Unmarshal([]byte(phoneOTP), &phoneOTPResponse)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "OTP sent successfully",
		"data":          phoneOTPResponse,
	})
	return
}

func VerifyPhoneOTP(c *gin.Context) {
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var (
		verifyPhoneOTPRequestBody VerifyPhoneOTPRequestBody
		verifyPhoneOTPBody        VerifyPhoneOTPBody
	)
	if err := c.BindJSON(&verifyPhoneOTPRequestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	validationErr := Validate.Struct(verifyPhoneOTPRequestBody)
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

	verifyPhoneOTP, err := services.VerifyOTP(verifyPhoneOTPRequestBody.PinId, verifyPhoneOTPRequestBody.Pin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Somehting went wrong. Please try again",
			"data":          "",
		})
		return
	}
	fmt.Println("verifyPhoneOTP", verifyPhoneOTP)

	e := json.Unmarshal([]byte(verifyPhoneOTP), &verifyPhoneOTPBody)
	if e != nil {
		log.Println("Error:", e)
		return
	}
	if verifyPhoneOTPBody.Status == 400 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Token expired.",
			"data":          "",
		})
		return
	}
	fmt.Println("verifyPhoneOTPBody", verifyPhoneOTPBody)

	if verifyPhoneOTPBody.Verified == false {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Invalid token",
			"data":          "",
		})
		return
	}
	user := database.User{
		PhoneNumber:   verifyPhoneOTPRequestBody.Phone,
		PhoneVerified: true,
	}
	result := database.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"error":         false,
			"response code": 200,
			"message":       "Couldn't add user details",
			"data":          result.Error,
		})
		return
	}
	fmt.Println("result", result)
	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "OTP verified successfully",
		"data":          verifyPhoneOTPBody,
	})
	return
}

func AddUser(c *gin.Context) {
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var (
		userRequestBody UserRequestBody
		// verifyPhoneOTPBody        VerifyPhoneOTPBody
	)
	if err := c.BindJSON(&userRequestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	validationErr := Validate.Struct(userRequestBody)
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

	var existingUser database.User
	if err := database.DB.Where("email = ?", userRequestBody.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "User with this email address exists. Please login",
			"data":          "",
		})
		return
	}

	// Check if the user with the provided phone number exists
	if err := database.DB.Where("phone_number = ?", userRequestBody.PhoneNumber).First(&existingUser).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "User with this phone number doesn't exist",
			"data":          "",
		})
		return
	}

	userName, err := helpers.ExtractUsernameFromEmail(userRequestBody.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	// Update existing user with provided details
	if err := database.DB.Model(&existingUser).Updates(database.User{
		FirstName: userRequestBody.FirstName,
		LastName:  userRequestBody.LastName,
		Email:     userRequestBody.Email,
		Role:      "user",
		Tag:       userName,
		RefCode:   helpers.GenerateVerificationCode(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to update user",
			"data":          "",
		})
		return
	}
	returnedOTP, err := services.CreateOtp(&existingUser, userRequestBody.Email)
	defer cancel()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Error sending email verification otp",
			"data":          "",
		})
		return
	}
	c.JSON(http.StatusFound, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "Email verification otp sent successfully",
		"data":          returnedOTP,
	})

}

func EmailVerification(c *gin.Context) {
	var (
		emailVerificationCodeRequest EmailVerificationCodeRequest
		user                         database.User
	)

	if err := c.BindJSON(&emailVerificationCodeRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	if err := database.DB.Where("email = ?", emailVerificationCodeRequest.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "User does not exist",
			"data":          "",
		})
		return
	}

	// Check if email is already verified
	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Email address is already verified",
			"data":          "",
		})
		return
	}

	// Find OTP by token
	var otp database.Otp
	if err := database.DB.Where("token = ?", emailVerificationCodeRequest.Code).First(&otp).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "OTP not valid",
			"data":          "",
		})
		return
	}

	// Check if OTP has expired
	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "OTP has expired",
			"data":          "",
		})
		return
	}

	// Update user's email verification status
	if err := database.DB.Model(&user).Update("email_verified", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Failed to update email verification status",
			"data":          "",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "Email Verified successfully",
		"data":          "",
	})
}

func UserRoutes(rg *gin.RouterGroup) {
	userRoute := rg.Group("/user")
	userRoute.POST("/send-phone-otp", RequestPhoneOTP)
	userRoute.POST("/verify-phone-otp", VerifyPhoneOTP)
	userRoute.POST("/add-user", AddUser)
	userRoute.POST("/verify-email", EmailVerification)
}
