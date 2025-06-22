package main

import (
	"embed"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

var db *sql.DB

//go:embed web
var staticFS embed.FS

func main() {
	db = InitDB("daisy.db")

	r := gin.Default()

	r.StaticFS("/web", http.FS(staticFS))
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("web/index.htm", http.FS(staticFS))
	})
	r.GET("/login", func(c *gin.Context) {
		c.FileFromFS("web/login.htm", http.FS(staticFS))
	})
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("web/favicon.ico", http.FS(staticFS))
	})
	r.POST("/api/login", loginHandler)
	r.GET("/api/authenticated", authenticatedHandler)

	api := r.Group("/api")
	api.Use(AuthMiddleware())
	api.GET("/entries", getEntries)
	api.GET("/entries/:id", getEntry)
	api.POST("/entries", createEntry)
	api.PUT("/entries/:id", updateEntry)
	api.DELETE("/entries/:id", deleteEntry)

	r.Run(":8069")
}

func loginHandler(c *gin.Context) {
	var login struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := c.BindJSON(&login)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Find user by username
	user, err := FindUserByUsername(login.Username)
	if err != nil {
		c.JSON(401, gin.H{"error": "Invalid username or password"})
		return
	}

	if user == nil {
		c.JSON(401, gin.H{"error": "Invalid username or password"})
		return
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid username or password"})
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user.ID,
	})
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{"success": true, "token": tokenString})
}

func authenticatedHandler(c *gin.Context) {
  // Get the token from the Authorization header
  tokenString := c.GetHeader("Authorization")

  // If the token is empty, return an error
  if tokenString == "" {
    c.JSON(401, gin.H{"error": "Unauthorized"})
    return
  }

  // Parse the token
  token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return []byte("your-secret-key"), nil
  })

  // If the token is invalid, return an error
  if err != nil {
    c.JSON(401, gin.H{"error": "Invalid token"})
    return
  }

  // If the token is valid, check if it contains the "user" claim
  if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
    if _, ok := claims["user"]; !ok {
      c.JSON(401, gin.H{"error": "Invalid token"})
      return
    }
  } else {
    c.JSON(401, gin.H{"error": "Invalid token"})
    return
  }

  // If the token is valid and contains the "user" claim, return success
  c.JSON(200, gin.H{"authenticated": true})
}

func getEntries(c *gin.Context) {
	rows, err := db.Query("SELECT id, date, rating, note, tags FROM entries ORDER BY date DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		var dateStr string
		err := rows.Scan(&e.ID, &dateStr, &e.Rating, &e.Note, &e.Tags)
		if err != nil {
			continue
		}
		e.Date, _ = time.Parse(time.RFC3339, dateStr)
		entries = append(entries, e)
	}
	c.JSON(http.StatusOK, entries)
}

func getEntry(c *gin.Context) {
	id := c.Param("id")
	row := db.QueryRow("SELECT id, date, rating, note, tags FROM entries WHERE id = ?", id)

	var e Entry
	var dateStr string
	err := row.Scan(&e.ID, &dateStr, &e.Rating, &e.Note, &e.Tags)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entry not found"})
		return
	}
	e.Date, _ = time.Parse(time.RFC3339, dateStr)
	c.JSON(http.StatusOK, e)
}

func createEntry(c *gin.Context) {
	var e Entry
	if err := c.BindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if e.Date.IsZero() {
		e.Date = time.Now()
	}

	_, err := db.Exec(
		"INSERT INTO entries (date, rating, note, tags) VALUES (?, ?, ?, ?)",
		e.Date.Format(time.RFC3339), e.Rating, e.Note, e.Tags,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func updateEntry(c *gin.Context) {
	id := c.Param("id")
	var e Entry
	if err := c.BindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := db.Exec(
		"UPDATE entries SET date = ?, rating = ?, note = ?, tags = ? WHERE id = ?",
		e.Date.Format(time.RFC3339), e.Rating, e.Note, e.Tags, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func deleteEntry(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM entries WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
