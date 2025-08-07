package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

func TestServiceResource_Schema(t *testing.T) {
	r := NewServiceResource()
	
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	
	r.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("Schema attributes should not be nil")
	}
}

func TestServiceResource_toAPIModel(t *testing.T) {
	resource := &ServiceResource{}
	
	tests := []struct {
		name     string
		data     *ServiceResourceModel
		expected *client.ApplicationService
	}{
		{
			name: "basic service with new fields",
			data: &ServiceResourceModel{
				ID:            types.Int64Value(1),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("test-service"),
				Type:          types.StringValue("postgresql"),
				Version:       types.StringValue("15"),
				MemoryRequest: types.StringValue("1Gi"),
				StorageSize:   types.StringValue("10Gi"),
				Extensions:    types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("uuid-ossp"),
					types.StringValue("pgcrypto"),
				}),
			},
			expected: &client.ApplicationService{
				ID:            1,
				ApplicationID: 100,
				Name:          "test-service",
				Type:          "postgresql",
				Version:       "15",
				MemoryRequest: "1Gi",
				StorageSize:   "10Gi",
				Extensions:    []string{"uuid-ossp", "pgcrypto"},
			},
		},
		{
			name: "service without extensions",
			data: &ServiceResourceModel{
				ID:            types.Int64Value(2),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("redis-service"),
				Type:          types.StringValue("redis"),
				MemoryRequest: types.StringValue("512Mi"),
				StorageSize:   types.StringValue("5Gi"),
				Extensions:    types.ListNull(types.StringType),
			},
			expected: &client.ApplicationService{
				ID:            2,
				ApplicationID: 100,
				Name:          "redis-service",
				Type:          "redis",
				MemoryRequest: "512Mi",
				StorageSize:   "5Gi",
				Extensions:    nil,
			},
		},
		{
			name: "service with null values",
			data: &ServiceResourceModel{
				ID:            types.Int64Value(3),
				ApplicationID: types.Int64Value(100),
				Type:          types.StringValue("mysql"),
				MemoryRequest: types.StringNull(),
				StorageSize:   types.StringNull(),
				Extensions:    types.ListNull(types.StringType),
			},
			expected: &client.ApplicationService{
				ID:            3,
				ApplicationID: 100,
				Type:          "mysql",
				MemoryRequest: "",
				StorageSize:   "",
				Extensions:    nil,
			},
		},
		{
			name: "service with empty string values",
			data: &ServiceResourceModel{
				ID:            types.Int64Value(4),
				ApplicationID: types.Int64Value(100),
				Type:          types.StringValue("mysql"),
				MemoryRequest: types.StringValue(""),
				StorageSize:   types.StringValue(""),
				Extensions:    types.ListValueMust(types.StringType, []attr.Value{}),
			},
			expected: &client.ApplicationService{
				ID:            4,
				ApplicationID: 100,
				Type:          "mysql",
				MemoryRequest: "",
				StorageSize:   "",
				Extensions:    []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resource.toAPIModel(tt.data)
			
			if result.ID != tt.expected.ID {
				t.Errorf("Expected ID %d, got %d", tt.expected.ID, result.ID)
			}
			if result.ApplicationID != tt.expected.ApplicationID {
				t.Errorf("Expected ApplicationID %d, got %d", tt.expected.ApplicationID, result.ApplicationID)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Expected Name %s, got %s", tt.expected.Name, result.Name)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Expected Type %s, got %s", tt.expected.Type, result.Type)
			}
			if result.Version != tt.expected.Version {
				t.Errorf("Expected Version %s, got %s", tt.expected.Version, result.Version)
			}
			if result.MemoryRequest != tt.expected.MemoryRequest {
				t.Errorf("Expected MemoryRequest %s, got %s", tt.expected.MemoryRequest, result.MemoryRequest)
			}
			if result.StorageSize != tt.expected.StorageSize {
				t.Errorf("Expected StorageSize %s, got %s", tt.expected.StorageSize, result.StorageSize)
			}
			if !reflect.DeepEqual(result.Extensions, tt.expected.Extensions) {
				t.Errorf("Expected Extensions %v, got %v", tt.expected.Extensions, result.Extensions)
			}
		})
	}
}

