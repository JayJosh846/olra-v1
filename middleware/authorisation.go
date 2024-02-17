package middleware

// import (
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/dgrijalva/jwt-go"
// )

// type Claims struct {
// 	UserID uint `json:"user_id"`
// 	jwt.StandardClaims
// }

// func authMiddleware(c *gin.Context) {
// 	tokenString := c.GetHeader("Authorization")
// 	if tokenString == "" {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		c.Abort()
// 		return
// 	}

// 	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
// 		return jwtKey, nil
// 	})
// 	if err != nil || !token.Valid {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		c.Abort()
// 		return
// 	}

// 	claims, ok := token.Claims.(*Claims)
// 	if !ok {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
// 		c.Abort()
// 		return
// 	}

// 	c.Set("userID", claims.UserID)
// 	c.Next()
// }
