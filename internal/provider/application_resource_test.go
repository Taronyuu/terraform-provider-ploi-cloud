package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

func TestApplicationResource_Schema(t *testing.T) {
	r := NewApplicationResource()
	
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	
	r.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("Schema attributes should not be nil")
	}
}

func TestApplicationResource_StartCommand_toAPIModel(t *testing.T) {
	resource := &ApplicationResource{}
	
	tests := []struct {
		name         string
		data         *ApplicationResourceModel
		expectedCmd  string
	}{
		{
			name: "application with custom start command",
			data: &ApplicationResourceModel{
				ID:           types.Int64Value(1),
				Name:         types.StringValue("test-app"),
				Type:         types.StringValue("laravel"),
				StartCommand: types.StringValue("php artisan octane:start --host=0.0.0.0 --port=8000"),
			},
			expectedCmd: "php artisan octane:start --host=0.0.0.0 --port=8000",
		},
		{
			name: "application with nodejs start command",
			data: &ApplicationResourceModel{
				ID:           types.Int64Value(2),
				Name:         types.StringValue("node-app"),
				Type:         types.StringValue("nodejs"),
				StartCommand: types.StringValue("npm start"),
			},
			expectedCmd: "npm start",
		},
		{
			name: "application with null start command",
			data: &ApplicationResourceModel{
				ID:           types.Int64Value(3),
				Name:         types.StringValue("default-app"),
				Type:         types.StringValue("laravel"),
				StartCommand: types.StringNull(),
			},
			expectedCmd: "",
		},
		{
			name: "application with empty start command",
			data: &ApplicationResourceModel{
				ID:           types.Int64Value(4),
				Name:         types.StringValue("empty-cmd-app"),
				Type:         types.StringValue("laravel"),
				StartCommand: types.StringValue(""),
			},
			expectedCmd: "",
		},
		{
			name: "application with complex start command",
			data: &ApplicationResourceModel{
				ID:           types.Int64Value(5),
				Name:         types.StringValue("complex-app"),
				Type:         types.StringValue("nodejs"),
				StartCommand: types.StringValue("node --max-old-space-size=2048 server.js --port=3000"),
			},
			expectedCmd: "node --max-old-space-size=2048 server.js --port=3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resource.toAPIModel(tt.data)
			
			if result.StartCommand != tt.expectedCmd {
				t.Errorf("Expected StartCommand '%s', got '%s'", tt.expectedCmd, result.StartCommand)
			}
			
			// Verify other fields are preserved
			if result.ID != tt.data.ID.ValueInt64() {
				t.Errorf("Expected ID %d, got %d", tt.data.ID.ValueInt64(), result.ID)
			}
			if result.Name != tt.data.Name.ValueString() {
				t.Errorf("Expected Name '%s', got '%s'", tt.data.Name.ValueString(), result.Name)
			}
			if result.Type != tt.data.Type.ValueString() {
				t.Errorf("Expected Type '%s', got '%s'", tt.data.Type.ValueString(), result.Type)
			}
		})
	}
}

func TestApplicationResource_StartCommand_fromAPIModel(t *testing.T) {
	resource := &ApplicationResource{}
	
	tests := []struct {
		name        string
		app         *client.Application
		expectedCmd types.String
	}{
		{
			name: "application with start command from API",
			app: &client.Application{
				ID:           1,
				Name:         "test-app",
				Type:         "laravel",
				StartCommand: "php artisan octane:start --host=0.0.0.0 --port=8000",
				Status:       "running",
			},
			expectedCmd: types.StringValue("php artisan octane:start --host=0.0.0.0 --port=8000"),
		},
		{
			name: "application with empty start command from API",
			app: &client.Application{
				ID:           2,
				Name:         "default-app",
				Type:         "laravel",
				StartCommand: "",
				Status:       "running",
			},
			expectedCmd: types.StringNull(),
		},
		{
			name: "application with nodejs start command",
			app: &client.Application{
				ID:           3,
				Name:         "node-app",
				Type:         "nodejs",
				StartCommand: "npm start",
				Status:       "running",
			},
			expectedCmd: types.StringValue("npm start"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data ApplicationResourceModel
			
			// Initialize with null to test fromAPIModel behavior
			data.StartCommand = types.StringNull()
			
			resource.fromAPIModel(tt.app, &data)
			
			if !data.StartCommand.Equal(tt.expectedCmd) {
				t.Errorf("Expected StartCommand %v, got %v", tt.expectedCmd, data.StartCommand)
			}
			
			// Verify other fields are set correctly
			expectedID := types.Int64Value(tt.app.ID)
			if !data.ID.Equal(expectedID) {
				t.Errorf("Expected ID %v, got %v", expectedID, data.ID)
			}
			
			expectedName := types.StringValue(tt.app.Name)
			if !data.Name.Equal(expectedName) {
				t.Errorf("Expected Name %v, got %v", expectedName, data.Name)
			}
			
			expectedType := types.StringValue(tt.app.Type)
			if !data.Type.Equal(expectedType) {
				t.Errorf("Expected Type %v, got %v", expectedType, data.Type)
			}
		})
	}
}

