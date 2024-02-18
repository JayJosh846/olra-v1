package middleware

import (
	"errors"
	"net/http"
	"olra-v1/internal/database"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var SECRET_KEY = os.Getenv("SECRETS")
var JwtKey = []byte(SECRET_KEY)

type Claims struct {
	UserID   uint   `json:"user_id"`
	DeviceID string `json:"device_id"`
	jwt.StandardClaims
}

func AuthMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}

	c.Set("userID", claims.UserID)
	c.Next()
}

func LogoutUserFromDevice(deviceID string) error {
	// Retrieve user information from the database based on the device ID
	var user database.User
	var dbTokenMap database.DeviceTokenMapping
	if err := database.DB.Where("device_id = ?", deviceID).First(&user).Error; err != nil {
		return errors.New("user device not found")
	}
	// Update the user's record to remove the current device information
	user.DeviceID = ""
	// Assuming you're using GORM, save the updated user record
	if err := database.DB.Save(&user).Error; err != nil {
		// Handle error (e.g., database update failed)
		return errors.New("failed to update user record")
	}
	// 3. Optionally, perform additional actions like logging out the user from the current session
	// (This step may vary based on your application's architecture)
	result := database.DB.Where("device_id = ?", deviceID).Delete(&dbTokenMap)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("device ID not found")
	}

	// 4. Return nil if the logout process was successful
	return nil
}
