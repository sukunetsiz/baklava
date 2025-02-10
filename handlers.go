package main

import (
	"html/template"
	"net/http"
	"strconv"
	"time"
)

func init() {
	// Define template helper functions.
	funcMap := template.FuncMap{
		"iterate": func(start, end int) []int {
			var result []int
			for i := start; i < end; i++ {
				result = append(result, i)
			}
			return result
		},
		"add": func(a, b int) int { return a + b },
	}

	// Parse the templates.
	templates = template.Must(template.New("").Funcs(funcMap).ParseFiles(
		"templates/queue.html",
		"templates/captcha.html",
		"templates/assign.html",
	))
}

// showQueue displays the waiting queue page and initializes the session.
func showQueue(w http.ResponseWriter, r *http.Request) {
	// Retrieve the session (named "captcha-session").
	session, err := store.Get(r, "captcha-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If the session already has a stage, redirect appropriately.
	if stage, ok := session.Values["flow_stage"].(string); ok {
		if stage != "queue" {
			if stage == "captcha" {
				http.Redirect(w, r, "/captcha", http.StatusSeeOther)
				return
			} else if stage == "assign" {
				http.Redirect(w, r, "/assign", http.StatusSeeOther)
				return
			}
		}
	} else {
		// No session data yet: initialize it.
		session.Values["flow_stage"] = "queue"
		session.Values["start_time"] = strconv.FormatInt(time.Now().Unix(), 10)
	}

	// Save the session.
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Render the "queue" template.
	if err := templates.ExecuteTemplate(w, "queue.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// showCaptcha handles both GET and POST for the captcha page.
func showCaptcha(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "captcha-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure the user has waited at least 20 seconds in the queue.
	startTimeStr, ok := session.Values["start_time"].(string)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil || time.Now().Unix()-startTime < 20 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Mark that the user is now in the captcha stage.
	session.Values["flow_stage"] = "captcha"

	// Handle POST: processing the captcha answer.
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		answer := r.FormValue("captcha_answer")
		if !validateCoordinates(answer) {
			// Prepare view data with an error message.
			data := prepareViewData(session)
			data.Message = "Please enter coordinates in correct format"
			session.Values["view_data"] = data
			if err := session.Save(r, w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err := templates.ExecuteTemplate(w, "captcha.html", data); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		formattedAnswer := formatCoordinateAnswer(answer)
		if formattedAnswer == session.Values["captcha_answer"] {
			session.Values["captcha_solved"] = "true"
			session.Values["flow_stage"] = "assign"
			// Clean up captcha-related session values.
			delete(session.Values, "captcha_answer")
			delete(session.Values, "game_letter")
			delete(session.Values, "view_data")
			if err := session.Save(r, w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/assign", http.StatusSeeOther)
			return
		}
		// If the answer is incorrect, prepare new view data with an error.
		data := prepareViewData(session)
		data.Message = "Incorrect answer. Please try again"
		session.Values["view_data"] = data
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := templates.ExecuteTemplate(w, "captcha.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// For GET requests, display the captcha.
	var data PageData
	if v, exists := session.Values["view_data"]; exists {
		if d, ok := v.(PageData); ok {
			data = d
		} else {
			data = prepareViewData(session)
			session.Values["view_data"] = data
		}
	} else {
		data = prepareViewData(session)
		session.Values["view_data"] = data
	}

	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "captcha.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// showAssign displays the assignment page to users who have solved the captcha.
func showAssign(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "captcha-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if solved, ok := session.Values["captcha_solved"].(string); !ok || solved != "true" || session.Values["flow_stage"] != "assign" {
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/captcha", http.StatusSeeOther)
		return
	}
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := templates.ExecuteTemplate(w, "assign.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