func TestServiceResource_fromAPIModel(t *testing.T) {
	resource := &ServiceResource{}
	
	tests := []struct {
		name     string
		service  *client.ApplicationService
		expected ServiceResourceModel
	}{
		{
			name: "service with all new fields",
			service: &client.ApplicationService{
				ID:            1,
				ApplicationID: 100,
				Name:          "test-service",
				Type:          "postgresql",
				Version:       "15",
				Status:        "running",
				MemoryRequest: "1Gi",
				StorageSize:   "10Gi",
				Extensions:    []string{"uuid-ossp", "pgcrypto"},
			},
			expected: ServiceResourceModel{
				ID:            types.Int64Value(1),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("test-service"),
				Type:          types.StringValue("postgresql"),
				Version:       types.StringValue("15"),
				Status:        types.StringValue("running"),
				MemoryRequest: types.StringValue("1Gi"),
				StorageSize:   types.StringValue("10Gi"),
			},
		},
		{
			name: "service without extensions",
			service: &client.ApplicationService{
				ID:            2,
				ApplicationID: 100,
				Name:          "redis-service",
				Type:          "redis",
				Status:        "running",
				MemoryRequest: "512Mi",
				StorageSize:   "5Gi",
				Extensions:    nil,
			},
			expected: ServiceResourceModel{
				ID:            types.Int64Value(2),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("redis-service"),
				Type:          types.StringValue("redis"),
				Status:        types.StringValue("running"),
				MemoryRequest: types.StringValue("512Mi"),
				StorageSize:   types.StringValue("5Gi"),
			},
		},
		{
			name: "service with empty extensions",
			service: &client.ApplicationService{
				ID:            3,
				ApplicationID: 100,
				Name:          "mysql-service",
				Type:          "mysql",
				Status:        "running",
				MemoryRequest: "256Mi",
				StorageSize:   "1Gi",
				Extensions:    []string{},
			},
			expected: ServiceResourceModel{
				ID:            types.Int64Value(3),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("mysql-service"),
				Type:          types.StringValue("mysql"),
				Status:        types.StringValue("running"),
				MemoryRequest: types.StringValue("256Mi"),
				StorageSize:   types.StringValue("1Gi"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data ServiceResourceModel
			resource.fromAPIModel(tt.service, &data)
			
			if !data.ID.Equal(tt.expected.ID) {
				t.Errorf("Expected ID %v, got %v", tt.expected.ID, data.ID)
			}
			if !data.ApplicationID.Equal(tt.expected.ApplicationID) {
				t.Errorf("Expected ApplicationID %v, got %v", tt.expected.ApplicationID, data.ApplicationID)
			}
			if !data.Name.Equal(tt.expected.Name) {
				t.Errorf("Expected Name %v, got %v", tt.expected.Name, data.Name)
			}
			if !data.Type.Equal(tt.expected.Type) {
				t.Errorf("Expected Type %v, got %v", tt.expected.Type, data.Type)
			}
			if !data.Version.Equal(tt.expected.Version) {
				t.Errorf("Expected Version %v, got %v", tt.expected.Version, data.Version)
			}
			if !data.Status.Equal(tt.expected.Status) {
				t.Errorf("Expected Status %v, got %v", tt.expected.Status, data.Status)
			}
			if !data.MemoryRequest.Equal(tt.expected.MemoryRequest) {
				t.Errorf("Expected MemoryRequest %v, got %v", tt.expected.MemoryRequest, data.MemoryRequest)
			}
			if !data.StorageSize.Equal(tt.expected.StorageSize) {
				t.Errorf("Expected StorageSize %v, got %v", tt.expected.StorageSize, data.StorageSize)
			}
			
			// Test extensions handling
			if len(tt.service.Extensions) > 0 {
				if data.Extensions.IsNull() {
					t.Error("Expected Extensions to not be null")
				}
				// Verify extensions content
				var extensions []string
				data.Extensions.ElementsAs(context.Background(), &extensions, false)
				if !reflect.DeepEqual(extensions, tt.service.Extensions) {
					t.Errorf("Expected Extensions %v, got %v", tt.service.Extensions, extensions)
				}
			} else {
				if !data.Extensions.IsNull() {
					t.Error("Expected Extensions to be null when service has no extensions")
				}
			}
		})
	}
}

func TestServiceResource_PostgreSQLExtensionsValidation(t *testing.T) {
	resource := &ServiceResource{}
	
	tests := []struct {
		name       string
		serviceType string
		extensions []string
		shouldPass bool
	}{
		{
			name:       "postgresql with valid extensions",
			serviceType: "postgresql",
			extensions: []string{"uuid-ossp", "pgcrypto", "hstore"},
			shouldPass: true,
		},
		{
			name:       "mysql with extensions should be ignored",
			serviceType: "mysql",
			extensions: []string{"uuid-ossp"},
			shouldPass: true,
		},
		{
			name:       "redis with extensions should be ignored",
			serviceType: "redis",
			extensions: []string{"some-extension"},
			shouldPass: true,
		},
		{
			name:       "postgresql without extensions",
			serviceType: "postgresql",
			extensions: nil,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &ServiceResourceModel{
				ApplicationID: types.Int64Value(100),
				Type:          types.StringValue(tt.serviceType),
			}
			
			if tt.extensions != nil {
				extensionValues := make([]attr.Value, len(tt.extensions))
				for i, ext := range tt.extensions {
					extensionValues[i] = types.StringValue(ext)
				}
				data.Extensions = types.ListValueMust(types.StringType, extensionValues)
			} else {
				data.Extensions = types.ListNull(types.StringType)
			}
			
			result := resource.toAPIModel(data)
			
			// For non-PostgreSQL services, extensions should still be passed through
			// The API should handle ignoring them for inappropriate service types
			if tt.extensions != nil {
				if !reflect.DeepEqual(result.Extensions, tt.extensions) {
					t.Errorf("Expected Extensions %v, got %v", tt.extensions, result.Extensions)
				}
			} else {
				if result.Extensions != nil {
					t.Errorf("Expected Extensions to be nil, got %v", result.Extensions)
				}
			}
		})
	}
}

