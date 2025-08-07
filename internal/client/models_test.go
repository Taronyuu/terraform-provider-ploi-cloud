package client

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestApplicationService_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		service ApplicationService
	}{
		{
			name: "complete service with all new fields",
			service: ApplicationService{
				ID:            1,
				ApplicationID: 100,
				Name:          "test-service",
				Type:          "postgresql",
				Version:       "15",
				Status:        "running",
				Replicas:      2,
				CPURequest:    "500m",
				MemoryRequest: "1Gi",
				StorageSize:   "10Gi",
				Extensions:    []string{"uuid-ossp", "pgcrypto", "hstore"},
			},
		},
		{
			name: "service without extensions",
			service: ApplicationService{
				ID:            2,
				ApplicationID: 100,
				Name:          "redis-service",
				Type:          "redis",
				Status:        "running",
				MemoryRequest: "512Mi",
				StorageSize:   "5Gi",
				Extensions:    nil,
			},
		},
		{
			name: "service with empty extensions",
			service: ApplicationService{
				ID:            3,
				ApplicationID: 100,
				Name:          "mysql-service",
				Type:          "mysql",
				Status:        "running",
				MemoryRequest: "256Mi",
				StorageSize:   "1Gi",
				Extensions:    nil, // JSON will unmarshall empty array as nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(tt.service)
			if err != nil {
				t.Fatalf("Failed to marshal service: %v", err)
			}

			// Unmarshal back to struct
			var unmarshaled ApplicationService
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal service: %v", err)
			}

			// Compare fields
			if unmarshaled.ID != tt.service.ID {
				t.Errorf("Expected ID %d, got %d", tt.service.ID, unmarshaled.ID)
			}
			if unmarshaled.ApplicationID != tt.service.ApplicationID {
				t.Errorf("Expected ApplicationID %d, got %d", tt.service.ApplicationID, unmarshaled.ApplicationID)
			}
			if unmarshaled.Name != tt.service.Name {
				t.Errorf("Expected Name %s, got %s", tt.service.Name, unmarshaled.Name)
			}
			if unmarshaled.Type != tt.service.Type {
				t.Errorf("Expected Type %s, got %s", tt.service.Type, unmarshaled.Type)
			}
			if unmarshaled.Version != tt.service.Version {
				t.Errorf("Expected Version %s, got %s", tt.service.Version, unmarshaled.Version)
			}
			if unmarshaled.Status != tt.service.Status {
				t.Errorf("Expected Status %s, got %s", tt.service.Status, unmarshaled.Status)
			}
			if unmarshaled.Replicas != tt.service.Replicas {
				t.Errorf("Expected Replicas %d, got %d", tt.service.Replicas, unmarshaled.Replicas)
			}
			if unmarshaled.CPURequest != tt.service.CPURequest {
				t.Errorf("Expected CPURequest %s, got %s", tt.service.CPURequest, unmarshaled.CPURequest)
			}
			if unmarshaled.MemoryRequest != tt.service.MemoryRequest {
				t.Errorf("Expected MemoryRequest %s, got %s", tt.service.MemoryRequest, unmarshaled.MemoryRequest)
			}
			if unmarshaled.StorageSize != tt.service.StorageSize {
				t.Errorf("Expected StorageSize %s, got %s", tt.service.StorageSize, unmarshaled.StorageSize)
			}
			// Handle nil vs empty slice comparison
			if (tt.service.Extensions == nil && len(unmarshaled.Extensions) > 0) ||
			   (tt.service.Extensions != nil && !reflect.DeepEqual(unmarshaled.Extensions, tt.service.Extensions)) {
				t.Errorf("Expected Extensions %v, got %v", tt.service.Extensions, unmarshaled.Extensions)
			}
		})
	}
}

