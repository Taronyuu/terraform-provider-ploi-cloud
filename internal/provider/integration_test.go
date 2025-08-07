package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

// TestIntegration_AllEnhancedFeatures tests all enhanced features working together
func TestIntegration_AllEnhancedFeatures(t *testing.T) {
	// Mock server that handles all resource types with enhanced fields
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch {
		case r.URL.Path == "/applications" && r.Method == http.MethodPost:
			response := `{
				"success": true,
				"data": {
					"id": 1,
					"name": "integration-app",
					"application_type": "laravel",
					"start_command": "php artisan octane:start --host=0.0.0.0",
					"status": "running",
					"url": "https://integration-app.ploi.cloud",
					"needs_deployment": false,
					"memory_request": "1Gi",
					"replicas": 2
				}
			}`
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))
			
		case r.URL.Path == "/applications/1/services" && r.Method == http.MethodPost:
			response := `{
				"success": true,
				"data": {
					"id": 1,
					"application_id": 1,
					"name": "postgres-db",
					"type": "postgresql",
					"version": "15",
					"status": "running",
					"memory_request": "512Mi",
					"storage_size": "10Gi",
					"extensions": ["uuid-ossp", "pgcrypto"]
				}
			}`
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))
			
		case r.URL.Path == "/applications/1/workers" && r.Method == http.MethodPost:
			response := `{
				"success": true,
				"data": {
					"id": 1,
					"application_id": 1,
					"name": "queue-worker",
					"command": "php artisan queue:work",
					"type": "queue",
					"replicas": 2,
					"memory_request": "256Mi",
					"cpu_request": "250m",
					"status": "running"
				}
			}`
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))
			
		case r.URL.Path == "/applications/1/volumes" && r.Method == http.MethodPost:
			response := `{
				"success": true,
				"data": {
					"id": 1,
					"application_id": 1,
					"name": "data-volume",
					"size": 20,
					"path": "/var/lib/data",
					"storage_class": "fast-ssd",
					"resize_status": "completed"
				}
			}`
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))
			
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client with test server
	c := client.NewClient("test-token", &server.URL)
	
	// Test 1: Create application with start_command
	app := &client.Application{
		Name:         "integration-app",
		Type:         "laravel",
		StartCommand: "php artisan octane:start --host=0.0.0.0",
	}
	
	createdApp, err := c.CreateApplication(app)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}
	
	if createdApp.StartCommand != "php artisan octane:start --host=0.0.0.0" {
		t.Errorf("Expected StartCommand 'php artisan octane:start --host=0.0.0.0', got '%s'", createdApp.StartCommand)
	}
	
	// Test 2: Create service with memory_request, storage_size, and extensions
	service := &client.ApplicationService{
		ApplicationID: createdApp.ID,
		Name:          "postgres-db",
		Type:          "postgresql",
		Version:       "15",
		MemoryRequest: "512Mi",
		StorageSize:   "10Gi",
		Extensions:    []string{"uuid-ossp", "pgcrypto"},
	}
	
	createdService, err := c.CreateService(service)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	
	if createdService.MemoryRequest != "512Mi" {
		t.Errorf("Expected service MemoryRequest '512Mi', got '%s'", createdService.MemoryRequest)
	}
	if createdService.StorageSize != "10Gi" {
		t.Errorf("Expected service StorageSize '10Gi', got '%s'", createdService.StorageSize)
	}
	if len(createdService.Extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(createdService.Extensions))
	}
	
	// Test 3: Create worker with type, memory_request, and cpu_request
	worker := &client.Worker{
		ApplicationID: createdApp.ID,
		Name:          "queue-worker",
		Command:       "php artisan queue:work",
		Type:          "queue",
		Replicas:      2,
		MemoryRequest: "256Mi",
		CPURequest:    "250m",
	}
	
	createdWorker, err := c.CreateWorker(worker)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	if createdWorker.Type != "queue" {
		t.Errorf("Expected worker Type 'queue', got '%s'", createdWorker.Type)
	}
	if createdWorker.MemoryRequest != "256Mi" {
		t.Errorf("Expected worker MemoryRequest '256Mi', got '%s'", createdWorker.MemoryRequest)
	}
	if createdWorker.CPURequest != "250m" {
		t.Errorf("Expected worker CPURequest '250m', got '%s'", createdWorker.CPURequest)
	}
	
	// Test 4: Create volume with storage_class
	volume := &client.ApplicationVolume{
		ApplicationID: createdApp.ID,
		Name:          "data-volume",
		Size:          20,
		MountPath:     "/var/lib/data",
		StorageClass:  "fast-ssd",
	}
	
	createdVolume, err := c.CreateVolume(volume)
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}
	
	if createdVolume.StorageClass != "fast-ssd" {
		t.Errorf("Expected volume StorageClass 'fast-ssd', got '%s'", createdVolume.StorageClass)
	}
}