func TestServiceResource_BackwardCompatibility(t *testing.T) {
	resource := &ServiceResource{}
	
	// Test that existing service configurations without new fields still work
	data := &ServiceResourceModel{
		ID:            types.Int64Value(1),
		ApplicationID: types.Int64Value(100),
		Name:          types.StringValue("legacy-service"),
		Type:          types.StringValue("mysql"),
		Version:       types.StringValue("8.0"),
		// New fields are null/unset
		MemoryRequest: types.StringNull(),
		StorageSize:   types.StringNull(),
		Extensions:    types.ListNull(types.StringType),
	}
	
	result := resource.toAPIModel(data)
	
	// Verify basic fields are preserved
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
	if result.ApplicationID != 100 {
		t.Errorf("Expected ApplicationID 100, got %d", result.ApplicationID)
	}
	if result.Name != "legacy-service" {
		t.Errorf("Expected Name 'legacy-service', got %s", result.Name)
	}
	if result.Type != "mysql" {
		t.Errorf("Expected Type 'mysql', got %s", result.Type)
	}
	if result.Version != "8.0" {
		t.Errorf("Expected Version '8.0', got %s", result.Version)
	}
	
	// Verify new fields have default/empty values
	if result.MemoryRequest != "" {
		t.Errorf("Expected MemoryRequest to be empty, got %s", result.MemoryRequest)
	}
	if result.StorageSize != "" {
		t.Errorf("Expected StorageSize to be empty, got %s", result.StorageSize)
	}
	if result.Extensions != nil {
		t.Errorf("Expected Extensions to be nil, got %v", result.Extensions)
	}
}

func TestServiceResource_DefaultBehaviors(t *testing.T) {
	resource := &ServiceResource{}
	
	tests := []struct {
		name    string
		data    *ServiceResourceModel
		field   string
		expect  string
	}{
		{
			name: "memory request with default behavior",
			data: &ServiceResourceModel{
				ApplicationID: types.Int64Value(100),
				Type:          types.StringValue("postgresql"),
				MemoryRequest: types.StringValue(""),
			},
			field:  "MemoryRequest",
			expect: "",
		},
		{
			name: "storage size with default behavior", 
			data: &ServiceResourceModel{
				ApplicationID: types.Int64Value(100),
				Type:          types.StringValue("postgresql"),
				StorageSize:   types.StringValue(""),
			},
			field:  "StorageSize",
			expect: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resource.toAPIModel(tt.data)
			
			switch tt.field {
			case "MemoryRequest":
				if result.MemoryRequest != tt.expect {
					t.Errorf("Expected %s to be '%s', got '%s'", tt.field, tt.expect, result.MemoryRequest)
				}
			case "StorageSize":
				if result.StorageSize != tt.expect {
					t.Errorf("Expected %s to be '%s', got '%s'", tt.field, tt.expect, result.StorageSize)
				}
			}
		})
	}
}

