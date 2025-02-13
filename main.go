package main

import (
	"encoding/gob"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

// Global session store.
var store *sessions.FilesystemStore

// Global keys.
var (
	sessionKey string
	csrfKey    string
)

func init() {
	// Load environment variables from .env file if it exists.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; proceeding with system environment variables")
	}

	// Seed the random generator.
	rand.Seed(time.Now().UnixNano())

	// Register types stored in sessions.
	gob.Register(PageData{})
	gob.Register(CoordinateExample{})
	gob.Register(Cell{})
	gob.Register(CellStyle{})

	// Retrieve the session key from the environment.
	sessionKey = os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("SESSION_KEY environment variable not set")
	}

	// Retrieve the CSRF key from the environment.
	csrfKey = os.Getenv("CSRF_KEY")
	if csrfKey == "" {
		log.Println("Warning: CSRF_KEY environment variable not set; falling back to SESSION_KEY")
		csrfKey = sessionKey
	}

	// Create a new FilesystemStore.
	store = sessions.NewFilesystemStore("", []byte(sessionKey))
	store.MaxLength(0)
}

// securityHeadersMiddleware adds important security headers to all responses.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline';")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Create a new ServeMux.
	mux := http.NewServeMux()

	// Serve static files (e.g. CSS, JS, images) from the "static" folder.
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// Define route handlers.
	mux.HandleFunc("/", showQueue)
	mux.HandleFunc("/captcha", showCaptcha)
	mux.HandleFunc("/assign", showAssign)

	// For local development over HTTP, disable the "Secure" flag.
	secureFlag := true
	if os.Getenv("DEV") != "" {
		secureFlag = false
	}

	// Wrap our mux with the CSRF middleware.
	csrfMiddleware := csrf.Protect([]byte(csrfKey), csrf.Secure(secureFlag))
	// Wrap with our security headers middleware as well.
	handler := securityHeadersMiddleware(csrfMiddleware(mux))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

