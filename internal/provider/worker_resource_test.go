package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

func TestWorkerResource_Schema(t *testing.T) {
	r := NewWorkerResource()
	
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	
	r.Schema(context.Background(), req, resp)

	if resp.Schema.Attributes == nil {
		t.Fatal("Schema attributes should not be nil")
	}
}

func TestWorkerResource_toAPIModel(t *testing.T) {
	resource := &WorkerResource{}
	
	tests := []struct {
		name     string
		data     *WorkerResourceModel
		expected *client.Worker
	}{
		{
			name: "worker with all new fields",
			data: &WorkerResourceModel{
				ID:            types.Int64Value(1),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("queue-worker"),
				Command:       types.StringValue("php artisan queue:work"),
				Type:          types.StringValue("queue"),
				Replicas:      types.Int64Value(2),
				MemoryRequest: types.StringValue("512Mi"),
				CPURequest:    types.StringValue("250m"),
			},
			expected: &client.Worker{
				ID:            1,
				ApplicationID: 100,
				Name:          "queue-worker",
				Command:       "php artisan queue:work",
				Type:          "queue",
				Replicas:      2,
				MemoryRequest: "512Mi",
				CPURequest:    "250m",
			},
		},
		{
			name: "worker with default type",
			data: &WorkerResourceModel{
				ID:            types.Int64Value(2),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("default-worker"),
				Command:       types.StringValue("php artisan schedule:work"),
				Type:          types.StringValue("queue"), // Default value
				Replicas:      types.Int64Value(1),
				MemoryRequest: types.StringValue("256Mi"),
				CPURequest:    types.StringValue("100m"),
			},
			expected: &client.Worker{
				ID:            2,
				ApplicationID: 100,
				Name:          "default-worker",
				Command:       "php artisan schedule:work",
				Type:          "queue",
				Replicas:      1,
				MemoryRequest: "256Mi",
				CPURequest:    "100m",
			},
		},
		{
			name: "worker with scheduler type",
			data: &WorkerResourceModel{
				ID:            types.Int64Value(3),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("scheduler-worker"),
				Command:       types.StringValue("php artisan schedule:work"),
				Type:          types.StringValue("scheduler"),
				Replicas:      types.Int64Value(1),
				MemoryRequest: types.StringValue("128Mi"),
				CPURequest:    types.StringValue("50m"),
			},
			expected: &client.Worker{
				ID:            3,
				ApplicationID: 100,
				Name:          "scheduler-worker",
				Command:       "php artisan schedule:work",
				Type:          "scheduler",
				Replicas:      1,
				MemoryRequest: "128Mi",
				CPURequest:    "50m",
			},
		},
		{
			name: "worker with custom type",
			data: &WorkerResourceModel{
				ID:            types.Int64Value(4),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("custom-worker"),
				Command:       types.StringValue("npm run worker"),
				Type:          types.StringValue("custom"),
				Replicas:      types.Int64Value(3),
				MemoryRequest: types.StringValue("1Gi"),
				CPURequest:    types.StringValue("500m"),
			},
			expected: &client.Worker{
				ID:            4,
				ApplicationID: 100,
				Name:          "custom-worker",
				Command:       "npm run worker",
				Type:          "custom",
				Replicas:      3,
				MemoryRequest: "1Gi",
				CPURequest:    "500m",
			},
		},
		{
			name: "worker with null resource requests",
			data: &WorkerResourceModel{
				ID:            types.Int64Value(5),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("minimal-worker"),
				Command:       types.StringValue("php artisan queue:work"),
				Type:          types.StringValue("queue"),
				Replicas:      types.Int64Value(1),
				MemoryRequest: types.StringNull(),
				CPURequest:    types.StringNull(),
			},
			expected: &client.Worker{
				ID:            5,
				ApplicationID: 100,
				Name:          "minimal-worker",
				Command:       "php artisan queue:work",
				Type:          "queue",
				Replicas:      1,
				MemoryRequest: "",
				CPURequest:    "",
			},
		},
		{
			name: "worker with empty string resource requests",
			data: &WorkerResourceModel{
				ID:            types.Int64Value(6),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("empty-resources-worker"),
				Command:       types.StringValue("php artisan queue:work"),
				Type:          types.StringValue("queue"),
				Replicas:      types.Int64Value(1),
				MemoryRequest: types.StringValue(""),
				CPURequest:    types.StringValue(""),
			},
			expected: &client.Worker{
				ID:            6,
				ApplicationID: 100,
				Name:          "empty-resources-worker",
				Command:       "php artisan queue:work",
				Type:          "queue",
				Replicas:      1,
				MemoryRequest: "",
				CPURequest:    "",
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
			if result.Command != tt.expected.Command {
				t.Errorf("Expected Command %s, got %s", tt.expected.Command, result.Command)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Expected Type %s, got %s", tt.expected.Type, result.Type)
			}
			if result.Replicas != tt.expected.Replicas {
				t.Errorf("Expected Replicas %d, got %d", tt.expected.Replicas, result.Replicas)
			}
			if result.MemoryRequest != tt.expected.MemoryRequest {
				t.Errorf("Expected MemoryRequest %s, got %s", tt.expected.MemoryRequest, result.MemoryRequest)
			}
			if result.CPURequest != tt.expected.CPURequest {
				t.Errorf("Expected CPURequest %s, got %s", tt.expected.CPURequest, result.CPURequest)
			}
		})
	}
}

