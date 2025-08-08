package client

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewClient_LoggingConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		tfLog    string
		ploiDebug string
		expected bool
	}{
		{"no debug env vars", "", "", false},
		{"TF_LOG=DEBUG", "DEBUG", "", true},
		{"PLOI_DEBUG=1", "", "1", true},
		{"both enabled", "DEBUG", "1", true},
		{"TF_LOG=INFO", "INFO", "", false},
		{"PLOI_DEBUG=0", "", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			oldTfLog := os.Getenv("TF_LOG")
			oldPloiDebug := os.Getenv("PLOI_DEBUG")
			defer func() {
				os.Setenv("TF_LOG", oldTfLog)
				os.Setenv("PLOI_DEBUG", oldPloiDebug)
			}()

			if tt.tfLog != "" {
				os.Setenv("TF_LOG", tt.tfLog)
			} else {
				os.Unsetenv("TF_LOG")
			}

			if tt.ploiDebug != "" {
				os.Setenv("PLOI_DEBUG", tt.ploiDebug)
			} else {
				os.Unsetenv("PLOI_DEBUG")
			}

			client := NewClient("test-token", nil)

			if client.logger.enabled != tt.expected {
				t.Errorf("Expected logger.enabled %v, got %v", tt.expected, client.logger.enabled)
			}
			if client.logger.debug != tt.expected {
				t.Errorf("Expected logger.debug %v, got %v", tt.expected, client.logger.debug)
			}
		})
	}
}

func TestValidateServiceRequest(t *testing.T) {
	client := NewClient("test-token", nil)

	tests := []struct {
		name        string
		service     *ApplicationService
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil service",
			service:     nil,
			expectError: true,
			errorMsg:    "service cannot be nil",
		},
		{
			name: "invalid application id",
			service: &ApplicationService{
				ApplicationID: 0,
				Type:          "mysql",
			},
			expectError: true,
			errorMsg:    "application_id must be greater than 0",
		},
		{
			name: "missing type",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "",
			},
			expectError: true,
			errorMsg:    "service type is required",
		},
		{
			name: "invalid service type",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "invalid",
			},
			expectError: true,
			errorMsg:    "invalid service type 'invalid'. Must be one of: mysql, postgresql, redis, valkey, rabbitmq, mongodb, minio, sftp, worker",
		},
		{
			name: "worker without command",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "worker",
			},
			expectError: true,
			errorMsg:    "command is required for worker type services",
		},
		{
			name: "invalid memory format",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "mysql",
				MemoryRequest: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid memory_request format 'invalid'. Use format like '256Mi' or '1Gi'",
		},
		{
			name: "invalid cpu format",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "mysql",
				CPURequest:    "invalid",
			},
			expectError: true,
			errorMsg:    "invalid cpu_request format 'invalid'. Use format like '250m', '1', or '2'",
		},
		{
			name: "invalid storage format",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "mysql",
				StorageSize:   "invalid",
			},
			expectError: true,
			errorMsg:    "invalid storage_size format 'invalid'. Use format like '1Gi' or '10Gi'",
		},
		{
			name: "valid mysql service",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "mysql",
				MemoryRequest: "1Gi",
				CPURequest:    "500m",
				StorageSize:   "10Gi",
			},
			expectError: false,
		},
		{
			name: "valid worker service",
			service: &ApplicationService{
				ApplicationID: 1,
				Type:          "worker",
				Command:       "php artisan queue:work",
				MemoryRequest: "512Mi",
				CPURequest:    "250m",
				Replicas:      2,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.ValidateServiceRequest(tt.service)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestIsValidResourceSpec(t *testing.T) {
	tests := []struct {
		name       string
		spec       string
		validUnits []string
		expected   bool
	}{
		{"valid memory Mi", "256Mi", []string{"Mi", "Gi"}, true},
		{"valid memory Gi", "1Gi", []string{"Mi", "Gi"}, true},
		{"valid storage with decimal", "1.5Gi", []string{"Mi", "Gi", "Ti"}, true},
		{"invalid unit", "256MB", []string{"Mi", "Gi"}, false},
		{"no unit", "256", []string{"Mi", "Gi"}, false},
		{"empty spec", "", []string{"Mi", "Gi"}, false},
		{"invalid number", "aMi", []string{"Mi", "Gi"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidResourceSpec(tt.spec, tt.validUnits)
			if result != tt.expected {
				t.Errorf("Expected %v for spec '%s', got %v", tt.expected, tt.spec, result)
			}
		})
	}
}

func TestIsValidCPUSpec(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		expected bool
	}{
		{"valid millicores", "250m", true},
		{"valid cores integer", "1", true},
		{"valid cores decimal", "1.5", true},
		{"invalid millicores", "am", false},
		{"invalid format", "250x", false},
		{"empty spec", "", false},
		{"zero millicores", "0m", true},
		{"zero cores", "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidCPUSpec(tt.spec)
			if result != tt.expected {
				t.Errorf("Expected %v for spec '%s', got %v", tt.expected, tt.spec, result)
			}
		})
	}
}

