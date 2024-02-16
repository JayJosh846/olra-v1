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

type BVNRequest struct {
	Bvn string `json:"bvn"`
}

type VerifyBVNResponse struct {
	Status  bool                    `json:"status"`
	Message string                  `json:"message"`
	Data    VerifyBVNResponseEntity `json:"data"`
}

type VerifyBVNResponseEntity struct {
	Entity VerifyBVNResponseBvn `json:"entity"`
}

type VerifyBVNResponseBvn struct {
	Bvn VerifyBVNResponseData `json:"bvn"`
}

type VerifyBVNResponseData struct {
	Status bool `json:"status"`
}

type CallbackData struct {
	Title   string `json:"Title"`
	Message string `json:"Message"`
	Data    Data   `json:"Data"`
}

// Data represents the dynamic data structure within the callback data
type Data struct {
	NUBANName   string `json:"NUBANName"`
	NUBAN       string `json:"NUBAN"`
	NUBANStatus string `json:"NUBANStatus"`
	NUBANType   int    `json:"NUBANType"`
	Request     int    `json:"Request"`
}

type WalletRequest struct {
	Gender      string `json:"gender"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Dob         string `json:"dob"`
	PhoneNumber string `json:"phoneNumber"`
}

type GenerateWalletResponse struct {
	Successful bool   `json:"Successful"`
	Message    string `json:"Message"`
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

func VerifyBVN(c *gin.Context) {
	var (
		bvnRequest    BVNRequest
		verifyResonse VerifyBVNResponse
	)

	if err := c.BindJSON(&bvnRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userEmail := c.Query("email")

	validationErr := Validate.Struct(bvnRequest)
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

	bvnVerify, err := services.VerifyBVN(bvnRequest.Bvn)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	err = json.Unmarshal([]byte(bvnVerify), &verifyResonse)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	if !verifyResonse.Status {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Could not verify BVN. Please try again",
			"data":          "",
		})
		return
	}

	user := database.User{}
	err = database.DB.Where("email = ?", userEmail).First(&user).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "User does not exist",
			"data":          "",
		})
		return
	}

	user.BvnVerified = true
	user.KycStatus = true
	err = database.DB.Save(&user).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Failed to update user",
			"data":          "",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "BVN verified successfully",
		"data":          "",
	})
}

func CallBack(c *gin.Context) {
	var callbackData CallbackData
	// if err := c.BindJSON(&callbackData); err != nil {
	// 	log.Println("Error binding callback data:", err)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid callback data"})
	// 	return
	// }

	response, err := services.Callback()

	err = json.Unmarshal([]byte(response), &callbackData)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	// Process the callback data
	fmt.Println("Received callback:", callbackData)
	fmt.Println("Title:", callbackData.Title)
	fmt.Println("Message:", callbackData.Message)
	fmt.Println("NUBAN Name:", callbackData.Data.NUBANName)
	fmt.Println("NUBAN:", callbackData.Data.NUBAN)
	fmt.Println("NUBAN Status:", callbackData.Data.NUBANStatus)
	fmt.Println("NUBAN Type:", callbackData.Data.NUBANType)
	fmt.Println("Request:", callbackData.Data.Request)

	// Handle different types of requests
	switch callbackData.Data.Request {
	case 1:
		fmt.Println("Request type: Wallet Creation")
		// Handle wallet creation request
		// Update your system with the newly generated wallet information
	case 2:
		fmt.Println("Request type: Account Creation")
		// Handle account creation request
	case 3:
		fmt.Println("Request type: Pin Validation")
		// Handle pin validation request
	case 4:
		fmt.Println("Request type: Payment Response")
		// Handle payment response request
	default:
		fmt.Println("Unknown request type")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Callback processed successfully"})
}

func GenerateWallet(c *gin.Context) {

	var (
		walletData       WalletRequest
		generateResponse GenerateWalletResponse
	)
	if err := c.BindJSON(&walletData); err != nil {
		log.Println("Error binding callback data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid callback data"})
		return
	}

	response, err := services.GenerateWallet(
		walletData.Gender,
		walletData.Email,
		walletData.FirstName,
		walletData.LastName,
		walletData.Dob,
		walletData.PhoneNumber,
	)

	err = json.Unmarshal([]byte(response), &generateResponse)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": generateResponse})

}

func UserRoutes(rg *gin.RouterGroup) {
	userRoute := rg.Group("/user")
	userRoute.POST("/send-phone-otp", RequestPhoneOTP)
	userRoute.POST("/verify-phone-otp", VerifyPhoneOTP)
	userRoute.POST("/add-user", AddUser)
	userRoute.POST("/verify-email", EmailVerification)
	userRoute.POST("/verify-bvn", VerifyBVN)
	userRoute.POST("/callback", CallBack)
	userRoute.POST("/generate-wallet", GenerateWallet)

}
