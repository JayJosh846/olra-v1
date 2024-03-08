package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"olra-v1/internal/database"
	"olra-v1/internal/structs"
	"olra-v1/middleware"
	"olra-v1/services"
	helpers "olra-v1/utils"
)

var Validate = validator.New()

func RequestPhoneOTP(c *gin.Context) {
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var (
		phoneOTPRequestBody structs.PhoneOTPRequestBody
		phoneOTPResponse    structs.PhoneOTPResponse
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

	phoneOTP, err := services.SendPhoneWelcomeOTP(phoneOTPRequestBody.Mobile)
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

}

func VerifyPhoneOTP(c *gin.Context) {
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var (
		verifyPhoneOTPRequestBody structs.VerifyPhoneOTPRequestBody
		verifyPhoneOTPBody        structs.VerifyPhoneOTPBody
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

	if !verifyPhoneOTPBody.Verified {
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
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

}

func AddUser(c *gin.Context) {
	var _, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var (
		userRequestBody structs.UserRequestBody
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
			"message":       "User with this phone number exist",
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
	returnedOTP, err := services.CreateEmailOtp(&existingUser, userRequestBody.Email)
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
		emailVerificationCodeRequest structs.EmailVerificationCodeRequest
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
	if err := database.DB.Model(&user).Updates(database.User{
		EmailVerified: true,
		SignupLevel:   2,
	}).Error; err != nil {
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
		bvnRequest    structs.BVNRequest
		verifyResonse structs.VerifyBVNResponse
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
	user.SignupLevel = 3
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
	var callbackData structs.CallbackData
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
		walletData       structs.WalletRequest
		generateResponse structs.GenerateWalletResponse
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

func CreateTag(c *gin.Context) {
	var (
		tagRequest structs.TagRequest
	)

	// Bind the request body to tagRequest struct
	if err := c.BindJSON(&tagRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract the email from query parameters
	userEmail := c.Query("email")
	if userEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "parameter required",
			"data":          "",
		})
		return
	}

	// Validate the tagRequest
	validationErr := Validate.Struct(tagRequest)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       validationErr.Error(),
			"data":          "",
		})
		return
	}

	validationError := helpers.ValidateTagRequest(tagRequest.Tag)
	if validationError != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       validationError.Error(),
			"data":          "",
		})
		return
	}

	// Check the user's KYC status
	var user database.User
	if err := database.DB.Where("email = ?", userEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to fetch user details",
			"data":          "",
		})
		return
	}
	if !user.KycStatus {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Please verify your KYC before creating a tag",
			"data":          "",
		})
		return
	}
	var existingUser database.User
	if err := database.DB.Where("tag = ?", tagRequest.Tag).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Tag name already exists",
			"data":          "",
		})
		return
	}
	generatedAccountNumber := helpers.GenerateRandomAccountNumber()
	// Create a wallet for the user in the VirtualAccount table
	virtualAccount := database.VirtualAccount{
		VirtualAccountBank:    "Guaranty Trust Bank",
		VirtualAccountAccount: generatedAccountNumber,
		VirtualAccountName:    user.FirstName + " " + user.LastName,
		UserID:                user.UserID,
		Balance:               0, // Default balance
	}
	if err := database.DB.Create(&virtualAccount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to create virtual account",
			"data":          "",
		})
		return
	}
	// Update the user's records in the user table
	user.Tag = tagRequest.Tag
	user.SignupLevel = 4
	// Update the tag field
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to update user records",
			"data":          "",
		})
		return
	}

	// Respond with success message
	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "Tag created successfully",
		"data":          "",
	})
}

func CreatePasscode(c *gin.Context) {
	var (
		passcodeRequest structs.PasscodeRequest
	)

	// Bind the request body to passcodeRequest struct
	if err := c.BindJSON(&passcodeRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract the email from query parameters
	userEmail := c.Query("email")
	if userEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "parameter required",
			"data":          "",
		})
		return
	}

	// Validate the tagRequest
	validationErr := Validate.Struct(passcodeRequest)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       validationErr.Error(),
			"data":          "",
		})
		return
	}
	if passcodeRequest.Passcode != passcodeRequest.ConfirmPasscode {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Passcodes are not the same",
			"data":          "",
		})
		return
	}

	validationError := helpers.ValidateOnlyNumbers(passcodeRequest.Passcode)
	if !validationError {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Only 6 digits numbers are allowed",
			"data":          "",
		})
		return
	}

	// Check the user's KYC status
	var user database.User
	if err := database.DB.Where("email = ?", userEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to fetch user details",
			"data":          "",
		})
		return
	}
	// if err := database.DB.Model(&user).Update(
	// 	"password_hash", helpers.HashPassword(passcodeRequest.Passcode),
	// )
	if err := database.DB.Model(&user).Updates(database.User{
		PasswordHash: helpers.HashPassword(passcodeRequest.Passcode),
		SignupLevel:  5,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to update user profile",
			"data":          "",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "Passcode set successfully",
		"data":          "",
	})
}