func TestSanitizeToken(t *testing.T) {
	client := NewClient("test-token", nil)

	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{"short token", "abc", "***"},
		{"normal token", "abcd1234efgh5678", "abcd***5678"},
		{"very short", "ab", "***"},
		{"exactly 8 chars", "12345678", "***"},
		{"9 chars", "123456789", "1234***6789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.sanitizeToken(tt.token)
			if result != tt.expected {
				t.Errorf("Expected '%s' for token '%s', got '%s'", tt.expected, tt.token, result)
			}
		})
	}
}

func TestSanitizeURL(t *testing.T) {
	client := NewClient("test-token", nil)

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"no query params", "https://api.ploi.io/v1/applications", "https://api.ploi.io/v1/applications"},
		{"with query params", "https://api.ploi.io/v1/applications?token=secret", "https://api.ploi.io/v1/applications?[params sanitized]"},
		{"multiple query params", "https://api.ploi.io/v1/applications?page=1&size=10", "https://api.ploi.io/v1/applications?[params sanitized]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.sanitizeURL(tt.url)
			if result != tt.expected {
				t.Errorf("Expected '%s' for URL '%s', got '%s'", tt.expected, tt.url, result)
			}
		})
	}
}

func TestGenerateValidationSuggestion(t *testing.T) {
	client := NewClient("test-token", nil)

	tests := []struct {
		name      string
		operation string
		errors    map[string][]string
		expected  string
	}{
		{
			name:      "no errors",
			operation: "create service",
			errors:    map[string][]string{},
			expected:  "Check the API documentation for required fields and valid values",
		},
		{
			name:      "type error",
			operation: "create service",
			errors: map[string][]string{
				"type": {"Invalid service type"},
			},
			expected: "Service type must be one of: mysql, postgresql, redis, valkey, rabbitmq, mongodb, minio, sftp",
		},
		{
			name:      "storage_size error",
			operation: "create service",
			errors: map[string][]string{
				"storage_size": {"Invalid format"},
			},
			expected: "Storage size must be specified with units (e.g., '1Gi', '10Gi')",
		},
		{
			name:      "multiple errors",
			operation: "create service",
			errors: map[string][]string{
				"type":         {"Invalid type"},
				"memory_request": {"Invalid format"},
			},
			expected: "Service type must be one of: mysql, postgresql, redis, valkey, rabbitmq, mongodb, minio, sftp",
		},
		{
			name:      "cpu_request error",
			operation: "create service",
			errors: map[string][]string{
				"cpu_request": {"Invalid CPU format"},
			},
			expected: "CPU request must be specified correctly (e.g., '250m', '1', '2')",
		},
		{
			name:      "memory_request error",
			operation: "create service",
			errors: map[string][]string{
				"memory_request": {"Invalid memory format"},
			},
			expected: "Memory request must be specified with units (e.g., '256Mi', '1Gi')",
		},
		{
			name:      "version error",
			operation: "create service",
			errors: map[string][]string{
				"version": {"Unsupported version"},
			},
			expected: "Check that the version is supported for the selected service type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.generateValidationSuggestion(tt.operation, tt.errors)
			if !strings.Contains(result, tt.expected) && result != tt.expected {
				t.Errorf("Expected suggestion to contain '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestHandleErrorResponse(t *testing.T) {
	client := NewClient("test-token", nil)

	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		operation      string
		expectedError  string
		expectedSuggestion string
	}{
		{
			name:       "422 validation error with structured errors",
			statusCode: 422,
			responseBody: `{
				"message": "Validation failed",
				"errors": {
					"type": ["Invalid service type"],
					"storage_size": ["Must include units"]
				}
			}`,
			operation:     "create service",
			expectedError: "failed to create service: Validation failed",
			expectedSuggestion: "Service type must be one of:",
		},
		{
			name:       "404 not found error",
			statusCode: 404,
			responseBody: `{
				"message": "Resource not found"
			}`,
			operation:     "update service",
			expectedError: "failed to update service: Resource not found",
			expectedSuggestion: "Check that the resource exists and the ID is correct",
		},
		{
			name:       "401 unauthorized error",
			statusCode: 401,
			responseBody: `{
				"message": "Unauthorized"
			}`,
			operation:     "create service",
			expectedError: "failed to create service: Unauthorized",
			expectedSuggestion: "Check that your API token is valid and has the required permissions",
		},
		{
			name:       "403 forbidden error",
			statusCode: 403,
			responseBody: `{
				"message": "Forbidden"
			}`,
			operation:     "delete service",
			expectedError: "failed to delete service: Forbidden",
			expectedSuggestion: "Check that your API token has permission to perform this operation",
		},
		{
			name:       "500 server error",
			statusCode: 500,
			responseBody: `{
				"message": "Internal server error"
			}`,
			operation:     "create service",
			expectedError: "failed to create service: Internal server error",
			expectedSuggestion: "This appears to be a server error. Please try again in a few moments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP response
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")

			err := client.handleErrorResponse(resp, tt.operation)
			
			if err == nil {
				t.Fatalf("Expected error but got none")
			}

			errorMsg := err.Error()
			if !strings.Contains(errorMsg, tt.expectedError) {
				t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedError, errorMsg)
			}

			if !strings.Contains(errorMsg, tt.expectedSuggestion) {
				t.Errorf("Expected error to contain suggestion '%s', got '%s'", tt.expectedSuggestion, errorMsg)
			}

			if !strings.Contains(errorMsg, "https://docs.ploi.io/cloud") {
				t.Errorf("Expected error to contain documentation link, got '%s'", errorMsg)
			}
		})
	}
}

func TestDoRequestWithRetry(t *testing.T) {
	tests := []struct {
		name           string
		statusCodes    []int
		expectRetries  int
		expectSuccess  bool
	}{
		{
			name:          "immediate success",
			statusCodes:   []int{200},
			expectRetries: 0,
			expectSuccess: true,
		},
		{
			name:          "retry on 500 then success",
			statusCodes:   []int{500, 200},
			expectRetries: 1,
			expectSuccess: true,
		},
		{
			name:          "retry on 503 then success",
			statusCodes:   []int{503, 504, 200},
			expectRetries: 2,
			expectSuccess: true,
		},
		{
			name:          "max retries exceeded",
			statusCodes:   []int{500, 500, 500, 500},
			expectRetries: 3,
			expectSuccess: true, // Returns last response even after retries
		},
		{
			name:          "no retry on 422",
			statusCodes:   []int{422},
			expectRetries: 0,
			expectSuccess: true,
		},
		{
			name:          "no retry on 404",
			statusCodes:   []int{404},
			expectRetries: 0,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if requestCount < len(tt.statusCodes) {
					statusCode := tt.statusCodes[requestCount]
					w.WriteHeader(statusCode)
					
					if statusCode >= 400 {
						w.Header().Set("Content-Type", "application/json")
						w.Write([]byte(`{"message": "Test error"}`))
					} else {
						w.Write([]byte(`{"success": true}`))
					}
				} else {
					w.WriteHeader(200)
					w.Write([]byte(`{"success": true}`))
				}
				requestCount++
			}))
			defer server.Close()

			client := NewClient("test-token", &server.URL)
			
			resp, err := client.doRequestWithRetry("GET", "/test", nil, 3)
			
			if tt.expectSuccess && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			}

			if tt.expectSuccess && resp == nil {
				t.Error("Expected response but got nil")
			}

			actualRetries := requestCount - 1
			if actualRetries != tt.expectRetries {
				t.Errorf("Expected %d retries, got %d", tt.expectRetries, actualRetries)
			}
		})
	}
}

