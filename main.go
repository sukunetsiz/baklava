package main

import (
	"encoding/gob"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
)

// Global session store.
// Instead of a CookieStore, we’re now using a FilesystemStore.
// The first parameter is the directory where session files will be stored.
// Passing an empty string ("") makes it use os.TempDir().
var store *sessions.FilesystemStore

func init() {
	// Seed the random generator.
	rand.Seed(time.Now().UnixNano())

	// Register types that will be stored in sessions.
	gob.Register(PageData{})
	gob.Register(CoordinateExample{})
	gob.Register(Cell{})
	gob.Register(CellStyle{})

	// Retrieve the session key from the environment.
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("SESSION_KEY environment variable not set")
	}

	// Create a new FilesystemStore.
	store = sessions.NewFilesystemStore("", []byte(sessionKey))
	// Optionally, remove any maximum length restrictions:
	store.MaxLength(0)
}

func main() {
	// Serve static files (e.g. CSS, JS, images) from the "static" folder.
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Define route handlers.
	http.HandleFunc("/", showQueue)
	http.HandleFunc("/captcha", showCaptcha)
	http.HandleFunc("/assign", showAssign)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