func LoginRequest(c *gin.Context) {
	var (
		user             database.User
		loginRequest     structs.LoginRequestBody
		phoneOTPResponse structs.PhoneOTPResponse
	)
	if err := c.BindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Invalid request",
			"data":          "",
		})
		return
	}
	// Find user by phone number
	if err := database.DB.Where("phone_number = ?", loginRequest.PhoneNumber).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       "Invaid phone number",
			"data":          "",
		})
		return
	}
	// Validate passcode
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginRequest.Passcode)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       "Invaid passcode",
			"data":          "",
		})
		return
	}
	phoneOTP, err := services.SendPhoneOTP(loginRequest.PhoneNumber)
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

}

func VerifyLoginRequest(c *gin.Context) {
	userDeviceID := c.Query("deviceID")
	var (
		user                      database.User
		verifyPhoneOTPRequestBody structs.VerifyPhoneOTPRequestBody
		verifyPhoneOTPBody        structs.VerifyPhoneOTPBody
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
	verifyPhoneOTP, err := services.VerifyOTP(
		verifyPhoneOTPRequestBody.PinId,
		verifyPhoneOTPRequestBody.Pin,
	)
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
	if !verifyPhoneOTPBody.Verified {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Invalid token",
			"data":          "",
		})
		return
	}
	// Find user by phone number
	if err := database.DB.Where("phone_number = ?", verifyPhoneOTPRequestBody.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       "Invaid phone number",
			"data":          "",
		})
		return
	}
	// Logout user from previous device
	if user.DeviceID != "" && user.DeviceID != userDeviceID {
		err := middleware.LogoutUserFromDevice(user.DeviceID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":         true,
				"response code": 401,
				"message":       err.Error(),
				"data":          "",
			})
			return
		}

	}
	// Update current device for the user
	user.DeviceID = userDeviceID
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to update user device",
			"data":          "",
		})
		return
	}
	expirationTime := time.Now().Local().Add(time.Hour * time.Duration(168))
	claims := &middleware.Claims{
		UserID:   user.UserID,
		DeviceID: userDeviceID, // Include Device ID in the token payload
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middleware.JwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         true,
			"response code": 500,
			"message":       "Failed to generate token",
			"data":          "",
		})
		return
	}
	deviceTokenMapping := database.DeviceTokenMapping{
		DeviceID: userDeviceID,
		Token:    tokenString,
	}
	database.DB.Create(&deviceTokenMapping)
	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "Login successful",
		"data":          gin.H{"token": tokenString},
	})

}

func Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         true,
			"response code": 400,
			"message":       "Search query is required",
			"data":          "",
		})
		return
	}
	var users []database.User
	if err := database.DB.Where("tag LIKE ?", "%"+query+"%").Find(&users).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":         true,
			"response code": 404,
			"message":       "Error retrieving users",
			"data":          "",
		})
		return
	}
	var selectedUsers []structs.UsersTags
	for _, user := range users {
		selectedUser := structs.UsersTags{
			User_ID: user.UserID,
			Tag:     user.Tag,
		}
		selectedUsers = append(selectedUsers, selectedUser)
	}
	c.JSON(http.StatusOK, gin.H{
		"error":         false,
		"response code": 200,
		"message":       "Users retrieved successfully",
		"data":          selectedUsers,
	})
}

func UserRoutes(rg *gin.RouterGroup) {
	userRoute := rg.Group("/user")
	userRoute.POST("/send-phone-otp", RequestPhoneOTP)
	userRoute.POST("/verify-phone-otp", VerifyPhoneOTP)
	userRoute.POST("/add-user", AddUser)
	userRoute.POST("/verify-email", EmailVerification)
	userRoute.POST("/verify-bvn", VerifyBVN)
	userRoute.POST("/add-tag", CreateTag)
	userRoute.POST("/add-passcode", CreatePasscode)
	userRoute.POST("/login-request", LoginRequest)
	userRoute.POST("/verify-login", VerifyLoginRequest)
	userRoute.POST("/search", Search)

	userRoute.POST("/callback", CallBack)
	userRoute.POST("/generate-wallet", GenerateWallet)

}
