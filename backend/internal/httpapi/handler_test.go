package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalculateEndpoint(t *testing.T) {
	tests := []struct {
		name, body string
		status     int
		code       string
		result     float64
	}{
		{"add", `{"operation":"add","a":2,"b":3}`, 200, "", 5},
		{"subtract", `{"operation":"subtract","a":2,"b":3}`, 200, "", -1},
		{"multiply", `{"operation":"multiply","a":-2,"b":3.5}`, 200, "", -7},
		{"divide", `{"operation":"divide","a":7,"b":2}`, 200, "", 3.5},
		{"power", `{"operation":"power","a":2,"b":10}`, 200, "", 1024},
		{"sqrt", `{"operation":"sqrt","a":9}`, 200, "", 3},
		{"percentage", `{"operation":"percentage","a":200,"b":25}`, 200, "", 50},
		{"percentage zero rate", `{"operation":"percentage","a":200,"b":0}`, 200, "", 0},
		{"percentage missing rate", `{"operation":"percentage","a":200}`, 400, "invalid_calculation", 0},
		{"percentage null rate", `{"operation":"percentage","a":200,"b":null}`, 400, "invalid_calculation", 0},
		{"explicit zero", `{"operation":"add","a":0,"b":0}`, 200, "", 0},
		{"scientific notation", `{"operation":"add","a":1e2,"b":2}`, 200, "", 102},
		{"divide zero", `{"operation":"divide","a":2,"b":0}`, 400, "invalid_calculation", 0},
		{"negative root", `{"operation":"sqrt","a":-9}`, 400, "invalid_calculation", 0},
		{"overflow", `{"operation":"power","a":10,"b":309}`, 400, "invalid_calculation", 0},
		{"unknown operation", `{"operation":"unknown","a":1}`, 400, "invalid_calculation", 0},
		{"missing a", `{"operation":"add","b":2}`, 400, "invalid_calculation", 0},
		{"missing b", `{"operation":"add","a":2}`, 400, "invalid_calculation", 0},
		{"null a", `{"operation":"add","a":null,"b":2}`, 400, "invalid_calculation", 0},
		{"unary extra operand", `{"operation":"sqrt","a":9,"b":2}`, 400, "invalid_calculation", 0},
		{"null body", `null`, 400, "invalid_calculation", 0},
		{"string operand", `{"operation":"add","a":"2","b":3}`, 400, "invalid_json", 0},
		{"boolean operand", `{"operation":"sqrt","a":true}`, 400, "invalid_json", 0},
		{"out of range input", `{"operation":"sqrt","a":1e400}`, 400, "invalid_json", 0},
		{"invalid JSON", `{`, 400, "invalid_json", 0},
		{"empty body", ``, 400, "invalid_json", 0},
		{"array body", `[]`, 400, "invalid_json", 0},
		{"unknown fields", `{"operation":"sqrt","a":9,"typo":2}`, 400, "invalid_json", 0},
		{"multiple objects", `{"operation":"sqrt","a":9} {}`, 400, "invalid_json", 0},
		{"trailing garbage", `{"operation":"sqrt","a":9} xyz`, 400, "invalid_json", 0},
		{"oversized body", `{"operation":"` + strings.Repeat("x", 4096) + `"}`, 413, "body_too_large", 0},
		{"oversized trailing whitespace", `{"operation":"sqrt","a":9}` + strings.Repeat(" ", 4096), 413, "body_too_large", 0},
	}
	handler := NewHandler(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/calculate", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tt.status, response.Body)
			}
			if response.Header().Get("Content-Type") != "application/json" {
				t.Fatal("response must be JSON")
			}
			var body struct {
				Result float64  `json:"result"`
				Error  apiError `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tt.status == 200 && body.Result != tt.result {
				t.Errorf("result = %v, want %v", body.Result, tt.result)
			}
			if tt.code != "" && (body.Error.Code != tt.code || body.Error.Message == "") {
				t.Errorf("error = %+v, want code %s and a message", body.Error, tt.code)
			}
		})
	}
}

func TestRouting(t *testing.T) {
	frontend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("frontend"))
	})
	handler := NewHandler(frontend)
	for _, tt := range []struct {
		method, path, contentType string
		status                    int
		allow                     string
	}{
		{"GET", "/api/health", "", 200, ""},
		{"POST", "/api/health", "", 405, "GET"},
		{"GET", "/api/calculate", "", 405, "POST"},
		{"PUT", "/api/calculate", "", 405, "POST"},
		{"POST", "/api/calculate", "", 415, ""},
		{"POST", "/api/calculate", "text/plain", 415, ""},
		{"GET", "/api/missing", "", 404, ""},
		{"GET", "/", "", 200, ""},
	} {
		t.Run(tt.method+tt.path+tt.contentType, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", tt.contentType)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tt.status || response.Header().Get("Allow") != tt.allow {
				t.Errorf("status/allow = %d/%q, want %d/%q", response.Code, response.Header().Get("Allow"), tt.status, tt.allow)
			}
			if tt.path == "/api/health" && tt.status == 200 && response.Body.String() != "{\"status\":\"ok\"}\n" {
				t.Errorf("unexpected health response: %s", response.Body)
			}
			if tt.path == "/" && response.Body.String() != "frontend" {
				t.Errorf("frontend not served: %s", response.Body)
			}
		})
	}
}
