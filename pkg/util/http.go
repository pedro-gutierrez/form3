package util

import (
	"encoding/json"
	"fmt"
	"github.com/pedro-gutierrez/form3/pkg/logger"
	"net/http"
)

// HttpService is a simple base type for Http services
// Provides with some convenience functions that can be
// reused by more concrete implementations
type HttpService struct {
	BaseUrl string
}

// urlFor builds a new url by prefixing the given
// path with the base url configured for the service
func (s *HttpService) UrlFor(path string) string {
	return fmt.Sprintf("%s%s", s.BaseUrl, path)
}

// EmptyResponse represents an empty JSON response
type EmptyResponse struct{}

// HttpError handles a http error by returning a JSON response
// with the appropiate status code, and logging the root cause to the console
func HandleHttpError(w http.ResponseWriter, r *http.Request, status int, err error) {
	RenderJSON(w, r, status, &EmptyResponse{})

	// Our middleware is going to log the response
	// but we complete with more info incase we have a 5xx kind of error
	logger.Error(err)
}

// RenderJSON is a convenience function that marshalls the given interface
// value as json, with the given http status code
func RenderJSON(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	err := enc.Encode(data)
	if err != nil {
		logger.Error(err)
	}
}

// RenderNoContent returns a 204 and an empty response body
func RenderNoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