// TestIntegration_ResourceConversionConsistency tests that all resources handle conversion consistently
func TestIntegration_ResourceConversionConsistency(t *testing.T) {
	// Test application resource conversion
	appResource := &ApplicationResource{}
	appData := &ApplicationResourceModel{
		Name:         types.StringValue("test-app"),
		Type:         types.StringValue("laravel"),
		StartCommand: types.StringValue("php artisan serve"),
	}
	
	appAPIModel := appResource.toAPIModel(appData)
	if appAPIModel.StartCommand != "php artisan serve" {
		t.Errorf("Application conversion failed: expected 'php artisan serve', got '%s'", appAPIModel.StartCommand)
	}
	
	// Test service resource conversion
	serviceResource := &ServiceResource{}
	serviceData := &ServiceResourceModel{
		ApplicationID: types.Int64Value(1),
		Type:          types.StringValue("postgresql"),
		MemoryRequest: types.StringValue("1Gi"),
		StorageSize:   types.StringValue("10Gi"),
		Extensions:    types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("uuid-ossp"),
		}),
	}
	
	serviceAPIModel := serviceResource.toAPIModel(serviceData)
	if serviceAPIModel.MemoryRequest != "1Gi" {
		t.Errorf("Service conversion failed: expected '1Gi', got '%s'", serviceAPIModel.MemoryRequest)
	}
	if serviceAPIModel.StorageSize != "10Gi" {
		t.Errorf("Service conversion failed: expected '10Gi', got '%s'", serviceAPIModel.StorageSize)
	}
	if len(serviceAPIModel.Extensions) != 1 {
		t.Errorf("Service conversion failed: expected 1 extension, got %d", len(serviceAPIModel.Extensions))
	}
	
	// Test worker resource conversion
	workerResource := &WorkerResource{}
	workerData := &WorkerResourceModel{
		ApplicationID: types.Int64Value(1),
		Name:          types.StringValue("worker"),
		Command:       types.StringValue("php artisan queue:work"),
		Type:          types.StringValue("queue"),
		MemoryRequest: types.StringValue("512Mi"),
		CPURequest:    types.StringValue("250m"),
	}
	
	workerAPIModel := workerResource.toAPIModel(workerData)
	if workerAPIModel.Type != "queue" {
		t.Errorf("Worker conversion failed: expected 'queue', got '%s'", workerAPIModel.Type)
	}
	if workerAPIModel.MemoryRequest != "512Mi" {
		t.Errorf("Worker conversion failed: expected '512Mi', got '%s'", workerAPIModel.MemoryRequest)
	}
	if workerAPIModel.CPURequest != "250m" {
		t.Errorf("Worker conversion failed: expected '250m', got '%s'", workerAPIModel.CPURequest)
	}
	
	// Test volume resource conversion
	volumeResource := &VolumeResource{}
	volumeData := &VolumeResourceModel{
		ApplicationID: types.Int64Value(1),
		Name:          types.StringValue("volume"),
		Size:          types.Int64Value(10),
		MountPath:     types.StringValue("/data"),
		StorageClass:  types.StringValue("ssd"),
	}
	
	volumeAPIModel := volumeResource.toAPIModel(volumeData)
	if volumeAPIModel.StorageClass != "ssd" {
		t.Errorf("Volume conversion failed: expected 'ssd', got '%s'", volumeAPIModel.StorageClass)
	}
}

