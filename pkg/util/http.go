package util

import (
	"github.com/go-chi/render"
	"github.com/pedro-gutierrez/form3/pkg/logger"
	"net/http"
)

// EmptyResponse represents an empty JSON response
type EmptyResponse struct{}

// HttpError handles a http error by returning a JSON response
// with the appropiate status code, and logging the root cause to the console
func HandleHttpError(w http.ResponseWriter, r *http.Request, status int, err error) {
	RenderJSON(w, r, status, &EmptyResponse{})

	// Our middleware is going to log the response
	// but we complete with more info incase we have a 5xx kind of error
	if status >= http.StatusInternalServerError {
		logger.Error(err)
	}

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