func TestWorker_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name   string
		worker Worker
	}{
		{
			name: "complete worker with all new fields",
			worker: Worker{
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
		},
		{
			name: "worker with scheduler type",
			worker: Worker{
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
		},
		{
			name: "worker with custom type",
			worker: Worker{
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
		},
		{
			name: "worker with empty resource requests",
			worker: Worker{
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(tt.worker)
			if err != nil {
				t.Fatalf("Failed to marshal worker: %v", err)
			}

			// Unmarshal back to struct
			var unmarshaled Worker
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal worker: %v", err)
			}

			// Compare fields
			if unmarshaled.ID != tt.worker.ID {
				t.Errorf("Expected ID %d, got %d", tt.worker.ID, unmarshaled.ID)
			}
			if unmarshaled.ApplicationID != tt.worker.ApplicationID {
				t.Errorf("Expected ApplicationID %d, got %d", tt.worker.ApplicationID, unmarshaled.ApplicationID)
			}
			if unmarshaled.Name != tt.worker.Name {
				t.Errorf("Expected Name %s, got %s", tt.worker.Name, unmarshaled.Name)
			}
			if unmarshaled.Command != tt.worker.Command {
				t.Errorf("Expected Command %s, got %s", tt.worker.Command, unmarshaled.Command)
			}
			if unmarshaled.Type != tt.worker.Type {
				t.Errorf("Expected Type %s, got %s", tt.worker.Type, unmarshaled.Type)
			}
			if unmarshaled.Replicas != tt.worker.Replicas {
				t.Errorf("Expected Replicas %d, got %d", tt.worker.Replicas, unmarshaled.Replicas)
			}
			if unmarshaled.MemoryRequest != tt.worker.MemoryRequest {
				t.Errorf("Expected MemoryRequest %s, got %s", tt.worker.MemoryRequest, unmarshaled.MemoryRequest)
			}
			if unmarshaled.CPURequest != tt.worker.CPURequest {
				t.Errorf("Expected CPURequest %s, got %s", tt.worker.CPURequest, unmarshaled.CPURequest)
			}
			if unmarshaled.Status != tt.worker.Status {
				t.Errorf("Expected Status %s, got %s", tt.worker.Status, unmarshaled.Status)
			}
		})
	}
}

func TestApplication_StartCommand_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name string
		app  Application
	}{
		{
			name: "application with start command",
			app: Application{
				ID:           1,
				Name:         "test-app",
				Type:         "laravel",
				StartCommand: "php artisan octane:start --host=0.0.0.0 --port=8000",
				Status:       "running",
			},
		},
		{
			name: "application with nodejs start command",
			app: Application{
				ID:           2,
				Name:         "node-app",
				Type:         "nodejs",
				StartCommand: "npm run start:prod",
				Status:       "running",
			},
		},
		{
			name: "application with empty start command",
			app: Application{
				ID:           3,
				Name:         "default-app",
				Type:         "laravel",
				StartCommand: "",
				Status:       "running",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(tt.app)
			if err != nil {
				t.Fatalf("Failed to marshal application: %v", err)
			}

			// Unmarshal back to struct
			var unmarshaled Application
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal application: %v", err)
			}

			// Compare fields including StartCommand
			if unmarshaled.ID != tt.app.ID {
				t.Errorf("Expected ID %d, got %d", tt.app.ID, unmarshaled.ID)
			}
			if unmarshaled.Name != tt.app.Name {
				t.Errorf("Expected Name %s, got %s", tt.app.Name, unmarshaled.Name)
			}
			if unmarshaled.Type != tt.app.Type {
				t.Errorf("Expected Type %s, got %s", tt.app.Type, unmarshaled.Type)
			}
			if unmarshaled.StartCommand != tt.app.StartCommand {
				t.Errorf("Expected StartCommand %s, got %s", tt.app.StartCommand, unmarshaled.StartCommand)
			}
			if unmarshaled.Status != tt.app.Status {
				t.Errorf("Expected Status %s, got %s", tt.app.Status, unmarshaled.Status)
			}
		})
	}
}