// TestIntegration_BackwardCompatibilityAcrossResources tests that all resources maintain backward compatibility
func TestIntegration_BackwardCompatibilityAcrossResources(t *testing.T) {
	// Test that all resources handle missing new fields gracefully
	
	// Application without start_command
	appResource := &ApplicationResource{}
	appData := &ApplicationResourceModel{
		Name:         types.StringValue("legacy-app"),
		Type:         types.StringValue("laravel"),
		StartCommand: types.StringNull(), // Legacy apps won't have this
	}
	
	appAPIModel := appResource.toAPIModel(appData)
	if appAPIModel.StartCommand != "" {
		t.Errorf("Expected empty StartCommand for backward compatibility, got '%s'", appAPIModel.StartCommand)
	}
	
	// Service without new fields
	serviceResource := &ServiceResource{}
	serviceData := &ServiceResourceModel{
		ApplicationID: types.Int64Value(1),
		Type:          types.StringValue("mysql"),
		MemoryRequest: types.StringNull(),
		StorageSize:   types.StringNull(),
		Extensions:    types.ListNull(types.StringType),
	}
	
	serviceAPIModel := serviceResource.toAPIModel(serviceData)
	if serviceAPIModel.MemoryRequest != "" {
		t.Errorf("Expected empty MemoryRequest for backward compatibility, got '%s'", serviceAPIModel.MemoryRequest)
	}
	if serviceAPIModel.StorageSize != "" {
		t.Errorf("Expected empty StorageSize for backward compatibility, got '%s'", serviceAPIModel.StorageSize)
	}
	if serviceAPIModel.Extensions != nil {
		t.Errorf("Expected nil Extensions for backward compatibility, got %v", serviceAPIModel.Extensions)
	}
	
	// Worker without new fields
	workerResource := &WorkerResource{}
	workerData := &WorkerResourceModel{
		ApplicationID: types.Int64Value(1),
		Name:          types.StringValue("legacy-worker"),
		Command:       types.StringValue("php artisan queue:work"),
		Type:          types.StringNull(),
		MemoryRequest: types.StringNull(),
		CPURequest:    types.StringNull(),
	}
	
	workerAPIModel := workerResource.toAPIModel(workerData)
	if workerAPIModel.Type != "" {
		t.Errorf("Expected empty Type for backward compatibility, got '%s'", workerAPIModel.Type)
	}
	if workerAPIModel.MemoryRequest != "" {
		t.Errorf("Expected empty MemoryRequest for backward compatibility, got '%s'", workerAPIModel.MemoryRequest)
	}
	if workerAPIModel.CPURequest != "" {
		t.Errorf("Expected empty CPURequest for backward compatibility, got '%s'", workerAPIModel.CPURequest)
	}
	
	// Volume without storage_class
	volumeResource := &VolumeResource{}
	volumeData := &VolumeResourceModel{
		ApplicationID: types.Int64Value(1),
		Name:          types.StringValue("legacy-volume"),
		Size:          types.Int64Value(10),
		MountPath:     types.StringValue("/data"),
		StorageClass:  types.StringNull(),
	}
	
	volumeAPIModel := volumeResource.toAPIModel(volumeData)
	if volumeAPIModel.StorageClass != "" {
		t.Errorf("Expected empty StorageClass for backward compatibility, got '%s'", volumeAPIModel.StorageClass)
	}
}

