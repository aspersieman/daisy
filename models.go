package main

import (
	"time"
)

type Entry struct {
	ID     int       `json:"id"`
	Date   time.Time `json:"date"`
	Rating int       `json:"rating"`
	Note   string    `json:"note"`
	Tags   string    `json:"tags"` // comma-separated string
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}
