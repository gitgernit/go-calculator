package calc

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCalculateHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Valid Expression",
			method:         http.MethodPost,
			body:           `{"expression": "3+5"}`,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"result": 8.0},
		},
		{
			name:           "Invalid Expression",
			method:         http.MethodPost,
			body:           `{"expression": "3+5+"}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   map[string]interface{}{"error": "Expression is not valid"},
		},
		{
			name:           "Method Not Allowed",
			method:         http.MethodGet,
			body:           ``,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/calculate", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			CalculateHandler(rec, req)

			res := rec.Result()
			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if tt.expectedBody != nil {
				var responseBody map[string]interface{}
				err := json.NewDecoder(res.Body).Decode(&responseBody)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				for key, expectedValue := range tt.expectedBody {
					if responseBody[key] != expectedValue {
						t.Errorf("expected %s to be %v, got %v", key, expectedValue, responseBody[key])
					}
				}
			}
		})
	}
}
