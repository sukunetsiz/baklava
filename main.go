package main

import (
    "log"
    "math/rand"
    "net/http"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

func main() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    http.HandleFunc("/", showQueue)
    http.HandleFunc("/captcha", showCaptcha)
    http.HandleFunc("/verify", verifyCaptcha)
    http.HandleFunc("/assign", showAssign)

    log.Println("Server started on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
