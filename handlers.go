package main

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

// CaptchaTemplateData embeds PageData and adds a CSRFField for our form.
type CaptchaTemplateData struct {
	PageData
	CSRFField template.HTML
}

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

// showQueueInternal renders the queue page.
func showQueueInternal(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	// Ensure session is initialized.
	if session.Values["flow_stage"] == nil {
		session.Values["flow_stage"] = "queue"
		session.Values["start_time"] = strconv.FormatInt(time.Now().Unix(), 10)
		session.Save(r, w)
	}
	if err := templates.ExecuteTemplate(w, "queue.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// showCaptchaInternal handles both GET and POST for the captcha page.
func showCaptchaInternal(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	// Ensure waiting period has passed.
	startTimeStr, ok := session.Values["start_time"].(string)
	if !ok {
		showQueueInternal(w, r, session)
		return
	}
	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil || time.Now().Unix()-startTime < 20 {
		showQueueInternal(w, r, session)
		return
	}

	// Set stage to captcha.
	session.Values["flow_stage"] = "captcha"
	session.Save(r, w)

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
			session.Save(r, w)
			tplData := CaptchaTemplateData{
				PageData:  data,
				CSRFField: csrf.TemplateField(r),
			}
			if err := templates.ExecuteTemplate(w, "captcha.html", tplData); err != nil {
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
			session.Save(r, w)
			showAssignInternal(w, r, session)
			return
		}
		// Incorrect answer: prepare new view data with an error.
		data := prepareViewData(session)
		data.Message = "Incorrect answer. Please try again"
		session.Values["view_data"] = data
		session.Save(r, w)
		tplData := CaptchaTemplateData{
			PageData:  data,
			CSRFField: csrf.TemplateField(r),
		}
		if err := templates.ExecuteTemplate(w, "captcha.html", tplData); err != nil {
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
	session.Save(r, w)
	tplData := CaptchaTemplateData{
		PageData:  data,
		CSRFField: csrf.TemplateField(r),
	}
	if err := templates.ExecuteTemplate(w, "captcha.html", tplData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// showAssignInternal displays the assignment page to users who have solved the captcha.
func showAssignInternal(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	if solved, ok := session.Values["captcha_solved"].(string); !ok || solved != "true" || session.Values["flow_stage"] != "assign" {
		session.Save(r, w)
		showCaptchaInternal(w, r, session)
		return
	}
	session.Save(r, w)
	if err := templates.ExecuteTemplate(w, "assign.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

