package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestServiceCreationWithValidation tests comprehensive service creation with validation
func TestServiceCreationWithValidation(t *testing.T) {
	tests := []struct {
		name         string
		service      *ApplicationService
		expectedErr  string
		shouldFail   bool
		responseCode int
		responseBody string
	}{
		{
			name: "valid MySQL service",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "mysql",
				Version:       "8.0",
				MemoryRequest: "1Gi",
				CPURequest:    "500m",
				StorageSize:   "10Gi",
			},
			shouldFail:   false,
			responseCode: 201,
			responseBody: `{
				"success": true,
				"data": {
					"id": 1,
					"application_id": 1,
					"type": "mysql",
					"version": "8.0",
					"status": "creating"
				}
			}`,
		},
		{
			name: "invalid service type",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "invalid-type",
			},
			expectedErr: "invalid service type 'invalid-type'",
			shouldFail:  true,
		},
		{
			name: "worker service with command",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "worker",
				Command:       "php artisan queue:work",
				Replicas:      2,
			},
			shouldFail:   false,
			responseCode: 201,
			responseBody: `{
				"success": true,
				"data": {
					"id": 2,
					"application_id": 1,
					"type": "worker",
					"command": "php artisan queue:work",
					"replicas": 2,
					"status": "creating"
				}
			}`,
		},
		{
			name: "worker service without command",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "worker",
			},
			expectedErr: "command is required for worker type services",
			shouldFail:  true,
		},
		{
			name: "invalid memory format",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "redis",
				MemoryRequest: "invalid-memory",
			},
			expectedErr: "invalid memory_request format",
			shouldFail:  true,
		},
		{
			name: "API validation error response",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "postgresql",
				Version:       "invalid-version",
			},
			shouldFail:   true,
			responseCode: 422,
			responseBody: `{
				"message": "Validation failed",
				"errors": {
					"version": ["Version 'invalid-version' is not supported for PostgreSQL"],
					"storage_size": ["Storage size is required for database services"]
				}
			}`,
			expectedErr: "failed to create service: Validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			
			if !tt.shouldFail || tt.responseCode > 0 {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if tt.responseCode > 0 {
						w.WriteHeader(tt.responseCode)
						w.Write([]byte(tt.responseBody))
					} else {
						w.WriteHeader(http.StatusCreated)
						w.Write([]byte(`{"success": true, "data": {}}`))
					}
				}))
				defer server.Close()
			}

			var client *Client
			if server != nil {
				client = NewClient("test-token", &server.URL)
			} else {
				client = NewClient("test-token", nil)
			}

			result, err := client.CreateService(tt.service)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
			}
		})
	}
}

// TestEnhancedLoggingIntegration tests the complete logging functionality
func TestEnhancedLoggingIntegration(t *testing.T) {
	tests := []struct {
		name           string
		requestCount   int
		statusCodes    []int
		responseBodies []string
		expectRetries  bool
	}{
		{
			name:         "successful request",
			requestCount: 1,
			statusCodes:  []int{200},
			responseBodies: []string{`{
				"success": true,
				"data": {
					"id": 1,
					"name": "test-app"
				}
			}`},
			expectRetries: false,
		},
		{
			name:         "request with retry",
			requestCount: 3,
			statusCodes:  []int{500, 503, 200},
			responseBodies: []string{
				`{"message": "Internal server error"}`,
				`{"message": "Service unavailable"}`,
				`{"success": true, "data": {"id": 1}}`,
			},
			expectRetries: true,
		},
		{
			name:         "validation error (no retry)",
			requestCount: 1,
			statusCodes:  []int{422},
			responseBodies: []string{`{
				"message": "Validation failed",
				"errors": {
					"name": ["Name is required"]
				}
			}`},
			expectRetries: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestIndex := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				
				if requestIndex < len(tt.statusCodes) {
					w.WriteHeader(tt.statusCodes[requestIndex])
					if requestIndex < len(tt.responseBodies) {
						w.Write([]byte(tt.responseBodies[requestIndex]))
					}
				}
				requestIndex++
			}))
			defer server.Close()

			client := NewClient("test-token", &server.URL)
			
			app := &Application{
				Name: "test-app",
				Type: "laravel",
			}

			_, err := client.CreateApplication(app)

			if tt.statusCodes[len(tt.statusCodes)-1] >= 400 {
				if err == nil {
					t.Error("Expected error but got none")
				}
			}

			// Verify retry behavior
			if tt.expectRetries && requestIndex <= 1 {
				t.Error("Expected retries but none occurred")
			} else if !tt.expectRetries && requestIndex > 1 {
				t.Errorf("Expected no retries but %d requests were made", requestIndex)
			}
		})
	}
}

