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
		name, body   string
		status       int
		code, result string
	}{
		{"add", `{"operation":"add","a":"2","b":"3"}`, 200, "", "5"},
		{"exact decimal", `{"operation":"add","a":"0.1","b":"0.2"}`, 200, "", "0.3"},
		{"large decimal", `{"operation":"add","a":"1e25","b":"0.2"}`, 200, "", "1e+25"},
		{"large integer", `{"operation":"add","a":"9007199254740992","b":"1"}`, 200, "", "9007199254740993"},
		{"subtract", `{"operation":"subtract","a":"2","b":"3"}`, 200, "", "-1"},
		{"multiply", `{"operation":"multiply","a":"-2","b":"3.5"}`, 200, "", "-7"},
		{"divide", `{"operation":"divide","a":"7","b":"2"}`, 200, "", "3.5"},
		{"recurring division", `{"operation":"divide","a":"1","b":"3"}`, 200, "", "0." + strings.Repeat("3", 16)},
		{"power", `{"operation":"power","a":"2","b":"10"}`, 200, "", "1024"},
		{"small power", `{"operation":"power","a":"0.1","b":"35"}`, 200, "", "1e-35"},
		{"16 significant digits", `{"operation":"divide","a":"2","b":"3"}`, 200, "", "0.6666666666666667"},
		{"tiny division", `{"operation":"divide","a":"1e-100","b":"3"}`, 200, "", "3.333333333333333e-101"},
		{"fractional power", `{"operation":"power","a":"2","b":"0.5"}`, 200, "", "1.414213562373095"},
		{"sqrt", `{"operation":"sqrt","a":"9"}`, 200, "", "3"},
		{"sqrt null b", `{"operation":"sqrt","a":"9","b":null}`, 200, "", "3"},
		{"percentage", `{"operation":"percentage","a":"200","b":"25"}`, 200, "", "50"},
		{"percentage zero rate", `{"operation":"percentage","a":"200","b":"0"}`, 200, "", "0"},
		{"percentage missing rate", `{"operation":"percentage","a":"200"}`, 400, "invalid_calculation", ""},
		{"percentage null rate", `{"operation":"percentage","a":"200","b":null}`, 400, "invalid_calculation", ""},
		{"explicit zero", `{"operation":"add","a":"0","b":"0"}`, 200, "", "0"},
		{"scientific notation", `{"operation":"add","a":"1e2","b":"2"}`, 200, "", "102"},
		{"divide zero", `{"operation":"divide","a":"2","b":"0"}`, 400, "invalid_calculation", ""},
		{"negative root", `{"operation":"sqrt","a":"-9"}`, 400, "invalid_calculation", ""},
		{"imaginary power", `{"operation":"power","a":"-1","b":"0.5"}`, 400, "invalid_calculation", ""},
		{"power limit", `{"operation":"power","a":"10","b":"1001"}`, 400, "invalid_calculation", ""},
		{"large scientific result", `{"operation":"power","a":"1e5","b":"1000"}`, 200, "", "1e+5000"},
		{"unknown operation", `{"operation":"unknown","a":"1"}`, 400, "invalid_calculation", ""},
		{"missing a", `{"operation":"add","b":"2"}`, 400, "invalid_calculation", ""},
		{"missing b", `{"operation":"add","a":"2"}`, 400, "invalid_calculation", ""},
		{"null a", `{"operation":"add","a":null,"b":"2"}`, 400, "invalid_calculation", ""},
		{"unary extra operand", `{"operation":"sqrt","a":"9","b":"2"}`, 400, "invalid_calculation", ""},
		{"null body", `null`, 400, "invalid_calculation", ""},
		{"numeric operand", `{"operation":"add","a":2,"b":"3"}`, 400, "invalid_json", ""},
		{"numeric second operand", `{"operation":"add","a":"2","b":3}`, 400, "invalid_json", ""},
		{"boolean operand", `{"operation":"sqrt","a":true}`, 400, "invalid_json", ""},
		{"object operand", `{"operation":"sqrt","a":{}}`, 400, "invalid_json", ""},
		{"empty operand", `{"operation":"sqrt","a":""}`, 400, "invalid_calculation", ""},
		{"invalid decimal", `{"operation":"sqrt","a":"hello"}`, 400, "invalid_calculation", ""},
		{"nonfinite operand", `{"operation":"sqrt","a":"NaN"}`, 400, "invalid_calculation", ""},
		{"operand exponent limit", `{"operation":"sqrt","a":"1e1001"}`, 400, "invalid_calculation", ""},
		{"operand length limit", `{"operation":"sqrt","a":"` + strings.Repeat("9", 129) + `"}`, 400, "invalid_calculation", ""},
		{"invalid JSON", `{`, 400, "invalid_json", ""},
		{"empty body", ``, 400, "invalid_json", ""},
		{"array body", `[]`, 400, "invalid_json", ""},
		{"unknown fields", `{"operation":"sqrt","a":"9","typo":2}`, 400, "invalid_json", ""},
		{"multiple objects", `{"operation":"sqrt","a":"9"} {}`, 400, "invalid_json", ""},
		{"trailing garbage", `{"operation":"sqrt","a":"9"} xyz`, 400, "invalid_json", ""},
		{"oversized body", `{"operation":"` + strings.Repeat("x", 4096) + `"}`, 413, "body_too_large", ""},
		{"oversized trailing whitespace", `{"operation":"sqrt","a":"9"}` + strings.Repeat(" ", 4096), 413, "body_too_large", ""},
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
				Result string   `json:"result"`
				Error  apiError `json:"error"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tt.status == 200 && body.Result != tt.result {
				t.Errorf("result = %q, want %q", body.Result, tt.result)
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
