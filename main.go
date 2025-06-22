package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var db *sql.DB

func main() {
	db = InitDB("daisy.db")

	r := gin.Default()

	r.GET("/entries", getEntries)
	r.GET("/entries/:id", getEntry)
	r.POST("/entries", createEntry)
	r.PUT("/entries/:id", updateEntry)
	r.DELETE("/entries/:id", deleteEntry)

	r.Run(":8069")
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