// TestCompleteErrorHandlingWorkflow tests the complete error handling pipeline
func TestCompleteErrorHandlingWorkflow(t *testing.T) {
	tests := []struct {
		name                string
		operation          string
		statusCode         int
		responseBody       string
		expectedSuggestion string
		expectedDocsLink   bool
	}{
		{
			name:       "422 service validation error",
			operation:  "create service",
			statusCode: 422,
			responseBody: `{
				"message": "The given data was invalid.",
				"errors": {
					"type": ["The selected type is invalid."],
					"storage_size": ["The storage size field is required."]
				}
			}`,
			expectedSuggestion: "Service type must be one of:",
			expectedDocsLink:   true,
		},
		{
			name:       "404 resource not found",
			operation:  "update application",
			statusCode: 404,
			responseBody: `{
				"message": "Application not found"
			}`,
			expectedSuggestion: "Check that the resource exists and the ID is correct",
			expectedDocsLink:   true,
		},
		{
			name:       "401 authentication error",
			operation:  "delete application",
			statusCode: 401,
			responseBody: `{
				"message": "Unauthenticated."
			}`,
			expectedSuggestion: "Check that your API token is valid and has the required permissions",
			expectedDocsLink:   true,
		},
		{
			name:       "503 service unavailable",
			operation:  "create application",
			statusCode: 503,
			responseBody: `{
				"message": "Service Unavailable"
			}`,
			expectedSuggestion: "This appears to be a server error. Please try again in a few moments",
			expectedDocsLink:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient("test-token", &server.URL)
			
			var err error
			switch tt.operation {
			case "create service":
				_, err = client.CreateService(&ApplicationService{
					ApplicationID: 1,
					Type:          "mysql",
				})
			case "update application":
				_, err = client.UpdateApplication(999, map[string]interface{}{"name": "updated"})
			case "delete application":
				err = client.DeleteApplication(999)
			case "create application":
				_, err = client.CreateApplication(&Application{Name: "test", Type: "laravel"})
			}

			if err == nil {
				t.Fatal("Expected error but got none")
			}

			errorMsg := err.Error()
			
			// Check that suggestion is included
			if !strings.Contains(errorMsg, tt.expectedSuggestion) {
				t.Errorf("Expected error to contain suggestion '%s', got '%s'", tt.expectedSuggestion, errorMsg)
			}

			// Check that documentation link is included
			if tt.expectedDocsLink && !strings.Contains(errorMsg, "https://docs.ploi.io/cloud") {
				t.Errorf("Expected error to contain documentation link, got '%s'", errorMsg)
			}
		})
	}
}

// TestVolumeReadOnlyFunctionality tests volume read-only mode
func TestVolumeReadOnlyFunctionality(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "volume GET allowed",
			method:       "GET",
			path:         "/applications/1/volumes/1",
			expectedCode: 200,
		},
		{
			name:         "volume PATCH allowed",
			method:       "PATCH", 
			path:         "/applications/1/volumes/1",
			expectedCode: 200,
		},
		{
			name:         "volume POST not allowed",
			method:       "POST",
			path:         "/applications/1/volumes",
			expectedCode: 405,
			expectedMsg:  "Volume creation is not supported",
		},
		{
			name:         "volume PUT allowed for resize",
			method:       "PUT",
			path:         "/applications/1/volumes/1",
			expectedCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				
				if tt.method == "POST" && strings.Contains(r.URL.Path, "/volumes") && !strings.Contains(r.URL.Path, "/volumes/") {
					// POST to create volume - not allowed
					w.WriteHeader(405)
					w.Write([]byte(`{"message": "Volume creation is not supported. Volumes are automatically created with persistent storage services."}`))
				} else {
					// All other operations allowed
					w.WriteHeader(200)
					w.Write([]byte(`{"success": true, "data": {}}`))
				}
			}))
			defer server.Close()

			client := NewClient("test-token", &server.URL)
			
			var err error
			switch tt.method {
			case "GET":
				_, err = client.GetVolume(1, 1)
			case "POST":
				_, err = client.CreateVolume(&ApplicationVolume{
					ApplicationID: 1,
					Name:          "test-volume",
					Size:          10,
					MountPath:     "/data",
				})
			case "PUT":
				_, err = client.UpdateVolume(1, 1, &ApplicationVolume{Size: 20})
			case "DELETE":
				err = client.DeleteVolume(1, 1)
			}

			if tt.expectedCode >= 400 {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.expectedMsg != "" && !strings.Contains(err.Error(), tt.expectedMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
			}
		})
	}
}