func TestApplicationResource_StartCommand_BackwardCompatibility(t *testing.T) {
	resource := &ApplicationResource{}
	
	// Test that existing application configurations without start_command still work
	data := &ApplicationResourceModel{
		ID:           types.Int64Value(1),
		Name:         types.StringValue("legacy-app"),
		Type:         types.StringValue("laravel"),
		// StartCommand is null/unset (backward compatibility)
		StartCommand: types.StringNull(),
	}
	
	result := resource.toAPIModel(data)
	
	// Verify basic fields are preserved
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
	if result.Name != "legacy-app" {
		t.Errorf("Expected Name 'legacy-app', got %s", result.Name)
	}
	if result.Type != "laravel" {
		t.Errorf("Expected Type 'laravel', got %s", result.Type)
	}
	
	// Verify start_command has default/empty value
	if result.StartCommand != "" {
		t.Errorf("Expected StartCommand to be empty, got %s", result.StartCommand)
	}
}

func TestApplicationResource_StartCommand_WithOtherFields(t *testing.T) {
	resource := &ApplicationResource{}
	
	// Test that start_command works correctly with other application fields
	data := &ApplicationResourceModel{
		ID:                 types.Int64Value(1),
		Name:               types.StringValue("full-app"),
		Type:               types.StringValue("laravel"),
		ApplicationVersion: types.StringValue("11.x"),
		StartCommand:       types.StringValue("php artisan octane:start --host=0.0.0.0"),
		BuildCommands: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("composer install --no-dev"),
			types.StringValue("php artisan config:cache"),
		}),
		InitCommands: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("php artisan migrate --force"),
		}),
		Runtime: &RuntimeModel{
			PHPVersion: types.StringValue("8.3"),
		},
		Settings: &SettingsModel{
			MemoryRequest: types.StringValue("1Gi"),
			Replicas:      types.Int64Value(2),
		},
	}
	
	result := resource.toAPIModel(data)
	
	// Verify start_command is preserved
	if result.StartCommand != "php artisan octane:start --host=0.0.0.0" {
		t.Errorf("Expected StartCommand 'php artisan octane:start --host=0.0.0.0', got '%s'", result.StartCommand)
	}
	
	// Verify other fields are also preserved
	if result.ApplicationVersion != "11.x" {
		t.Errorf("Expected ApplicationVersion '11.x', got '%s'", result.ApplicationVersion)
	}
	if result.PHPVersion != "8.3" {
		t.Errorf("Expected PHPVersion '8.3', got '%s'", result.PHPVersion)
	}
	if result.MemoryRequest != "1Gi" {
		t.Errorf("Expected MemoryRequest '1Gi', got '%s'", result.MemoryRequest)
	}
	if result.Replicas != 2 {
		t.Errorf("Expected Replicas 2, got %d", result.Replicas)
	}
	if len(result.BuildCommands) != 2 {
		t.Errorf("Expected 2 build commands, got %d", len(result.BuildCommands))
	}
	if len(result.InitCommands) != 1 {
		t.Errorf("Expected 1 init command, got %d", len(result.InitCommands))
	}
}

