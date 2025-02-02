package main

import (
    "encoding/json"
    "html/template"
    "net/http"
)

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

func showQueue(w http.ResponseWriter, r *http.Request) {
    if err := templates.ExecuteTemplate(w, "queue.html", nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func showCaptcha(w http.ResponseWriter, r *http.Request) {
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

    if msg := r.URL.Query().Get("message"); msg != "" {
        data.Message = msg
    }

    if err := templates.ExecuteTemplate(w, "captcha.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func verifyCaptcha(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Error parsing form", http.StatusBadRequest)
        return
    }

    answer := r.FormValue("captcha_answer")
    if !validateCoordinates(answer) {
        http.Redirect(w, r, "/captcha?message=Please+enter+coordinates+in+correct+format", http.StatusSeeOther)
        return
    }

    formattedAnswer := formatCoordinateAnswer(answer)
    if formattedAnswer == sessions["captcha_answer"] {
        sessions["captcha_solved"] = "true"
        delete(sessions, "captcha_answer")
        delete(sessions, "game_letter")
        delete(sessions, "view_data")
        http.Redirect(w, r, "/assign", http.StatusSeeOther)
        return
    }

    http.Redirect(w, r, "/captcha?message=Incorrect+answer.+Please+try+again", http.StatusSeeOther)
}

func showAssign(w http.ResponseWriter, r *http.Request) {
    if err := templates.ExecuteTemplate(w, "assign.html", nil); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
