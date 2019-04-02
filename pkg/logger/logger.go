// logger provides with a simple interface for logging so
// that our application logs are well structured and it is easy
// to process them with tools such as Elastic search
package logger

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// LogHttpEntry a log entry that captures the information
// we are interested in about a http response
type HttpLogEntry struct {
	Severity string `json:"severity"`
	Method   string `json:"method"`
	Uri      string `json:"uri"`
	Time     int    `json:"ts"`
	Usec     int    `json:"usec"`
	Status   int    `json:"status"`
}

// LogGenericEntry is a generic log entry structure
// so that we can use this library from everywhere in our
// code and we keep application logs more or less consistent
// and structured
type GenericLogEntry struct {
	Severity string      `json:"severity"`
	Msg      string      `json:"msg"`
	Time     int         `json:"ts"`
	Data     interface{} `json:"data"`
}

// Info logs a structured log info entry with
// some extra data
func Info(msg string, data interface{}) {
	write(&GenericLogEntry{
		Severity: "info",
		Msg:      msg,
		Time:     millis(),
		Data:     data,
	})
}

// Error logs a structured log error entry
func Error(err error) {
	write(&GenericLogEntry{
		Severity: "error",
		Msg:      err.Error(),
		Time:     millis(),
	})
}

// statusWriter is a simple wrapper
// that helps us capture the http response status and content-length
type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

// WriteHeader mimic the http.ResponseWriter protocol
func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// WriteHeader mimic the http.ResponseWriter protocol
func (w *statusWriter) Write(b []byte) (int, error) {

	// Set a default http status code
	if w.status == 0 {
		w.status = http.StatusOK
	}

	// capture the content-length
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

// NewHttpLogger returns a simple middleware that produces
// structured logs from our http requests. This middleware also
// inspects the current request for extra errors (eg. database errors)
// that should be printed in the logs too
func NewHttpLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			// wrap our response write, so that we are able
			// to intercept the status code and the content length
			wrapper := &statusWriter{
				ResponseWriter: w,
			}

			// Define our log function to be called in
			// deferred mode, so that this gets logged even
			// if a panic ocurrs
			defer func() {

				severity := "info"
				if wrapper.status >= 500 {
					severity = "error"
				}

				write(&HttpLogEntry{
					Severity: severity,
					Time:     millis(),
					Method:   r.Method,
					Uri:      r.RequestURI,
					Usec:     int(time.Now().Sub(start) / time.Microsecond),
					Status:   wrapper.status,

					// Add more fields, such as user agent
					// peer ip/port etc..

				})
			}()
			next.ServeHTTP(wrapper, r)
		})
	}
}

// write is a convenience function that marshalls into json
// or fallbacks to log.Printf
func write(e interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.Encode(e)
}

// millis returns the current time in milliseconds
func millis() int {
	now := time.Now()
	return int(now.UnixNano() / 1000000)
}