// TestIntegration_ComplexScenario tests a complex scenario with multiple resources using enhanced features
func TestIntegration_ComplexScenario(t *testing.T) {
	// Create mock resources and test complex interactions
	
	// Scenario: Laravel app with Octane, PostgreSQL with extensions, Redis cache, multiple workers
	appResource := &ApplicationResource{}
	appData := &ApplicationResourceModel{
		Name:         types.StringValue("laravel-octane-app"),
		Type:         types.StringValue("laravel"),
		StartCommand: types.StringValue("php artisan octane:start --server=swoole --host=0.0.0.0 --port=8000"),
		Settings: &SettingsModel{
			MemoryRequest: types.StringValue("2Gi"),
			Replicas:      types.Int64Value(3),
		},
	}
	
	appAPIModel := appResource.toAPIModel(appData)
	
	// PostgreSQL service with extensions
	serviceResource := &ServiceResource{}
	pgData := &ServiceResourceModel{
		ApplicationID: types.Int64Value(1),
		Name:          types.StringValue("postgres"),
		Type:          types.StringValue("postgresql"),
		Version:       types.StringValue("15"),
		MemoryRequest: types.StringValue("1Gi"),
		StorageSize:   types.StringValue("20Gi"),
		Extensions:    types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("uuid-ossp"),
			types.StringValue("pgcrypto"),
			types.StringValue("hstore"),
		}),
	}
	
	pgAPIModel := serviceResource.toAPIModel(pgData)
	
	// Redis service for caching
	redisData := &ServiceResourceModel{
		ApplicationID: types.Int64Value(1),
		Name:          types.StringValue("redis"),
		Type:          types.StringValue("redis"),
		MemoryRequest: types.StringValue("512Mi"),
		StorageSize:   types.StringValue("1Gi"),
	}
	
	redisAPIModel := serviceResource.toAPIModel(redisData)
	
	// Queue worker
	workerResource := &WorkerResource{}
	queueWorkerData := &WorkerResourceModel{
		ApplicationID: types.Int64Value(1),
		Name:          types.StringValue("queue-worker"),
		Command:       types.StringValue("php artisan queue:work --timeout=300"),
		Type:          types.StringValue("queue"),
		Replicas:      types.Int64Value(2),
		MemoryRequest: types.StringValue("512Mi"),
		CPURequest:    types.StringValue("250m"),
	}
	
	queueWorkerAPIModel := workerResource.toAPIModel(queueWorkerData)
	
	// Scheduler worker
	schedulerWorkerData := &WorkerResourceModel{
		ApplicationID: types.Int64Value(1),
		Name:          types.StringValue("scheduler"),
		Command:       types.StringValue("php artisan schedule:work"),
		Type:          types.StringValue("scheduler"),
		Replicas:      types.Int64Value(1),
		MemoryRequest: types.StringValue("256Mi"),
		CPURequest:    types.StringValue("100m"),
	}
	
	schedulerWorkerAPIModel := workerResource.toAPIModel(schedulerWorkerData)
	
	// Storage volume for uploads
	volumeResource := &VolumeResource{}
	volumeData := &VolumeResourceModel{
		ApplicationID: types.Int64Value(1),
		Name:          types.StringValue("uploads"),
		Size:          types.Int64Value(100),
		MountPath:     types.StringValue("/var/www/storage/app/uploads"),
		StorageClass:  types.StringValue("fast-ssd"),
	}
	
	volumeAPIModel := volumeResource.toAPIModel(volumeData)
	
	// Verify all conversions work correctly together
	if appAPIModel.StartCommand != "php artisan octane:start --server=swoole --host=0.0.0.0 --port=8000" {
		t.Error("Application StartCommand conversion failed")
	}
	
	if pgAPIModel.MemoryRequest != "1Gi" || pgAPIModel.StorageSize != "20Gi" || len(pgAPIModel.Extensions) != 3 {
		t.Error("PostgreSQL service conversion failed")
	}
	
	if redisAPIModel.MemoryRequest != "512Mi" || redisAPIModel.StorageSize != "1Gi" {
		t.Error("Redis service conversion failed")
	}
	
	if queueWorkerAPIModel.Type != "queue" || queueWorkerAPIModel.Replicas != 2 {
		t.Error("Queue worker conversion failed")
	}
	
	if schedulerWorkerAPIModel.Type != "scheduler" || schedulerWorkerAPIModel.Replicas != 1 {
		t.Error("Scheduler worker conversion failed")
	}
	
	if volumeAPIModel.StorageClass != "fast-ssd" || volumeAPIModel.Size != 100 {
		t.Error("Volume conversion failed")
	}
}