// TestResourceValidationEdgeCases tests edge cases in resource validation
func TestResourceValidationEdgeCases(t *testing.T) {
	client := NewClient("test-token", nil)

	tests := []struct {
		name        string
		service     *ApplicationService
		expectedErr string
	}{
		{
			name:        "nil service",
			service:     nil,
			expectedErr: "service cannot be nil",
		},
		{
			name: "zero application ID",
			service: &ApplicationService{
				ApplicationID: 0,
				Type:          "redis",
			},
			expectedErr: "application_id must be greater than 0",
		},
		{
			name: "negative application ID",
			service: &ApplicationService{
				ApplicationID: -1,
				Type:          "redis",
			},
			expectedErr: "application_id must be greater than 0",
		},
		{
			name: "empty service type",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "",
			},
			expectedErr: "service type is required",
		},
		{
			name: "mixed case service type",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "MySQL", // Should fail - types are case sensitive
			},
			expectedErr: "invalid service type 'MySQL'",
		},
		{
			name: "worker with empty command",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "worker",
				Command:       "", // Empty command should fail
			},
			expectedErr: "command is required for worker type services",
		},
		{
			name: "invalid memory with valid CPU",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "redis",
				MemoryRequest: "invalid-mem",
				CPURequest:    "250m",
			},
			expectedErr: "invalid memory_request format 'invalid-mem'",
		},
		{
			name: "valid memory with invalid CPU",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "redis",
				MemoryRequest: "512Mi",
				CPURequest:    "invalid-cpu",
			},
			expectedErr: "invalid cpu_request format 'invalid-cpu'",
		},
		{
			name: "boundary case - valid minimum resources",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "redis",
				MemoryRequest: "1Mi",
				CPURequest:    "1m",
				StorageSize:   "1Gi",
			},
			expectedErr: "", // Should pass
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateServiceRequest(tt.service)
			
			if tt.expectedErr == "" {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error containing '%s' but got none", tt.expectedErr)
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
				}
			}
		})
	}
}

// TestLogEntryStructure tests that log entries are properly structured
func TestLogEntryStructure(t *testing.T) {
	logEntry := LogEntry{
		Timestamp:    time.Now(),
		Method:       "POST",
		URL:          "https://api.ploi.io/applications",
		RequestBody:  `{"name": "test-app"}`,
		StatusCode:   201,
		ResponseBody: `{"success": true, "data": {"id": 1}}`,
		Error:        "",
		Duration:     time.Millisecond * 150,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		t.Fatalf("Failed to marshal log entry: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled LogEntry
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	// Verify fields are preserved
	if unmarshaled.Method != logEntry.Method {
		t.Errorf("Expected Method %s, got %s", logEntry.Method, unmarshaled.Method)
	}
	if unmarshaled.StatusCode != logEntry.StatusCode {
		t.Errorf("Expected StatusCode %d, got %d", logEntry.StatusCode, unmarshaled.StatusCode)
	}
	if unmarshaled.Duration != logEntry.Duration {
		t.Errorf("Expected Duration %v, got %v", logEntry.Duration, unmarshaled.Duration)
	}
}

// TestDetailedErrorStructure tests detailed error response structure
func TestDetailedErrorStructure(t *testing.T) {
	detailedErr := DetailedError{
		StatusCode: 422,
		Message:    "Validation failed",
		Errors: map[string][]string{
			"type": {"Invalid service type"},
			"storage_size": {"Storage size is required"},
		},
		Suggestion: "Check service type and storage size configuration",
		DocsLink:   "https://docs.ploi.io/cloud",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(detailedErr)
	if err != nil {
		t.Fatalf("Failed to marshal detailed error: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled DetailedError
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal detailed error: %v", err)
	}

	// Verify fields are preserved
	if unmarshaled.StatusCode != detailedErr.StatusCode {
		t.Errorf("Expected StatusCode %d, got %d", detailedErr.StatusCode, unmarshaled.StatusCode)
	}
	if unmarshaled.Message != detailedErr.Message {
		t.Errorf("Expected Message %s, got %s", detailedErr.Message, unmarshaled.Message)
	}
	if len(unmarshaled.Errors) != len(detailedErr.Errors) {
		t.Errorf("Expected %d error fields, got %d", len(detailedErr.Errors), len(unmarshaled.Errors))
	}
}

// TestClientNilSafety tests that client handles nil values safely
func TestClientNilSafety(t *testing.T) {
	tests := []struct {
		name string
		test func() error
	}{
		{
			name: "nil client doRequest",
			test: func() error {
				var client *Client
				_, err := client.doRequest("GET", "/test", nil)
				return err
			},
		},
		{
			name: "client with nil httpClient",
			test: func() error {
				client := &Client{
					apiToken:    "test",
					apiEndpoint: "http://test.com",
					httpClient:  nil,
					logger:      &Logger{},
				}
				_, err := client.doRequest("GET", "/test", nil)
				return err
			},
		},
		{
			name: "client with empty endpoint",
			test: func() error {
				client := &Client{
					apiToken:    "test",
					apiEndpoint: "",
					httpClient:  &http.Client{},
					logger:      &Logger{},
				}
				_, err := client.doRequest("GET", "/test", nil)
				return err
			},
		},
		{
			name: "client with empty token",
			test: func() error {
				client := &Client{
					apiToken:    "",
					apiEndpoint: "http://test.com",
					httpClient:  &http.Client{},
					logger:      &Logger{},
				}
				_, err := client.doRequest("GET", "/test", nil)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.test()
			if err == nil {
				t.Error("Expected error for nil safety test but got none")
			}
		})
	}
}