func TestServiceResource_APIClientIntegration(t *testing.T) {
	// Mock server for testing API interactions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.URL.Path {
		case "/applications/100/services":
			if r.Method == http.MethodPost {
				// Return a service with enhanced fields
				response := `{
					"success": true,
					"data": {
						"id": 1,
						"application_id": 100,
						"name": "test-service",
						"type": "postgresql",
						"version": "15",
						"status": "running",
						"memory_request": "1Gi",
						"storage_size": "10Gi",
						"extensions": ["uuid-ossp", "pgcrypto"]
					}
				}`
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(response))
			}
		case "/applications/100/services/1":
			if r.Method == http.MethodGet {
				response := `{
					"success": true,
					"data": {
						"id": 1,
						"application_id": 100,
						"name": "test-service",
						"type": "postgresql",
						"version": "15",
						"status": "running",
						"memory_request": "1Gi",
						"storage_size": "10Gi",
						"extensions": ["uuid-ossp", "pgcrypto"]
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
	
	// Test service creation with new fields
	service := &client.ApplicationService{
		ApplicationID: 100,
		Name:          "test-service",
		Type:          "postgresql",
		Version:       "15",
		MemoryRequest: "1Gi",
		StorageSize:   "10Gi",
		Extensions:    []string{"uuid-ossp", "pgcrypto"},
	}
	
	created, err := c.CreateService(service)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	
	// Verify response includes new fields
	if created.MemoryRequest != "1Gi" {
		t.Errorf("Expected MemoryRequest '1Gi', got '%s'", created.MemoryRequest)
	}
	if created.StorageSize != "10Gi" {
		t.Errorf("Expected StorageSize '10Gi', got '%s'", created.StorageSize)
	}
	if len(created.Extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(created.Extensions))
	}
	if !reflect.DeepEqual(created.Extensions, []string{"uuid-ossp", "pgcrypto"}) {
		t.Errorf("Expected extensions ['uuid-ossp', 'pgcrypto'], got %v", created.Extensions)
	}
}

// Mock client for testing without network calls
type MockServiceClient struct {
	services map[int64]*client.ApplicationService
	nextID   int64
}

func NewMockServiceClient() *MockServiceClient {
	return &MockServiceClient{
		services: make(map[int64]*client.ApplicationService),
		nextID:   1,
	}
}

func (m *MockServiceClient) CreateService(service *client.ApplicationService) (*client.ApplicationService, error) {
	service.ID = m.nextID
	service.Status = "creating"
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()
	
	m.services[service.ID] = service
	m.nextID++
	
	return service, nil
}

func (m *MockServiceClient) GetService(appID, serviceID int64) (*client.ApplicationService, error) {
	service, exists := m.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("service not found")
	}
	return service, nil
}

func TestServiceResource_CRUDOperations(t *testing.T) {
	mockClient := NewMockServiceClient()
	resource := &ServiceResource{client: nil} // We'll mock the client methods
	
	// Test Create
	data := &ServiceResourceModel{
		ApplicationID: types.Int64Value(100),
		Name:          types.StringValue("test-service"),
		Type:          types.StringValue("postgresql"),
		Version:       types.StringValue("15"),
		MemoryRequest: types.StringValue("1Gi"),
		StorageSize:   types.StringValue("10Gi"),
		Extensions:    types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("uuid-ossp"),
			types.StringValue("pgcrypto"),
		}),
	}
	
	apiModel := resource.toAPIModel(data)
	created, err := mockClient.CreateService(apiModel)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	
	// Verify creation
	if created.ID == 0 {
		t.Error("Expected service to have an ID after creation")
	}
	if created.Status != "creating" {
		t.Errorf("Expected status 'creating', got '%s'", created.Status)
	}
	
	// Test Read
	retrieved, err := mockClient.GetService(100, created.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve service: %v", err)
	}
	
	if retrieved.MemoryRequest != "1Gi" {
		t.Errorf("Expected MemoryRequest '1Gi', got '%s'", retrieved.MemoryRequest)
	}
	if retrieved.StorageSize != "10Gi" {
		t.Errorf("Expected StorageSize '10Gi', got '%s'", retrieved.StorageSize)
	}
	if !reflect.DeepEqual(retrieved.Extensions, []string{"uuid-ossp", "pgcrypto"}) {
		t.Errorf("Expected extensions ['uuid-ossp', 'pgcrypto'], got %v", retrieved.Extensions)
	}
}