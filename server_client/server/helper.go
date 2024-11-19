package server

import (
	"net/http"
	"strings"
)

func getPath(path string) string {
	return "GET " + path
}

func postPath(path string) string {
	return "POST " + path
}

func writeResponse(resp http.ResponseWriter, statusCode int, explanations ...string) {
	resp.WriteHeader(statusCode)
	resp.Write([]byte(http.StatusText(statusCode)))

	if len(explanations) != 0 {
		resp.Write([]byte("\n"))
		fullText := strings.Join(explanations, "\n")
		resp.Write([]byte(fullText))
	}
}