func TestLogRequest(t *testing.T) {
	tests := []struct {
		name          string
		tfLog         string
		ploiDebug     string
		expectLogging bool
		expectDebug   bool
	}{
		{
			name:          "debug logging enabled via TF_LOG",
			tfLog:         "DEBUG",
			ploiDebug:     "",
			expectLogging: true,
			expectDebug:   true,
		},
		{
			name:          "debug logging enabled via PLOI_DEBUG",
			tfLog:         "",
			ploiDebug:     "1",
			expectLogging: true,
			expectDebug:   true,
		},
		{
			name:          "logging disabled",
			tfLog:         "",
			ploiDebug:     "",
			expectLogging: false,
			expectDebug:   false,
		},
		{
			name:          "TF_LOG INFO level",
			tfLog:         "INFO",
			ploiDebug:     "",
			expectLogging: false,
			expectDebug:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			oldTfLog := os.Getenv("TF_LOG")
			oldPloiDebug := os.Getenv("PLOI_DEBUG")
			defer func() {
				os.Setenv("TF_LOG", oldTfLog)
				os.Setenv("PLOI_DEBUG", oldPloiDebug)
			}()

			if tt.tfLog != "" {
				os.Setenv("TF_LOG", tt.tfLog)
			} else {
				os.Unsetenv("TF_LOG")
			}

			if tt.ploiDebug != "" {
				os.Setenv("PLOI_DEBUG", tt.ploiDebug)
			} else {
				os.Unsetenv("PLOI_DEBUG")
			}

			client := NewClient("test-token", nil)

			if client.logger.enabled != tt.expectLogging {
				t.Errorf("Expected logger.enabled %v, got %v", tt.expectLogging, client.logger.enabled)
			}

			if client.logger.debug != tt.expectDebug {
				t.Errorf("Expected logger.debug %v, got %v", tt.expectDebug, client.logger.debug)
			}

			// Test that logRequest doesn't panic and handles different scenarios
			client.logRequest("GET", "https://api.example.com/test", `{"test": "data"}`, 200, `{"success": true}`, "", time.Millisecond*100)
			client.logRequest("POST", "https://api.example.com/error", `{"test": "data"}`, 500, `{"error": "server error"}`, "HTTP 500: Internal Server Error", time.Millisecond*200)
		})
	}
}