func TestWorkerResource_fromAPIModel(t *testing.T) {
	resource := &WorkerResource{}
	
	tests := []struct {
		name     string
		worker   *client.Worker
		expected WorkerResourceModel
	}{
		{
			name: "worker with all fields",
			worker: &client.Worker{
				ID:            1,
				ApplicationID: 100,
				Name:          "queue-worker",
				Command:       "php artisan queue:work",
				Type:          "queue",
				Replicas:      2,
				MemoryRequest: "512Mi",
				CPURequest:    "250m",
				Status:        "running",
			},
			expected: WorkerResourceModel{
				ID:            types.Int64Value(1),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("queue-worker"),
				Command:       types.StringValue("php artisan queue:work"),
				Type:          types.StringValue("queue"),
				Replicas:      types.Int64Value(2),
				MemoryRequest: types.StringValue("512Mi"),
				CPURequest:    types.StringValue("250m"),
				Status:        types.StringValue("running"),
			},
		},
		{
			name: "worker with scheduler type",
			worker: &client.Worker{
				ID:            2,
				ApplicationID: 100,
				Name:          "scheduler-worker",
				Command:       "php artisan schedule:work",
				Type:          "scheduler",
				Replicas:      1,
				MemoryRequest: "128Mi",
				CPURequest:    "50m",
				Status:        "running",
			},
			expected: WorkerResourceModel{
				ID:            types.Int64Value(2),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("scheduler-worker"),
				Command:       types.StringValue("php artisan schedule:work"),
				Type:          types.StringValue("scheduler"),
				Replicas:      types.Int64Value(1),
				MemoryRequest: types.StringValue("128Mi"),
				CPURequest:    types.StringValue("50m"),
				Status:        types.StringValue("running"),
			},
		},
		{
			name: "worker with custom type",
			worker: &client.Worker{
				ID:            3,
				ApplicationID: 100,
				Name:          "custom-worker",
				Command:       "npm run worker",
				Type:          "custom",
				Replicas:      3,
				MemoryRequest: "1Gi",
				CPURequest:    "500m",
				Status:        "running",
			},
			expected: WorkerResourceModel{
				ID:            types.Int64Value(3),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("custom-worker"),
				Command:       types.StringValue("npm run worker"),
				Type:          types.StringValue("custom"),
				Replicas:      types.Int64Value(3),
				MemoryRequest: types.StringValue("1Gi"),
				CPURequest:    types.StringValue("500m"),
				Status:        types.StringValue("running"),
			},
		},
		{
			name: "worker with empty resource requests",
			worker: &client.Worker{
				ID:            4,
				ApplicationID: 100,
				Name:          "minimal-worker",
				Command:       "php artisan queue:work",
				Type:          "queue",
				Replicas:      1,
				MemoryRequest: "",
				CPURequest:    "",
				Status:        "running",
			},
			expected: WorkerResourceModel{
				ID:            types.Int64Value(4),
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("minimal-worker"),
				Command:       types.StringValue("php artisan queue:work"),
				Type:          types.StringValue("queue"),
				Replicas:      types.Int64Value(1),
				MemoryRequest: types.StringValue(""),
				CPURequest:    types.StringValue(""),
				Status:        types.StringValue("running"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data WorkerResourceModel
			resource.fromAPIModel(tt.worker, &data)
			
			if !data.ID.Equal(tt.expected.ID) {
				t.Errorf("Expected ID %v, got %v", tt.expected.ID, data.ID)
			}
			if !data.ApplicationID.Equal(tt.expected.ApplicationID) {
				t.Errorf("Expected ApplicationID %v, got %v", tt.expected.ApplicationID, data.ApplicationID)
			}
			if !data.Name.Equal(tt.expected.Name) {
				t.Errorf("Expected Name %v, got %v", tt.expected.Name, data.Name)
			}
			if !data.Command.Equal(tt.expected.Command) {
				t.Errorf("Expected Command %v, got %v", tt.expected.Command, data.Command)
			}
			if !data.Type.Equal(tt.expected.Type) {
				t.Errorf("Expected Type %v, got %v", tt.expected.Type, data.Type)
			}
			if !data.Replicas.Equal(tt.expected.Replicas) {
				t.Errorf("Expected Replicas %v, got %v", tt.expected.Replicas, data.Replicas)
			}
			if !data.MemoryRequest.Equal(tt.expected.MemoryRequest) {
				t.Errorf("Expected MemoryRequest %v, got %v", tt.expected.MemoryRequest, data.MemoryRequest)
			}
			if !data.CPURequest.Equal(tt.expected.CPURequest) {
				t.Errorf("Expected CPURequest %v, got %v", tt.expected.CPURequest, data.CPURequest)
			}
			if !data.Status.Equal(tt.expected.Status) {
				t.Errorf("Expected Status %v, got %v", tt.expected.Status, data.Status)
			}
		})
	}
}

func TestWorkerResource_WorkerTypeDefaultBehavior(t *testing.T) {
	resource := &WorkerResource{}
	
	tests := []struct {
		name         string
		inputType    types.String
		expectedType string
	}{
		{
			name:         "explicit queue type",
			inputType:    types.StringValue("queue"),
			expectedType: "queue",
		},
		{
			name:         "scheduler type",
			inputType:    types.StringValue("scheduler"),
			expectedType: "scheduler",
		},
		{
			name:         "custom type",
			inputType:    types.StringValue("custom"),
			expectedType: "custom",
		},
		{
			name:         "null type should pass empty string",
			inputType:    types.StringNull(),
			expectedType: "",
		},
		{
			name:         "empty string type",
			inputType:    types.StringValue(""),
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &WorkerResourceModel{
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("test-worker"),
				Command:       types.StringValue("php artisan queue:work"),
				Type:          tt.inputType,
				Replicas:      types.Int64Value(1),
			}
			
			result := resource.toAPIModel(data)
			
			if result.Type != tt.expectedType {
				t.Errorf("Expected Type '%s', got '%s'", tt.expectedType, result.Type)
			}
		})
	}
}

