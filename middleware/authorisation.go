package middleware

import (
	"errors"
	"net/http"
	"olra-v1/internal/database"
	"os"
	"time"

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

type User struct {
	UserId   *uint
	DeviceId *string
}

func AuthMiddleware(c *gin.Context) {

	// tokenString := c.Request.Header.Get("token")
	tokenString := c.GetHeader("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       "No token provided",
			"data":          "",
		})
		c.Abort()
		return
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       err.Error(),
			"data":          "",
		})
		c.Abort()
		return
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       "The Token is invalid",
			"data":          "",
		})
		c.Abort()
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":         true,
			"response code": 401,
			"message":       "The Token has expired",
			"data":          "",
		})
		c.Abort()
		return
	}
	user := User{
		UserId:   &claims.UserID,
		DeviceId: &claims.DeviceID,
	}
	c.Set("user", user)
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
	if err := database.DB.Save(&user).Error; err != nil {
		return errors.New("failed to update user record")
	}
	// Logout user previous session
	result := database.DB.Where("device_id = ?", deviceID).Delete(&dbTokenMap)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("device ID not found")
	}
	return nil
}