func TestSanitizeBody(t *testing.T) {
	client := NewClient("test-token", nil)

	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "regular JSON body",
			body:     `{"name": "test", "type": "mysql"}`,
			expected: `{"name": "test", "type": "mysql"}`,
		},
		{
			name:     "empty body",
			body:     "",
			expected: "",
		},
		{
			name:     "body with special characters",
			body:     `{"command": "php artisan queue:work --timeout=60"}`,
			expected: `{"command": "php artisan queue:work --timeout=60"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.sanitizeBody(tt.body)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestWorkerResourceDeprecation(t *testing.T) {
	// Test that worker validation suggests using services instead
	worker := &Worker{
		ApplicationID: 1,
		Name:          "test-worker",
		Command:       "php artisan queue:work",
		Replicas:      1,
	}

	// Create a mock server that returns 404 for worker endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/workers") {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "Worker endpoints are deprecated. Use services with type 'worker' instead."}`))
		}
	}))
	defer server.Close()

	testClient := NewClient("test-token", &server.URL)
	
	_, err := testClient.CreateWorker(worker)
	if err == nil {
		t.Error("Expected error for deprecated worker endpoint")
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "Worker endpoints are deprecated") {
		t.Errorf("Expected deprecation message, got: %s", errorMsg)
	}
}

