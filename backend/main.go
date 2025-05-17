
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("DB error:", err)
	}
	defer db.Close()

	r := gin.Default()
	r.Use(gin.Logger())

	r.POST("/auth", handleAuth)
	r.GET("/events", getEvents)
	r.POST("/events", createEvent)
	r.POST("/invite", createInvite)
	r.POST("/join", acceptInvite)

	r.Run(":8080")
}

func handleAuth(c *gin.Context) {
	var req struct {
		TelegramID int64 `json:"telegram_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	db.Exec("INSERT INTO users (telegram_id) VALUES ($1) ON CONFLICT DO NOTHING", req.TelegramID)

	var babyID int
	err := db.QueryRow("SELECT baby_id FROM baby_access WHERE telegram_id = $1 LIMIT 1", req.TelegramID).Scan(&babyID)
	if err == sql.ErrNoRows {
		tx, _ := db.Begin()
		tx.QueryRow("INSERT INTO babies DEFAULT VALUES RETURNING baby_id").Scan(&babyID)
		tx.Exec("INSERT INTO baby_access (baby_id, telegram_id, role) VALUES ($1, $2, 'admin')", babyID, req.TelegramID)
		tx.Commit()
	} else if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}

	c.JSON(200, gin.H{"baby_id": babyID})
}

func getEvents(c *gin.Context) {
	tgid := c.Query("telegram_id")
	rows, err := db.Query(`
		SELECT e.type, e.time_str, e.timestamp, e.event_id
		FROM events e
		JOIN baby_access ba ON ba.baby_id = e.baby_id
		WHERE ba.telegram_id = $1
		ORDER BY e.timestamp DESC
	`, tgid)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	var out []gin.H
	for rows.Next() {
		var t, s string
		var ts int64
		var id int
		rows.Scan(&t, &s, &ts, &id)
		out = append(out, gin.H{"type": t, "time_str": s, "timestamp": ts, "id": id})
	}
	c.JSON(200, out)
}

func createEvent(c *gin.Context) {
	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Type       string `json:"type"`
		TimeStr    string `json:"time_str"`
		Timestamp  int64  `json:"timestamp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}

	var babyID int
	db.QueryRow("SELECT baby_id FROM baby_access WHERE telegram_id = $1 LIMIT 1", req.TelegramID).Scan(&babyID)

	_, err := db.Exec("INSERT INTO events (baby_id, type, time_str, timestamp) VALUES ($1,$2,$3,$4)", babyID, req.Type, req.TimeStr, req.Timestamp)
	if err != nil {
		c.JSON(500, gin.H{"error": "insert error"})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func createInvite(c *gin.Context) {
	var req struct {
		TelegramID int64 `json:"telegram_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	var babyID int
	db.QueryRow("SELECT baby_id FROM baby_access WHERE telegram_id = $1 LIMIT 1", req.TelegramID).Scan(&babyID)
	code := randomCode(6)
	_, err := db.Exec("INSERT INTO invites (code, baby_id) VALUES ($1,$2) ON CONFLICT (code) DO NOTHING", code, babyID)
	if err != nil {
		c.JSON(500, gin.H{"error": "db error"})
		return
	}
	c.JSON(200, gin.H{"code": code})
}

func acceptInvite(c *gin.Context) {
	var req struct {
		TelegramID int64  `json:"telegram_id"`
		Code       string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "bad request"})
		return
	}
	var babyID int
	err := db.QueryRow("SELECT baby_id FROM invites WHERE code = $1", req.Code).Scan(&babyID)
	if err != nil {
		c.JSON(404, gin.H{"error": "invalid code"})
		return
	}
	db.Exec("INSERT INTO baby_access (baby_id, telegram_id, role) VALUES ($1,$2,'parent') ON CONFLICT DO NOTHING", babyID, req.TelegramID)
	c.JSON(200, gin.H{"status": "joined"})
}

func randomCode(n int) string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
