package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func InitDB(fp string) *sql.DB {
	log.Printf("Creating database at %s", fp)
	parentDir := filepath.Dir(fp)
	err := os.MkdirAll(parentDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		f, err := os.Create(fp)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		f.Close()
	}

	db, err := sql.Open("sqlite3", fp)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		rating INTEGER,
		note TEXT,
		tags TEXT
	);
	`
	createTableUsers := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		);
	`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec(createTableUsers)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// Insert default user if it doesn't exist
	defaultUser := os.Getenv("DEFAULT_USERNAME")
	defaultPassword :=  os.Getenv("DEFAULT_PASSWORD")

	if defaultUser == "" || defaultPassword == "" {
		return db	
	}
	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Check if the user already exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", defaultUser).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if user exists: %v", err)
	}

	if !exists {
		// Insert the default user
		_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", defaultUser, hashedPassword)
		if err != nil {
			log.Fatalf("Failed to insert default user: %v", err)
		}
	}

	return db
}

func FindUserByUsername(username string) (*User, error) {
  var user User
  row := db.QueryRow("SELECT * FROM users WHERE username = ?", username)

  err := row.Scan(&user.ID, &user.Username, &user.Password)
  if err != nil {
    if err == sql.ErrNoRows {
      return nil, nil
    }
    return nil, err
  }

  return &user, nil
}