func TestApplicationResource_StartCommand_UpdateAPIModel(t *testing.T) {
	resource := &ApplicationResource{}
	
	// Test the toUpdateAPIModel method with start_command
	data := &ApplicationResourceModel{
		StartCommand: types.StringValue("node --max-old-space-size=4096 app.js"),
		Runtime: &RuntimeModel{
			NodeJSVersion: types.StringValue("20"),
		},
		Settings: &SettingsModel{
			MemoryRequest: types.StringValue("2Gi"),
			Replicas:      types.Int64Value(3),
		},
	}
	
	result := resource.toUpdateAPIModel(data)
	
	// StartCommand is now included in update model as part of consistency fixes
	if _, exists := result["start_command"]; !exists {
		t.Error("StartCommand should be included in update model to prevent consistency errors")
	}
	
	if result["start_command"] != "node --max-old-space-size=4096 app.js" {
		t.Errorf("Expected start_command 'node --max-old-space-size=4096 app.js', got '%v'", result["start_command"])
	}
	
	// Verify other fields are included
	if result["nodejs_version"] != "20" {
		t.Errorf("Expected nodejs_version '20', got '%v'", result["nodejs_version"])
	}
	if result["memory_request"] != "2Gi" {
		t.Errorf("Expected memory_request '2Gi', got '%v'", result["memory_request"])
	}
	if result["replicas"] != int64(3) {
		t.Errorf("Expected replicas 3, got '%v'", result["replicas"])
	}
}

func TestApplicationResource_StartCommand_EmptyStringHandling(t *testing.T) {
	resource := &ApplicationResource{}
	
	tests := []struct {
		name        string
		input       types.String
		expected    string
		description string
	}{
		{
			name:        "null start command",
			input:       types.StringNull(),
			expected:    "",
			description: "null start command should result in empty string",
		},
		{
			name:        "empty start command",
			input:       types.StringValue(""),
			expected:    "",
			description: "empty start command should result in empty string",
		},
		{
			name:        "whitespace start command",
			input:       types.StringValue("   "),
			expected:    "   ",
			description: "whitespace-only start command should be preserved",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &ApplicationResourceModel{
				Name:         types.StringValue("test-app"),
				Type:         types.StringValue("nodejs"),
				StartCommand: tt.input,
			}
			
			result := resource.toAPIModel(data)
			
			if result.StartCommand != tt.expected {
				t.Errorf("%s: Expected '%s', got '%s'", tt.description, tt.expected, result.StartCommand)
			}
		})
	}
}

