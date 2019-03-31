package util

import (
	"github.com/go-chi/render"
	"log"
	"net/http"
)

// EmptyResponse represents an empty JSON response
type EmptyResponse struct{}

// HttpError handles a http error by returning a JSON response
// with the appropiate status code, and logging the root cause to the console
func HandleHttpError(w http.ResponseWriter, r *http.Request, status int, err error) {
	RenderJSON(w, r, status, &EmptyResponse{})
	log.Printf(err.Error())
}

// RenderJSON is a convenience function that marshalls the given interface
// value as json, with the given http status code
func RenderJSON(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	render.Status(r, status)
	render.JSON(w, r, data)
}

// RenderNoContent returns a 204 and an empty response body
func RenderNoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