// TestIntegration_ErrorHandling tests error handling across all enhanced features
func TestIntegration_ErrorHandling(t *testing.T) {
	// Test that resources handle errors gracefully with new fields
	
	// Mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"message": "Validation failed",
			"errors": {
				"memory_request": ["Invalid memory format"],
				"storage_class": ["Unknown storage class"]
			}
		}`))
	}))
	defer server.Close()

	c := client.NewClient("test-token", &server.URL)
	
	// Test service creation with invalid enhanced fields
	service := &client.ApplicationService{
		ApplicationID: 1,
		Type:          "postgresql",
		MemoryRequest: "invalid-format",
		StorageSize:   "bad-size",
	}
	
	_, err := c.CreateService(service)
	if err == nil {
		t.Error("Expected error for invalid service fields, got nil")
	}
	
	// Test worker creation with invalid enhanced fields
	worker := &client.Worker{
		ApplicationID: 1,
		Name:          "test-worker",
		Command:       "php artisan queue:work",
		Type:          "invalid-type",
		MemoryRequest: "bad-memory",
	}
	
	_, err = c.CreateWorker(worker)
	if err == nil {
		t.Error("Expected error for invalid worker fields, got nil")
	}
	
	// Test volume creation with invalid storage class
	volume := &client.ApplicationVolume{
		ApplicationID: 1,
		Name:          "test-volume",
		Size:          10,
		MountPath:     "/data",
		StorageClass:  "invalid-class",
	}
	
	_, err = c.CreateVolume(volume)
	if err == nil {
		t.Error("Expected error for invalid volume fields, got nil")
	}
}

// Mock integration client for comprehensive testing
type MockIntegrationClient struct {
	applications map[int64]*client.Application
	services     map[int64]*client.ApplicationService
	workers      map[int64]*client.Worker
	volumes      map[int64]*client.ApplicationVolume
	nextID       int64
}

func NewMockIntegrationClient() *MockIntegrationClient {
	return &MockIntegrationClient{
		applications: make(map[int64]*client.Application),
		services:     make(map[int64]*client.ApplicationService),
		workers:      make(map[int64]*client.Worker),
		volumes:      make(map[int64]*client.ApplicationVolume),
		nextID:       1,
	}
}

func (m *MockIntegrationClient) CreateApplication(app *client.Application) (*client.Application, error) {
	app.ID = m.nextID
	app.Status = "running"
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()
	
	m.applications[app.ID] = app
	m.nextID++
	return app, nil
}

func TestIntegration_FullStackDeployment(t *testing.T) {
	mockClient := NewMockIntegrationClient()
	
	// Create full stack deployment with all enhanced features
	app := &client.Application{
		Name:         "full-stack-app",
		Type:         "laravel",
		StartCommand: "php artisan octane:start --host=0.0.0.0 --port=8000",
	}
	
	createdApp, err := mockClient.CreateApplication(app)
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}
	
	// Verify application was created with enhanced features
	if createdApp.StartCommand != "php artisan octane:start --host=0.0.0.0 --port=8000" {
		t.Errorf("Expected StartCommand to be preserved, got '%s'", createdApp.StartCommand)
	}
	
	if createdApp.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", createdApp.Status)
	}
}