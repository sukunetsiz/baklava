package main

import (
	"encoding/gob"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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

const idleTimeout = 300 // Idle timeout in seconds (5 minutes)

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
	mux := http.NewServeMux()

	// Serve static files from the "static" folder.
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// All routes are handled by mainHandler at "/"
	mux.HandleFunc("/", mainHandler)

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

// mainHandler checks session state, idle timeout, and waiting time so that only "/" is shown in the address bar.
func mainHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect any URL that is not "/" to "/"
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Attempt to retrieve the session.
	session, err := store.Get(r, "captcha-session")
	if err != nil {
		// Session retrieval failed; create a new session without setting an error message.
		session, _ = store.New(r, "captcha-session")
		pageData := prepareViewData(session)
		session.Values["view_data"] = pageData
		session.Values["flow_stage"] = "captcha"
		session.Values["start_time"] = strconv.FormatInt(time.Now().Unix(), 10)
		session.Save(r, w)
		showCaptchaInternal(w, r, session)
		return
	}

	// Idle timeout check: if no activity within idleTimeout seconds, reset the session.
	if lastActiveStr, ok := session.Values["last_active"].(string); ok {
		if lastActive, err := strconv.ParseInt(lastActiveStr, 10, 64); err == nil {
			if time.Now().Unix()-lastActive > idleTimeout {
				// Session idle timeout exceeded, clear session and create a new one without an error message.
				session.Options.MaxAge = -1 // Mark session for deletion.
				session.Save(r, w)
				session, _ = store.New(r, "captcha-session")
				pageData := prepareViewData(session)
				session.Values["view_data"] = pageData
				session.Values["flow_stage"] = "captcha"
				session.Values["start_time"] = strconv.FormatInt(time.Now().Unix(), 10)
				session.Save(r, w)
				showCaptchaInternal(w, r, session)
				return
			}
		}
	}

	// Update last_active timestamp for every request.
	session.Values["last_active"] = strconv.FormatInt(time.Now().Unix(), 10)

	// If session is new, initialize as queue.
	stage, ok := session.Values["flow_stage"].(string)
	if !ok {
		session.Values["flow_stage"] = "queue"
		session.Values["start_time"] = strconv.FormatInt(time.Now().Unix(), 10)
		stage = "queue"
	}

	// If in queue, check if waiting period is over.
	if stage == "queue" {
		if startTimeStr, ok := session.Values["start_time"].(string); ok {
			startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
			if err == nil && time.Now().Unix()-startTime >= 20 {
				session.Values["flow_stage"] = "captcha"
				stage = "captcha"
			}
		}
	}

	session.Save(r, w)

	switch stage {
	case "queue":
		showQueueInternal(w, r, session)
	case "captcha":
		showCaptchaInternal(w, r, session)
	case "assign":
		showAssignInternal(w, r, session)
	default:
		showQueueInternal(w, r, session)
	}
}