func TestWorkerResource_ResourceAllocationValidation(t *testing.T) {
	resource := &WorkerResource{}
	
	tests := []struct {
		name          string
		memoryRequest string
		cpuRequest    string
		shouldPass    bool
	}{
		{
			name:          "valid memory and cpu requests",
			memoryRequest: "512Mi",
			cpuRequest:    "250m",
			shouldPass:    true,
		},
		{
			name:          "memory in Gi format",
			memoryRequest: "1Gi",
			cpuRequest:    "500m",
			shouldPass:    true,
		},
		{
			name:          "cpu in full cores",
			memoryRequest: "256Mi",
			cpuRequest:    "1",
			shouldPass:    true,
		},
		{
			name:          "empty resource requests",
			memoryRequest: "",
			cpuRequest:    "",
			shouldPass:    true,
		},
		{
			name:          "minimal resources",
			memoryRequest: "64Mi",
			cpuRequest:    "50m",
			shouldPass:    true,
		},
		{
			name:          "high resources",
			memoryRequest: "4Gi",
			cpuRequest:    "2",
			shouldPass:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &WorkerResourceModel{
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("resource-test-worker"),
				Command:       types.StringValue("php artisan queue:work"),
				Type:          types.StringValue("queue"),
				Replicas:      types.Int64Value(1),
				MemoryRequest: types.StringValue(tt.memoryRequest),
				CPURequest:    types.StringValue(tt.cpuRequest),
			}
			
			result := resource.toAPIModel(data)
			
			if result.MemoryRequest != tt.memoryRequest {
				t.Errorf("Expected MemoryRequest '%s', got '%s'", tt.memoryRequest, result.MemoryRequest)
			}
			if result.CPURequest != tt.cpuRequest {
				t.Errorf("Expected CPURequest '%s', got '%s'", tt.cpuRequest, result.CPURequest)
			}
		})
	}
}

