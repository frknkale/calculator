// Package httpapi translates HTTP requests into calculator calls.
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"calculator/backend/internal/calculator"
)

type request struct {
	Operation string   `json:"operation"`
	A         *float64 `json:"a"`
	B         *float64 `json:"b,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewHandler returns the API routes and optional frontend handler.
func NewHandler(frontend http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/calculate", calculate)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Use GET for this endpoint.")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "API endpoint not found.")
	})
	if frontend != nil {
		mux.Handle("/", frontend)
	}
	return mux
}

func calculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Use POST for this endpoint.")
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Send Content-Type: application/json.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var input request
	if err := decoder.Decode(&input); err != nil {
		invalidJSON(w, err)
		return
	}
	// A second value (or malformed trailing data) must not be silently ignored.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		invalidJSON(w, err)
		return
	}
	result, err := calculator.Calculate(input.Operation, input.A, input.B)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_calculation", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"result": result})
}

func invalidJSON(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "body_too_large", "Request body must not exceed 4096 bytes.")
		return
	}
	writeError(w, http.StatusBadRequest, "invalid_json", "Send one JSON object with operation and numeric a/b fields only.")
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, struct {
		Error apiError `json:"error"`
	}{Error: apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	// Response values contain only strings and finite numbers.
	_ = json.NewEncoder(w).Encode(value)
}