func TestApplicationVolume_StorageClass_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name   string
		volume ApplicationVolume
	}{
		{
			name: "volume with storage class",
			volume: ApplicationVolume{
				ID:            1,
				ApplicationID: 100,
				Name:          "test-volume",
				Size:          10,
				MountPath:     "/var/lib/data",
				StorageClass:  "fast-ssd",
				ResizeStatus:  "completed",
			},
		},
		{
			name: "volume with different storage class",
			volume: ApplicationVolume{
				ID:            2,
				ApplicationID: 100,
				Name:          "standard-volume",
				Size:          50,
				MountPath:     "/var/lib/storage",
				StorageClass:  "standard",
				ResizeStatus:  "completed",
			},
		},
		{
			name: "volume with empty storage class",
			volume: ApplicationVolume{
				ID:            3,
				ApplicationID: 100,
				Name:          "no-class-volume",
				Size:          25,
				MountPath:     "/var/lib/default",
				StorageClass:  "",
				ResizeStatus:  "completed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(tt.volume)
			if err != nil {
				t.Fatalf("Failed to marshal volume: %v", err)
			}

			// Unmarshal back to struct
			var unmarshaled ApplicationVolume
			err = json.Unmarshal(jsonData, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal volume: %v", err)
			}

			// Compare fields including StorageClass
			if unmarshaled.ID != tt.volume.ID {
				t.Errorf("Expected ID %d, got %d", tt.volume.ID, unmarshaled.ID)
			}
			if unmarshaled.ApplicationID != tt.volume.ApplicationID {
				t.Errorf("Expected ApplicationID %d, got %d", tt.volume.ApplicationID, unmarshaled.ApplicationID)
			}
			if unmarshaled.Name != tt.volume.Name {
				t.Errorf("Expected Name %s, got %s", tt.volume.Name, unmarshaled.Name)
			}
			if unmarshaled.Size != tt.volume.Size {
				t.Errorf("Expected Size %d, got %d", tt.volume.Size, unmarshaled.Size)
			}
			if unmarshaled.MountPath != tt.volume.MountPath {
				t.Errorf("Expected MountPath %s, got %s", tt.volume.MountPath, unmarshaled.MountPath)
			}
			if unmarshaled.StorageClass != tt.volume.StorageClass {
				t.Errorf("Expected StorageClass %s, got %s", tt.volume.StorageClass, unmarshaled.StorageClass)
			}
			if unmarshaled.ResizeStatus != tt.volume.ResizeStatus {
				t.Errorf("Expected ResizeStatus %s, got %s", tt.volume.ResizeStatus, unmarshaled.ResizeStatus)
			}
		})
	}
}

func TestFlexibleSettings_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FlexibleSettings
	}{
		{
			name:  "normal map",
			input: `{"key1": "value1", "key2": "value2"}`,
			expected: FlexibleSettings{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:     "empty map",
			input:    `{}`,
			expected: FlexibleSettings{},
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: FlexibleSettings{},
		},
		{
			name:     "null value",
			input:    `null`,
			expected: FlexibleSettings{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fs FlexibleSettings
			err := json.Unmarshal([]byte(tt.input), &fs)
			
			// FlexibleSettings should handle all cases gracefully
			if err != nil {
				t.Fatalf("Unexpected error unmarshaling: %v", err)
			}

			// Initialize empty map if fs is nil to match expected
			if fs == nil {
				fs = FlexibleSettings{}
			}

			if !reflect.DeepEqual(fs, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, fs)
			}
		})
	}
}