func TestWorkerResource_BackwardCompatibility(t *testing.T) {
	resource := &WorkerResource{}
	
	// Test that existing worker configurations without new fields still work
	data := &WorkerResourceModel{
		ID:            types.Int64Value(1),
		ApplicationID: types.Int64Value(100),
		Name:          types.StringValue("legacy-worker"),
		Command:       types.StringValue("php artisan queue:work"),
		Replicas:      types.Int64Value(1),
		// New fields are null/unset
		Type:          types.StringNull(),
		MemoryRequest: types.StringNull(),
		CPURequest:    types.StringNull(),
	}
	
	result := resource.toAPIModel(data)
	
	// Verify basic fields are preserved
	if result.ID != 1 {
		t.Errorf("Expected ID 1, got %d", result.ID)
	}
	if result.ApplicationID != 100 {
		t.Errorf("Expected ApplicationID 100, got %d", result.ApplicationID)
	}
	if result.Name != "legacy-worker" {
		t.Errorf("Expected Name 'legacy-worker', got %s", result.Name)
	}
	if result.Command != "php artisan queue:work" {
		t.Errorf("Expected Command 'php artisan queue:work', got %s", result.Command)
	}
	if result.Replicas != 1 {
		t.Errorf("Expected Replicas 1, got %d", result.Replicas)
	}
	
	// Verify new fields have default/empty values
	if result.Type != "" {
		t.Errorf("Expected Type to be empty, got %s", result.Type)
	}
	if result.MemoryRequest != "" {
		t.Errorf("Expected MemoryRequest to be empty, got %s", result.MemoryRequest)
	}
	if result.CPURequest != "" {
		t.Errorf("Expected CPURequest to be empty, got %s", result.CPURequest)
	}
}

func TestWorkerResource_DefaultFieldValues(t *testing.T) {
	resource := &WorkerResource{}
	
	// Test the expected default values from schema
	tests := []struct {
		name         string
		field        string
		expectedType string
		expectedInt  int64
	}{
		{
			name:         "default worker type",
			field:        "type",
			expectedType: "queue",
		},
		{
			name:        "default replicas",
			field:       "replicas",
			expectedInt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test would normally verify schema defaults
			// For now, we verify the expected behavior in toAPIModel
			data := &WorkerResourceModel{
				ApplicationID: types.Int64Value(100),
				Name:          types.StringValue("default-test-worker"),
				Command:       types.StringValue("php artisan queue:work"),
			}
			
			switch tt.field {
			case "type":
				data.Type = types.StringValue(tt.expectedType)
				result := resource.toAPIModel(data)
				if result.Type != tt.expectedType {
					t.Errorf("Expected default Type '%s', got '%s'", tt.expectedType, result.Type)
				}
			case "replicas":
				data.Replicas = types.Int64Value(tt.expectedInt)
				result := resource.toAPIModel(data)
				if result.Replicas != tt.expectedInt {
					t.Errorf("Expected default Replicas %d, got %d", tt.expectedInt, result.Replicas)
				}
			}
		})
	}
}

