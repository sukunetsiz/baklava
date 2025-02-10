package main

import (
    "encoding/json"
    "html/template"
    "net/http"
    "strconv"
    "time"
)

// Note: For simplicity, we assume that 'sessions' is a global variable,
// for example: var sessions = map[string]string{}
// In production, you should use a proper per-user session management system.

func init() {
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

    templates = template.Must(template.New("").Funcs(funcMap).ParseFiles(
        "templates/queue.html",
        "templates/captcha.html",
        "templates/assign.html",
    ))
}

// showQueue displays the waiting queue page.
// It also initializes the session state if not already set.
func showQueue(w http.ResponseWriter, r *http.Request) {
    // If the session already has a stage, ensure the user is not jumping back.
    if stage, ok := sessions["flow_stage"]; ok {
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
        // No session data yet: initialize the session.
        sessions["flow_stage"] = "queue"
        sessions["start_time"] = strconv.FormatInt(time.Now().Unix(), 10)
    }

    if err := templates.ExecuteTemplate(w, "queue.html", nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// showCaptcha handles both displaying the captcha page (GET)
// and processing the captcha answer (POST).
func showCaptcha(w http.ResponseWriter, r *http.Request) {
    // Ensure that the user has waited 20 seconds in the queue.
    startTimeStr, ok := sessions["start_time"]
    if !ok {
        // No queue time recorded, force user to go to queue.
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
    if err != nil || time.Now().Unix()-startTime < 20 {
        // Either an error or the user hasn’t waited long enough.
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }
    // Mark that the user is now in the captcha stage.
    sessions["flow_stage"] = "captcha"

    // Handle form submission (POST)
    if r.Method == http.MethodPost {
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Error parsing form", http.StatusBadRequest)
            return
        }
        answer := r.FormValue("captcha_answer")
        if !validateCoordinates(answer) {
            data := prepareViewData()
            data.Message = "Please enter coordinates in correct format"
            if jsonData, err := json.Marshal(data); err == nil {
                sessions["view_data"] = string(jsonData)
            }
            if err := templates.ExecuteTemplate(w, "captcha.html", data); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
            }
            return
        }
        formattedAnswer := formatCoordinateAnswer(answer)
        if formattedAnswer == sessions["captcha_answer"] {
            sessions["captcha_solved"] = "true"
            // Advance the flow to the assignment stage.
            sessions["flow_stage"] = "assign"
            delete(sessions, "captcha_answer")
            delete(sessions, "game_letter")
            delete(sessions, "view_data")
            http.Redirect(w, r, "/assign", http.StatusSeeOther)
            return
        }
        // If the answer was incorrect, generate a new captcha with an error message.
        data := prepareViewData()
        data.Message = "Incorrect answer. Please try again"
        if jsonData, err := json.Marshal(data); err == nil {
            sessions["view_data"] = string(jsonData)
        }
        if err := templates.ExecuteTemplate(w, "captcha.html", data); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    // For GET requests, display the captcha.
    var data PageData
    if viewDataJSON, exists := sessions["view_data"]; exists {
        if err := json.Unmarshal([]byte(viewDataJSON), &data); err != nil {
            data = prepareViewData()
            if jsonData, err := json.Marshal(data); err == nil {
                sessions["view_data"] = string(jsonData)
            }
        }
    } else {
        data = prepareViewData()
        if jsonData, err := json.Marshal(data); err == nil {
            sessions["view_data"] = string(jsonData)
        }
    }

    if err := templates.ExecuteTemplate(w, "captcha.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

// showAssign displays the assignment page.
// Only users who have successfully solved the captcha (and thus are in the "assign" stage)
// are allowed to see it.
func showAssign(w http.ResponseWriter, r *http.Request) {
    if solved, ok := sessions["captcha_solved"]; !ok || solved != "true" || sessions["flow_stage"] != "assign" {
        http.Redirect(w, r, "/captcha", http.StatusSeeOther)
        return
    }
    if err := templates.ExecuteTemplate(w, "assign.html", nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

