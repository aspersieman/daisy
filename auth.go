package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for login page
		if c.Request.URL.Path == "/login" {
			c.Next()
			return
		}
		if c.Request.URL.Path == "/" {
			c.Next()
			return
		}
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}
		// TODO: Generate secret key for JWT and read from .env
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return []byte("your-secret-key"), nil
		})
		if err != nil {
			log.Printf("ERROR: Error parsing token: %v", err)
			c.JSON(401, gin.H{"error": "Invalid token a"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("user", claims["user"])
		} else {
			c.JSON(401, gin.H{"error": "Invalid token b"})
			c.Abort()
		}
	}
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Find user by username
	user, err := FindUserByUsername(username)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user.ID,
	})
	// TODO: Generate secret key for JWT and read from .env
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{"token": tokenString})
}

func Logout(c *gin.Context) {
	// Remove token from header
	c.Header("Authorization", "")
	c.JSON(200, gin.H{"message": "Logged out"})
}