func TestVolumeCreateRestriction(t *testing.T) {
	// Test that volume creation returns appropriate error
	volume := &ApplicationVolume{
		ApplicationID: 1,
		Name:          "test-volume",
		Size:          10,
		MountPath:     "/var/lib/data",
	}

	// Create a mock server that returns 405 Method Not Allowed for volume POST
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/volumes") && r.Method == "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message": "Volume creation is not supported. Volumes are automatically created with persistent storage services."}`))
		}
	}))
	defer server.Close()

	testClient := NewClient("test-token", &server.URL)
	
	_, err := testClient.CreateVolume(volume)
	if err == nil {
		t.Error("Expected error for volume creation")
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "not supported") {
		t.Errorf("Expected volume creation restriction message, got: %s", errorMsg)
	}
}

func TestErrorResponseParsing(t *testing.T) {
	client := NewClient("test-token", nil)

	tests := []struct {
		name                string
		statusCode          int
		responseBody        string
		expectedErrorFormat string
		expectParsing       bool
	}{
		{
			name:       "valid JSON error response",
			statusCode: 422,
			responseBody: `{
				"message": "Validation failed",
				"errors": {
					"type": ["Invalid service type"],
					"memory_request": "Memory format is invalid"
				}
			}`,
			expectedErrorFormat: "failed to create service: Validation failed",
			expectParsing:       true,
		},
		{
			name:                "invalid JSON response",
			statusCode:          500,
			responseBody:        `invalid json {`,
			expectedErrorFormat: "failed to create service: 500 Internal Server Error",
			expectParsing:       false,
		},
		{
			name:       "empty errors map",
			statusCode: 422,
			responseBody: `{
				"message": "Validation failed",
				"errors": {}
			}`,
			expectedErrorFormat: "failed to create service: Validation failed",
			expectParsing:       true,
		},
		{
			name:       "mixed error value types",
			statusCode: 422,
			responseBody: `{
				"message": "Validation failed",
				"errors": {
					"type": ["Invalid service type"],
					"memory_request": "Memory format is invalid",
					"cpu_request": ["CPU format", "must be valid"],
					"replicas": 123
				}
			}`,
			expectedErrorFormat: "failed to create service: Validation failed",
			expectParsing:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Status:     fmt.Sprintf("%d %s", tt.statusCode, http.StatusText(tt.statusCode)),
				Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
				Header:     make(http.Header),
			}
			resp.Header.Set("Content-Type", "application/json")

			err := client.handleErrorResponse(resp, "create service")
			
			if err == nil {
				t.Fatalf("Expected error but got none")
			}

			errorMsg := err.Error()
			if !strings.Contains(errorMsg, tt.expectedErrorFormat) {
				t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedErrorFormat, errorMsg)
			}

			if tt.expectParsing {
				if !strings.Contains(errorMsg, "Suggestion:") {
					t.Errorf("Expected error to contain suggestion, got '%s'", errorMsg)
				}
				if !strings.Contains(errorMsg, "Documentation:") {
					t.Errorf("Expected error to contain documentation link, got '%s'", errorMsg)
				}
			}
		})
	}
}

func TestRetryLogicEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		statusCodes   []int
		expectRetries int
		expectError   bool
	}{
		{
			name:          "502 bad gateway retry",
			statusCodes:   []int{502, 200},
			expectRetries: 1,
			expectError:   false,
		},
		{
			name:          "503 service unavailable retry",
			statusCodes:   []int{503, 200},
			expectRetries: 1,
			expectError:   false,
		},
		{
			name:          "504 gateway timeout retry",
			statusCodes:   []int{504, 200},
			expectRetries: 1,
			expectError:   false,
		},
		{
			name:          "multiple 5xx errors then success",
			statusCodes:   []int{500, 502, 503, 200},
			expectRetries: 3,
			expectError:   false,
		},
		{
			name:          "max retries with 5xx errors",
			statusCodes:   []int{500, 500, 500, 500, 500},
			expectRetries: 3,
			expectError:   false, // Returns last response even after max retries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if requestCount < len(tt.statusCodes) {
					statusCode := tt.statusCodes[requestCount]
					w.WriteHeader(statusCode)
					if statusCode >= 400 {
						w.Header().Set("Content-Type", "application/json")
						w.Write([]byte(`{"message": "Test error"}`))
					} else {
						w.Write([]byte(`{"success": true}`))
					}
				} else {
					// Fallback for edge cases
					w.WriteHeader(200)
					w.Write([]byte(`{"success": true}`))
				}
				requestCount++
			}))
			defer server.Close()

			client := NewClient("test-token", &server.URL)
			
			_, err := client.doRequestWithRetry("GET", "/test", nil, 3)
			
			actualRetries := requestCount - 1
			if actualRetries != tt.expectRetries {
				t.Errorf("Expected %d retries, got %d", tt.expectRetries, actualRetries)
			}

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestServiceCreateWithValidationFlow(t *testing.T) {
	// Test the complete flow from validation to API error handling
	service := &ApplicationService{
		ApplicationID: 1,
		Type:          "mysql",
		MemoryRequest: "256Mi",
		CPURequest:    "250m",
		StorageSize:   "1Gi",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/services") {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"message": "The given data was invalid",
				"errors": {
					"version": ["The version field is required for mysql services"],
					"storage_size": ["Storage size must be at least 1Gi for mysql"]
				}
			}`))
		}
	}))
	defer server.Close()

	client := NewClient("test-token", &server.URL)
	
	_, err := client.CreateService(service)
	if err == nil {
		t.Fatal("Expected error from service creation")
	}

	errorMsg := err.Error()
	expectedParts := []string{
		"failed to create service:",
		"The given data was invalid",
		"Suggestion:",
		"Documentation:",
		"https://docs.ploi.io/cloud",
	}

	for _, part := range expectedParts {
		if !strings.Contains(errorMsg, part) {
			t.Errorf("Expected error message to contain '%s', got: %s", part, errorMsg)
		}
	}
}

func TestNilClientHandling(t *testing.T) {
	var client *Client
	
	_, err := client.doRequestWithRetry("GET", "/test", nil, 3)
	if err == nil {
		t.Error("Expected error for nil client")
	}
	
	if !strings.Contains(err.Error(), "client is nil") {
		t.Errorf("Expected 'client is nil' error, got: %v", err)
	}
}

func TestClientFieldValidation(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		endpoint    string
		expectedErr string
	}{
		{
			name:        "empty token",
			token:       "",
			endpoint:    "https://api.test.com",
			expectedErr: "api token is empty",
		},
		{
			name:        "empty endpoint",
			token:       "test-token",
			endpoint:    "",
			expectedErr: "api endpoint is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				httpClient:  &http.Client{},
				apiToken:    tt.token,
				apiEndpoint: tt.endpoint,
				logger:      &Logger{enabled: false, debug: false},
			}

			_, err := client.doRequestWithRetry("GET", "/test", nil, 3)
			if err == nil {
				t.Error("Expected error but got none")
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error to contain '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestLoggerOutput(t *testing.T) {
	// Test that logger actually produces output when enabled
	oldTfLog := os.Getenv("TF_LOG")
	defer os.Setenv("TF_LOG", oldTfLog)
	
	os.Setenv("TF_LOG", "DEBUG")
	
	// Capture log output
	var logOutput strings.Builder
	oldOutput := log.Writer()
	log.SetOutput(&logOutput)
	defer log.SetOutput(oldOutput)

	client := NewClient("test-token", nil)
	
	// Generate some log entries
	client.logRequest("GET", "https://api.test.com/test?token=secret", `{"test": "data"}`, 200, `{"success": true}`, "", time.Millisecond*50)
	client.logRequest("POST", "https://api.test.com/error", `{"bad": "data"}`, 422, `{"error": "validation failed"}`, "HTTP 422: Unprocessable Entity", time.Millisecond*100)

	output := logOutput.String()
	
	expectedLogParts := []string{
		"[DEBUG] Ploi API Request: GET",
		"api.test.com/test?[params sanitized]",
		"[DEBUG] Request Body:",
		"[DEBUG] Response Status: 200",
		"[DEBUG] Duration:",
		"[DEBUG] Ploi API Request: POST",
		"[DEBUG] Error: HTTP 422",
	}

	for _, part := range expectedLogParts {
		if !strings.Contains(output, part) {
			t.Errorf("Expected log output to contain '%s', got: %s", part, output)
		}
	}
}

func TestVolumeReadOnlyOperations(t *testing.T) {
	// Test that volume GET and UPDATE operations work (read-only mode)
	volume := &ApplicationVolume{
		ID:            1,
		ApplicationID: 1,
		Name:          "test-volume",
		Size:          20,
		MountPath:     "/var/lib/data",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/volumes/1") {
			if r.Method == "GET" {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				response := `{
					"data": {
						"id": 1,
						"application_id": 1,
						"name": "test-volume",
						"size": 20,
						"path": "/var/lib/data"
					}
				}`
				w.Write([]byte(response))
			} else if r.Method == "PUT" {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				response := `{
					"data": {
						"id": 1,
						"application_id": 1,
						"name": "test-volume",
						"size": 30,
						"path": "/var/lib/data"
					}
				}`
				w.Write([]byte(response))
			}
		}
	}))
	defer server.Close()

	client := NewClient("test-token", &server.URL)
	
	// Test GET operation (should work)
	retrievedVolume, err := client.GetVolume(1, 1)
	if err != nil {
		t.Errorf("Expected no error for volume GET, got: %v", err)
	}
	if retrievedVolume == nil {
		t.Error("Expected volume data but got nil")
	}

	// Test UPDATE operation (should work - volume resize)
	volume.Size = 30
	updatedVolume, err := client.UpdateVolume(1, 1, volume)
	if err != nil {
		t.Errorf("Expected no error for volume UPDATE, got: %v", err)
	}
	if updatedVolume == nil {
		t.Error("Expected updated volume data but got nil")
	}
}

func TestWorkerToServiceMigrationMessage(t *testing.T) {
	// Test that worker operations provide clear migration guidance
	worker := &Worker{
		ApplicationID: 1,
		Name:          "test-worker",
		Command:       "php artisan queue:work",
		Replicas:      2,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/workers") {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"message": "Worker endpoints are deprecated. Use services with type 'worker' instead."
			}`))
		}
	}))
	defer server.Close()

	client := NewClient("test-token", &server.URL)
	
	// Test CreateWorker operation
	t.Run("create_worker", func(t *testing.T) {
		_, err := client.CreateWorker(worker)
		if err == nil {
			t.Error("Expected error for deprecated worker endpoint")
		}

		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "Worker endpoints are deprecated") {
			t.Errorf("Expected error message to contain deprecation notice, got: %s", errorMsg)
		}
	})

	// Test UpdateWorker operation  
	t.Run("update_worker", func(t *testing.T) {
		_, err := client.UpdateWorker(1, 1, worker)
		if err == nil {
			t.Error("Expected error for deprecated worker endpoint")
		}

		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "Worker endpoints are deprecated") {
			t.Errorf("Expected error message to contain deprecation notice, got: %s", errorMsg)
		}
	})

	// Test DeleteWorker operation
	t.Run("delete_worker", func(t *testing.T) {
		err := client.DeleteWorker(1, 1)
		if err == nil {
			t.Error("Expected error for deprecated worker endpoint")
		}

		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "Worker endpoints are deprecated") {
			t.Errorf("Expected error message to contain deprecation notice, got: %s", errorMsg)
		}
	})
}