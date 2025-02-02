package main

import (
    "regexp"
    "strings"
)

var coordRegex = regexp.MustCompile(`^[1-8]-[1-8]$`)

func formatCoordinateAnswer(answer string) string {
    return strings.ReplaceAll(answer, " ", "")
}

func validateCoordinates(answer string) bool {
    return coordRegex.MatchString(answer)
}