func TestApplicationResource_StartCommand_APIClientIntegration(t *testing.T) {
	// Mock server for testing API interactions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.URL.Path {
		case "/applications":
			if r.Method == http.MethodPost {
				// Return an application with start_command
				response := `{
					"success": true,
					"data": {
						"id": 1,
						"name": "test-app",
						"application_type": "nodejs",
						"start_command": "npm run start:prod",
						"status": "running",
						"url": "https://test-app.ploi.cloud",
						"needs_deployment": false
					}
				}`
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(response))
			}
		case "/applications/1":
			if r.Method == http.MethodGet {
				response := `{
					"success": true,
					"data": {
						"id": 1,
						"name": "test-app",
						"application_type": "nodejs",
						"start_command": "npm run start:prod",
						"status": "running",
						"url": "https://test-app.ploi.cloud",
						"needs_deployment": false
					}
				}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client with test server
	c := client.NewClient("test-token", &server.URL)
	
	// Test application creation with start_command
	app := &client.Application{
		Name:         "test-app",
		Type:         "nodejs",
		StartCommand: "npm run start:prod",
	}
	
	created, err := c.CreateApplication(app)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}
	
	// Verify response includes start_command
	if created.StartCommand != "npm run start:prod" {
		t.Errorf("Expected StartCommand 'npm run start:prod', got '%s'", created.StartCommand)
	}
}

// Mock client for testing without network calls
type MockApplicationClient struct {
	apps   map[int64]*client.Application
	nextID int64
}

func NewMockApplicationClient() *MockApplicationClient {
	return &MockApplicationClient{
		apps:   make(map[int64]*client.Application),
		nextID: 1,
	}
}

func (m *MockApplicationClient) CreateApplication(app *client.Application) (*client.Application, error) {
	app.ID = m.nextID
	app.Status = "creating"
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()
	
	m.apps[app.ID] = app
	m.nextID++
	
	return app, nil
}

func (m *MockApplicationClient) GetApplication(appID int64) (*client.Application, error) {
	app, exists := m.apps[appID]
	if !exists {
		return nil, fmt.Errorf("application not found")
	}
	return app, nil
}

func TestApplicationResource_StartCommand_CRUDOperations(t *testing.T) {
	mockClient := NewMockApplicationClient()
	resource := &ApplicationResource{client: nil} // We'll mock the client methods
	
	// Test Create with start_command
	data := &ApplicationResourceModel{
		Name:         types.StringValue("test-app"),
		Type:         types.StringValue("nodejs"),
		StartCommand: types.StringValue("node --experimental-modules server.js"),
	}
	
	apiModel := resource.toAPIModel(data)
	created, err := mockClient.CreateApplication(apiModel)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}
	
	// Verify creation includes start_command
	if created.ID == 0 {
		t.Error("Expected application to have an ID after creation")
	}
	if created.StartCommand != "node --experimental-modules server.js" {
		t.Errorf("Expected StartCommand 'node --experimental-modules server.js', got '%s'", created.StartCommand)
	}
	
	// Test Read
	retrieved, err := mockClient.GetApplication(created.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve application: %v", err)
	}
	
	if retrieved.StartCommand != "node --experimental-modules server.js" {
		t.Errorf("Expected StartCommand 'node --experimental-modules server.js', got '%s'", retrieved.StartCommand)
	}
}

func TestApplicationResource_StartCommand_ConversionAccuracy(t *testing.T) {
	resource := &ApplicationResource{}
	
	// Test round-trip conversion (terraform -> api -> terraform)
	originalData := &ApplicationResourceModel{
		Name:         types.StringValue("conversion-test"),
		Type:         types.StringValue("laravel"),
		StartCommand: types.StringValue("php artisan octane:start --watch"),
	}
	
	// Convert to API model
	apiModel := resource.toAPIModel(originalData)
	
	// Convert back from API model
	var convertedData ApplicationResourceModel
	resource.fromAPIModel(apiModel, &convertedData)
	
	// Verify round-trip accuracy
	if !convertedData.StartCommand.Equal(originalData.StartCommand) {
		t.Errorf("Round-trip conversion failed: expected %v, got %v", 
			originalData.StartCommand, convertedData.StartCommand)
	}
}

func TestApplicationResource_AdditionalDomains_toAPIModel(t *testing.T) {
	resource := &ApplicationResource{}
	
	tests := []struct {
		name            string
		data            *ApplicationResourceModel
		expectedDomains []string
	}{
		{
			name: "application with additional domains",
			data: &ApplicationResourceModel{
				ID:   types.Int64Value(1),
				Name: types.StringValue("test-app"),
				Type: types.StringValue("laravel"),
				AdditionalDomains: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("api.example.com"),
					types.StringValue("admin.example.com"),
				}),
			},
			expectedDomains: []string{"api.example.com", "admin.example.com"},
		},
		{
			name: "application with single additional domain",
			data: &ApplicationResourceModel{
				ID:   types.Int64Value(2),
				Name: types.StringValue("single-domain-app"),
				Type: types.StringValue("nodejs"),
				AdditionalDomains: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("www.example.com"),
				}),
			},
			expectedDomains: []string{"www.example.com"},
		},
		{
			name: "application with null additional domains",
			data: &ApplicationResourceModel{
				ID:                types.Int64Value(3),
				Name:              types.StringValue("no-domains-app"),
				Type:              types.StringValue("laravel"),
				AdditionalDomains: types.ListNull(types.StringType),
			},
			expectedDomains: []string{},
		},
		{
			name: "application with empty additional domains list",
			data: &ApplicationResourceModel{
				ID:                types.Int64Value(4),
				Name:              types.StringValue("empty-domains-app"),
				Type:              types.StringValue("laravel"),
				AdditionalDomains: types.ListValueMust(types.StringType, []attr.Value{}),
			},
			expectedDomains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resource.toAPIModel(tt.data)
			
			if len(result.Domains) != len(tt.expectedDomains) {
				t.Errorf("Expected %d domains, got %d", len(tt.expectedDomains), len(result.Domains))
				return
			}
			
			// Verify each domain
			for i, expectedDomain := range tt.expectedDomains {
				if result.Domains[i].Domain != expectedDomain {
					t.Errorf("Expected domain[%d] '%s', got '%s'", i, expectedDomain, result.Domains[i].Domain)
				}
			}
			
			// Verify other fields are preserved
			if result.ID != tt.data.ID.ValueInt64() {
				t.Errorf("Expected ID %d, got %d", tt.data.ID.ValueInt64(), result.ID)
			}
			if result.Name != tt.data.Name.ValueString() {
				t.Errorf("Expected Name '%s', got '%s'", tt.data.Name.ValueString(), result.Name)
			}
		})
	}
}

func TestApplicationResource_AdditionalDomains_fromAPIModel(t *testing.T) {
	resource := &ApplicationResource{}
	
	tests := []struct {
		name               string
		app                *client.Application
		expectedDomains    types.List
	}{
		{
			name: "application with domains from API",
			app: &client.Application{
				ID:   1,
				Name: "test-app",
				Type: "laravel",
				Domains: []client.ApplicationDomain{
					{Domain: "api.example.com"},
					{Domain: "admin.example.com"},
				},
				Status: "running",
			},
			expectedDomains: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("api.example.com"),
				types.StringValue("admin.example.com"),
			}),
		},
		{
			name: "application with single domain from API",
			app: &client.Application{
				ID:   2,
				Name: "single-domain-app",
				Type: "nodejs",
				Domains: []client.ApplicationDomain{
					{Domain: "www.example.com"},
				},
				Status: "running",
			},
			expectedDomains: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("www.example.com"),
			}),
		},
		{
			name: "application with empty domains from API",
			app: &client.Application{
				ID:      3,
				Name:    "no-domains-app",
				Type:    "laravel",
				Domains: []client.ApplicationDomain{},
				Status:  "running",
			},
			expectedDomains: types.ListNull(types.StringType),
		},
		{
			name: "application with nil domains from API",
			app: &client.Application{
				ID:      4,
				Name:    "nil-domains-app",
				Type:    "laravel",
				Domains: nil,
				Status:  "running",
			},
			expectedDomains: types.ListNull(types.StringType),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data ApplicationResourceModel
			
			// Initialize with null to test fromAPIModel behavior
			data.AdditionalDomains = types.ListNull(types.StringType)
			
			resource.fromAPIModel(tt.app, &data)
			
			// For empty/nil domains, both should be null/empty
			if len(tt.app.Domains) == 0 {
				if !data.AdditionalDomains.IsNull() && len(data.AdditionalDomains.Elements()) > 0 {
					t.Errorf("Expected AdditionalDomains to be null or empty, got %v", data.AdditionalDomains)
				}
			} else {
				if !data.AdditionalDomains.Equal(tt.expectedDomains) {
					t.Errorf("Expected AdditionalDomains %v, got %v", tt.expectedDomains, data.AdditionalDomains)
				}
			}
		})
	}
}

func TestApplicationResource_AdditionalDomains_UpdateAPIModel(t *testing.T) {
	resource := &ApplicationResource{}
	
	tests := []struct {
		name            string
		data            *ApplicationResourceModel
		expectedDomains []string
		shouldInclude   bool
	}{
		{
			name: "update with additional domains",
			data: &ApplicationResourceModel{
				AdditionalDomains: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("new-api.example.com"),
					types.StringValue("new-admin.example.com"),
				}),
			},
			expectedDomains: []string{"new-api.example.com", "new-admin.example.com"},
			shouldInclude:   true,
		},
		{
			name: "update with null domains",
			data: &ApplicationResourceModel{
				AdditionalDomains: types.ListNull(types.StringType),
			},
			expectedDomains: []string{},
			shouldInclude:   false,
		},
		{
			name: "update with empty domains list",
			data: &ApplicationResourceModel{
				AdditionalDomains: types.ListValueMust(types.StringType, []attr.Value{}),
			},
			expectedDomains: []string{},
			shouldInclude:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resource.toUpdateAPIModel(tt.data)
			
			domains, exists := result["additional_domains"]
			if tt.shouldInclude {
				if !exists {
					t.Error("Expected additional_domains to be included in update")
					return
				}
				
				domainStrings, ok := domains.([]string)
				if !ok {
					t.Errorf("Expected additional_domains to be []string, got %T", domains)
					return
				}
				
				if len(domainStrings) != len(tt.expectedDomains) {
					t.Errorf("Expected %d domains, got %d", len(tt.expectedDomains), len(domainStrings))
					return
				}
				
				for i, expected := range tt.expectedDomains {
					if domainStrings[i] != expected {
						t.Errorf("Expected domain[%d] '%s', got '%s'", i, expected, domainStrings[i])
					}
				}
			} else {
				if exists {
					t.Error("Expected additional_domains to not be included in update when null/empty")
				}
			}
		})
	}
}

func TestApplicationResource_AdditionalDomains_BackwardCompatibility(t *testing.T) {
	resource := &ApplicationResource{}
	
	// Test that existing application configurations without additional_domains still work
	data := &ApplicationResourceModel{
		ID:   types.Int64Value(1),
		Name: types.StringValue("legacy-app"),
		Type: types.StringValue("laravel"),
		// AdditionalDomains is null/unset (backward compatibility)
		AdditionalDomains: types.ListNull(types.StringType),
	}
	
	result := resource.toAPIModel(data)
	
	// Verify basic fields are preserved
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
	if result.Name != "legacy-app" {
		t.Errorf("Expected Name 'legacy-app', got %s", result.Name)
	}
	if result.Type != "laravel" {
		t.Errorf("Expected Type 'laravel', got %s", result.Type)
	}
	
	// Verify domains are empty when null
	if len(result.Domains) != 0 {
		t.Errorf("Expected no domains, got %d", len(result.Domains))
	}
}

func TestApplicationResource_AdditionalDomains_ConversionAccuracy(t *testing.T) {
	resource := &ApplicationResource{}
	
	// Test round-trip conversion (terraform -> api -> terraform)
	originalData := &ApplicationResourceModel{
		Name: types.StringValue("conversion-test"),
		Type: types.StringValue("laravel"),
		AdditionalDomains: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("test1.example.com"),
			types.StringValue("test2.example.com"),
		}),
	}
	
	// Convert to API model
	apiModel := resource.toAPIModel(originalData)
	
	// Convert back from API model
	var convertedData ApplicationResourceModel
	resource.fromAPIModel(apiModel, &convertedData)
	
	// Verify round-trip accuracy
	if !convertedData.AdditionalDomains.Equal(originalData.AdditionalDomains) {
		t.Errorf("Round-trip conversion failed: expected %v, got %v", 
			originalData.AdditionalDomains, convertedData.AdditionalDomains)
	}
}

func TestApplicationResource_AdditionalDomains_WithOtherFields(t *testing.T) {
	resource := &ApplicationResource{}
	
	// Test that additional_domains works correctly with other application fields
	data := &ApplicationResourceModel{
		ID:                 types.Int64Value(1),
		Name:               types.StringValue("full-app"),
		Type:               types.StringValue("laravel"),
		ApplicationVersion: types.StringValue("11.x"),
		StartCommand:       types.StringValue("php artisan serve"),
		AdditionalDomains: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("api.full-app.com"),
			types.StringValue("admin.full-app.com"),
		}),
		BuildCommands: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("composer install --no-dev"),
		}),
		Runtime: &RuntimeModel{
			PHPVersion: types.StringValue("8.3"),
		},
		Settings: &SettingsModel{
			MemoryRequest: types.StringValue("1Gi"),
			Replicas:      types.Int64Value(2),
		},
	}
	
	result := resource.toAPIModel(data)
	
	// Verify additional_domains are preserved
	if len(result.Domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(result.Domains))
	}
	if result.Domains[0].Domain != "api.full-app.com" {
		t.Errorf("Expected first domain 'api.full-app.com', got '%s'", result.Domains[0].Domain)
	}
	if result.Domains[1].Domain != "admin.full-app.com" {
		t.Errorf("Expected second domain 'admin.full-app.com', got '%s'", result.Domains[1].Domain)
	}
	
	// Verify other fields are also preserved
	if result.ApplicationVersion != "11.x" {
		t.Errorf("Expected ApplicationVersion '11.x', got '%s'", result.ApplicationVersion)
	}
	if result.StartCommand != "php artisan serve" {
		t.Errorf("Expected StartCommand 'php artisan serve', got '%s'", result.StartCommand)
	}
	if result.PHPVersion != "8.3" {
		t.Errorf("Expected PHPVersion '8.3', got '%s'", result.PHPVersion)
	}
	if result.MemoryRequest != "1Gi" {
		t.Errorf("Expected MemoryRequest '1Gi', got '%s'", result.MemoryRequest)
	}
	if result.Replicas != 2 {
		t.Errorf("Expected Replicas 2, got %d", result.Replicas)
	}
	if len(result.BuildCommands) != 1 {
		t.Errorf("Expected 1 build command, got %d", len(result.BuildCommands))
	}
}