func TestWorkerResource_APIClientIntegration(t *testing.T) {
	// Mock server for testing API interactions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.URL.Path {
		case "/applications/100/workers":
			if r.Method == http.MethodPost {
				// Return a worker with enhanced fields
				response := `{
					"success": true,
					"data": {
						"id": 1,
						"application_id": 100,
						"name": "test-worker",
						"command": "php artisan queue:work",
						"type": "queue",
						"replicas": 2,
						"memory_request": "512Mi",
						"cpu_request": "250m",
						"status": "running"
					}
				}`
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(response))
			}
		case "/applications/100/workers/1":
			if r.Method == http.MethodGet {
				response := `{
					"success": true,
					"data": {
						"id": 1,
						"application_id": 100,
						"name": "test-worker",
						"command": "php artisan queue:work",
						"type": "queue",
						"replicas": 2,
						"memory_request": "512Mi",
						"cpu_request": "250m",
						"status": "running"
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
	
	// Test worker creation with new fields
	worker := &client.Worker{
		ApplicationID: 100,
		Name:          "test-worker",
		Command:       "php artisan queue:work",
		Type:          "queue",
		Replicas:      2,
		MemoryRequest: "512Mi",
		CPURequest:    "250m",
	}
	
	created, err := c.CreateWorker(worker)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Verify response includes new fields
	if created.Type != "queue" {
		t.Errorf("Expected Type 'queue', got '%s'", created.Type)
	}
	if created.MemoryRequest != "512Mi" {
		t.Errorf("Expected MemoryRequest '512Mi', got '%s'", created.MemoryRequest)
	}
	if created.CPURequest != "250m" {
		t.Errorf("Expected CPURequest '250m', got '%s'", created.CPURequest)
	}
}

// Mock client for testing without network calls
type MockWorkerClient struct {
	workers map[int64]*client.Worker
	nextID  int64
}

func NewMockWorkerClient() *MockWorkerClient {
	return &MockWorkerClient{
		workers: make(map[int64]*client.Worker),
		nextID:  1,
	}
}

func (m *MockWorkerClient) CreateWorker(worker *client.Worker) (*client.Worker, error) {
	worker.ID = m.nextID
	worker.Status = "creating"
	worker.CreatedAt = time.Now()
	worker.UpdatedAt = time.Now()
	
	m.workers[worker.ID] = worker
	m.nextID++
	
	return worker, nil
}

func (m *MockWorkerClient) GetWorker(appID, workerID int64) (*client.Worker, error) {
	worker, exists := m.workers[workerID]
	if !exists {
		return nil, fmt.Errorf("worker not found")
	}
	return worker, nil
}

func TestWorkerResource_CRUDOperations(t *testing.T) {
	mockClient := NewMockWorkerClient()
	resource := &WorkerResource{client: nil} // We'll mock the client methods
	
	// Test Create
	data := &WorkerResourceModel{
		ApplicationID: types.Int64Value(100),
		Name:          types.StringValue("test-worker"),
		Command:       types.StringValue("php artisan queue:work"),
		Type:          types.StringValue("queue"),
		Replicas:      types.Int64Value(2),
		MemoryRequest: types.StringValue("512Mi"),
		CPURequest:    types.StringValue("250m"),
	}
	
	apiModel := resource.toAPIModel(data)
	created, err := mockClient.CreateWorker(apiModel)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Verify creation
	if created.ID == 0 {
		t.Error("Expected worker to have an ID after creation")
	}
	if created.Status != "creating" {
		t.Errorf("Expected status 'creating', got '%s'", created.Status)
	}
	
	// Test Read
	retrieved, err := mockClient.GetWorker(100, created.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve worker: %v", err)
	}
	
	if retrieved.Type != "queue" {
		t.Errorf("Expected Type 'queue', got '%s'", retrieved.Type)
	}
	if retrieved.MemoryRequest != "512Mi" {
		t.Errorf("Expected MemoryRequest '512Mi', got '%s'", retrieved.MemoryRequest)
	}
	if retrieved.CPURequest != "250m" {
		t.Errorf("Expected CPURequest '250m', got '%s'", retrieved.CPURequest)
	}
	if retrieved.Replicas != 2 {
		t.Errorf("Expected Replicas 2, got %d", retrieved.Replicas)
	}
}