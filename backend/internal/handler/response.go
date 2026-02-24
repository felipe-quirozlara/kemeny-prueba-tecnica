package handler

import (
    "encoding/json"
    "log"
    "net/http"
)

// ErrorResponse is the JSON shape returned to clients on errors.
type ErrorResponse struct {
    Error string `json:"error"`
    Code  int    `json:"code,omitempty"`
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data interface{}) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(data)
}

// Error logs internal error details (if provided) and returns a safe JSON error
// message to the client. Avoids exposing internal error strings to clients.
func Error(w http.ResponseWriter, r *http.Request, status int, clientMsg string, err error, code int) {
    if err != nil {
        log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
    }
    resp := ErrorResponse{Error: clientMsg}
    if code != 0 {
        resp.Code = code
    }
    // Best-effort encode; if encoding fails, fall back to a plain status.
    if encodeErr := JSON(w, status, resp); encodeErr != nil {
        log.Printf("failed to write error response: %v", encodeErr)
        http.Error(w, http.StatusText(status), status)
    }
}