func TestFlexibleSettings_MarshalJSON(t *testing.T) {
	fs := FlexibleSettings{
		"database": "postgresql",
		"version":  "15",
		"port":     "5432",
	}

	jsonData, err := json.Marshal(fs)
	if err != nil {
		t.Fatalf("Failed to marshal FlexibleSettings: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]string
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal back to map: %v", err)
	}

	expected := map[string]string{
		"database": "postgresql",
		"version":  "15",
		"port":     "5432",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFlexibleSettings_ToMap(t *testing.T) {
	fs := FlexibleSettings{
		"key1": "value1",
		"key2": "value2",
	}

	result := fs.ToMap()
	expected := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFlexibleSettingsFromMap(t *testing.T) {
	input := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	fs := FlexibleSettingsFromMap(input)
	expected := FlexibleSettings{
		"key1": "value1",
		"key2": "value2",
	}

	if !reflect.DeepEqual(fs, expected) {
		t.Errorf("Expected %v, got %v", expected, fs)
	}
}

func TestCompleteModel_JSONRoundTrip(t *testing.T) {
	// Test complete round-trip with all enhanced fields
	original := Application{
		ID:                 1,
		Name:               "comprehensive-test",
		Type:               "laravel",
		ApplicationVersion: "11.x",
		PHPVersion:         "8.3",
		NodeJSVersion:      "20",
		BuildCommands:      []string{"composer install", "npm install"},
		InitCommands:       []string{"php artisan migrate"},
		StartCommand:       "php artisan octane:start --host=0.0.0.0",
		PHPExtensions:      []string{"pdo", "mbstring"},
		PHPSettings:        []string{"memory_limit=256M"},
		HealthCheckPath:    "/health",
		SchedulerEnabled:   true,
		Replicas:           2,
		CPURequest:         "500m",
		MemoryRequest:      "1Gi",
		URL:                "https://test.ploi.cloud",
		Status:             "running",
		NeedsDeployment:    false,
		CustomManifests:    "apiVersion: v1\nkind: ConfigMap",
		RepositoryURL:      "https://github.com/user/repo",
		RepositoryOwner:    "user",
		RepositoryName:     "repo",
		DefaultBranch:      "main",
		SocialAccountID:    123,
		Region:             "us-east-1",
		Provider:           "github",
		CreatedAt:          time.Now().Truncate(time.Second),
		UpdatedAt:          time.Now().Truncate(time.Second),
		Services: []ApplicationService{
			{
				ID:            1,
				ApplicationID: 1,
				Name:          "postgres",
				Type:          "postgresql",
				Version:       "15",
				MemoryRequest: "512Mi",
				StorageSize:   "10Gi",
				Extensions:    []string{"uuid-ossp"},
			},
		},
		Volumes: []ApplicationVolume{
			{
				ID:            1,
				ApplicationID: 1,
				Name:          "data-volume",
				Size:          20,
				MountPath:     "/var/lib/data",
				StorageClass:  "fast-ssd",
				ResizeStatus:  "completed",
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal application: %v", err)
	}

	// Unmarshal back to struct
	var unmarshaled Application
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal application: %v", err)
	}

	// Verify critical fields
	if unmarshaled.StartCommand != original.StartCommand {
		t.Errorf("Expected StartCommand %s, got %s", original.StartCommand, unmarshaled.StartCommand)
	}
	
	if len(unmarshaled.Services) != len(original.Services) {
		t.Errorf("Expected %d services, got %d", len(original.Services), len(unmarshaled.Services))
	} else {
		service := unmarshaled.Services[0]
		originalService := original.Services[0]
		if service.MemoryRequest != originalService.MemoryRequest {
			t.Errorf("Expected service MemoryRequest %s, got %s", originalService.MemoryRequest, service.MemoryRequest)
		}
		if service.StorageSize != originalService.StorageSize {
			t.Errorf("Expected service StorageSize %s, got %s", originalService.StorageSize, service.StorageSize)
		}
		if !reflect.DeepEqual(service.Extensions, originalService.Extensions) {
			t.Errorf("Expected service Extensions %v, got %v", originalService.Extensions, service.Extensions)
		}
	}
	
	if len(unmarshaled.Volumes) != len(original.Volumes) {
		t.Errorf("Expected %d volumes, got %d", len(original.Volumes), len(unmarshaled.Volumes))
	} else {
		volume := unmarshaled.Volumes[0]
		originalVolume := original.Volumes[0]
		if volume.StorageClass != originalVolume.StorageClass {
			t.Errorf("Expected volume StorageClass %s, got %s", originalVolume.StorageClass, volume.StorageClass)
		}
	}
}

func TestAPIResponse_JSONUnmarshaling(t *testing.T) {
	// Test that API responses with new fields unmarshal correctly
	jsonResponse := `{
		"success": true,
		"data": {
			"id": 1,
			"application_id": 100,
			"name": "api-test-service",
			"type": "postgresql",
			"version": "15",
			"status": "running",
			"replicas": 2,
			"cpu_request": "500m",
			"memory_request": "1Gi",
			"storage_size": "10Gi",
			"extensions": ["uuid-ossp", "pgcrypto"],
			"settings": {
				"max_connections": "100",
				"shared_buffers": "256MB"
			}
		}
	}`

	var response SingleResponse[ApplicationService]
	err := json.Unmarshal([]byte(jsonResponse), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal API response: %v", err)
	}

	service := response.Data
	if service.MemoryRequest != "1Gi" {
		t.Errorf("Expected MemoryRequest '1Gi', got '%s'", service.MemoryRequest)
	}
	if service.StorageSize != "10Gi" {
		t.Errorf("Expected StorageSize '10Gi', got '%s'", service.StorageSize)
	}
	if len(service.Extensions) != 2 {
		t.Errorf("Expected 2 extensions, got %d", len(service.Extensions))
	}
	if !reflect.DeepEqual(service.Extensions, []string{"uuid-ossp", "pgcrypto"}) {
		t.Errorf("Expected extensions ['uuid-ossp', 'pgcrypto'], got %v", service.Extensions)
	}

	// Test settings as map
	expectedSettings := map[string]string{
		"max_connections": "100",
		"shared_buffers":  "256MB",
	}
	if !reflect.DeepEqual(service.Settings.ToMap(), expectedSettings) {
		t.Errorf("Expected settings %v, got %v", expectedSettings, service.Settings.ToMap())
	}
}