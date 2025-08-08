package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ploi/terraform-provider-ploicloud/internal/client"
)

// TestToUpdateAPIModel_ConsistencyFixes tests the toUpdateAPIModel function
// to ensure all necessary fields are included in updates per the consistency fixes
func TestToUpdateAPIModel_ConsistencyFixes(t *testing.T) {
	resource := &ApplicationResource{}

	tests := []struct {
		name           string
		data           *ApplicationResourceModel
		expectedFields map[string]interface{}
		description    string
	}{
		{
			name: "start_command included in updates",
			data: &ApplicationResourceModel{
				Name:         types.StringValue("test-app"),
				StartCommand: types.StringValue("npm run production"),
				Runtime: &RuntimeModel{
					NodeJSVersion: types.StringValue("20"),
				},
			},
			expectedFields: map[string]interface{}{
				"start_command":   "npm run production",
				"nodejs_version":  "20",
				"name":           "test-app",
			},
			description: "start_command should be included in update payload to fix consistency errors",
		},
		{
			name: "all runtime fields included",
			data: &ApplicationResourceModel{
				Runtime: &RuntimeModel{
					NodeJSVersion: types.StringValue("22"),
					PHPVersion:    types.StringValue("8.4"),
				},
			},
			expectedFields: map[string]interface{}{
				"nodejs_version": "22",
				"php_version":    "8.4",
			},
			description: "Both nodejs_version and php_version should be included when present",
		},
		{
			name: "all settings fields included",
			data: &ApplicationResourceModel{
				Settings: &SettingsModel{
					HealthCheckPath:  types.StringValue("/health"),
					SchedulerEnabled: types.BoolValue(true),
					Replicas:         types.Int64Value(3),
					CPURequest:       types.StringValue("500m"),
					MemoryRequest:    types.StringValue("1Gi"),
				},
			},
			expectedFields: map[string]interface{}{
				"health_check_path":  "/health",
				"scheduler_enabled":  true,
				"replicas":          int64(3),
				"cpu_request":       "500m",
				"memory_request":    "1Gi",
			},
			description: "All settings fields should be included in updates",
		},
		{
			name: "build and init commands included",
			data: &ApplicationResourceModel{
				BuildCommands: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("npm install"),
					types.StringValue("npm run build"),
				}),
				InitCommands: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("npm run migrate"),
				}),
			},
			expectedFields: map[string]interface{}{
				"build_commands": []string{"npm install", "npm run build"},
				"init_commands":  []string{"npm run migrate"},
			},
			description: "Build and init commands should be included in updates",
		},
		{
			name: "php configuration fields included",
			data: &ApplicationResourceModel{
				PHPExtensions: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("redis"),
					types.StringValue("pdo_mysql"),
				}),
				PHPSettings: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("memory_limit=256M"),
				}),
			},
			expectedFields: map[string]interface{}{
				"php_extensions": []string{"redis", "pdo_mysql"},
				"php_settings":   []string{"memory_limit=256M"},
			},
			description: "PHP extensions and settings should be included in updates",
		},
		{
			name: "additional domains included",
			data: &ApplicationResourceModel{
				AdditionalDomains: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("api.example.com"),
					types.StringValue("admin.example.com"),
				}),
			},
			expectedFields: map[string]interface{}{
				"additional_domains": []string{"api.example.com", "admin.example.com"},
			},
			description: "Additional domains should be included in updates",
		},
		{
			name: "null values excluded from updates",
			data: &ApplicationResourceModel{
				Name:         types.StringValue("test-app"),
				StartCommand: types.StringNull(),
				Runtime: &RuntimeModel{
					NodeJSVersion: types.StringNull(),
					PHPVersion:    types.StringNull(),
				},
				Settings: &SettingsModel{
					HealthCheckPath:  types.StringNull(),
					SchedulerEnabled: types.BoolNull(),
					Replicas:         types.Int64Null(),
					CPURequest:       types.StringNull(),
					MemoryRequest:    types.StringNull(),
				},
				BuildCommands:     types.ListNull(types.StringType),
				InitCommands:      types.ListNull(types.StringType),
				PHPExtensions:     types.ListNull(types.StringType),
				PHPSettings:       types.ListNull(types.StringType),
				AdditionalDomains: types.ListNull(types.StringType),
			},
			expectedFields: map[string]interface{}{
				"name": "test-app",
			},
			description: "Only non-null fields should be included in updates",
		},
		{
			name: "empty string values handled properly",
			data: &ApplicationResourceModel{
				StartCommand: types.StringValue(""),
				Runtime: &RuntimeModel{
					NodeJSVersion: types.StringValue(""),
					PHPVersion:    types.StringValue(""),
				},
			},
			expectedFields: map[string]interface{}{},
			description: "Empty strings should be excluded from updates",
		},
		{
			name: "comprehensive update payload",
			data: &ApplicationResourceModel{
				Name:         types.StringValue("full-app"),
				StartCommand: types.StringValue("php artisan octane:start"),
				Runtime: &RuntimeModel{
					NodeJSVersion: types.StringValue("20"),
					PHPVersion:    types.StringValue("8.4"),
				},
				Settings: &SettingsModel{
					HealthCheckPath:  types.StringValue("/status"),
					SchedulerEnabled: types.BoolValue(false),
					Replicas:         types.Int64Value(2),
					CPURequest:       types.StringValue("250m"),
					MemoryRequest:    types.StringValue("512Mi"),
				},
				BuildCommands: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("composer install"),
				}),
				InitCommands: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("php artisan migrate"),
				}),
				CustomManifests: types.StringValue("apiVersion: v1\nkind: ConfigMap"),
			},
			expectedFields: map[string]interface{}{
				"name":               "full-app",
				"start_command":      "php artisan octane:start",
				"nodejs_version":     "20",
				"php_version":        "8.4",
				"health_check_path":  "/status",
				"scheduler_enabled":  false,
				"replicas":          int64(2),
				"cpu_request":       "250m",
				"memory_request":    "512Mi",
				"build_commands":    []string{"composer install"},
				"init_commands":     []string{"php artisan migrate"},
				"custom_manifests":  "apiVersion: v1\nkind: ConfigMap",
			},
			description: "Comprehensive test with all field types included",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resource.toUpdateAPIModel(tt.data)

			for expectedKey, expectedValue := range tt.expectedFields {
				actualValue, exists := result[expectedKey]
				if !exists {
					t.Errorf("%s: Expected field '%s' to be included in update payload", 
						tt.description, expectedKey)
					continue
				}

				if !deepEqual(actualValue, expectedValue) {
					t.Errorf("%s: Expected '%s' = %v, got %v", 
						tt.description, expectedKey, expectedValue, actualValue)
				}
			}

			for resultKey := range result {
				if _, expected := tt.expectedFields[resultKey]; !expected {
					t.Errorf("%s: Unexpected field '%s' = %v in update payload",
						tt.description, resultKey, result[resultKey])
				}
			}
		})
	}
}

// TestFromAPIModel_ConsistencyFixes tests the fromAPIModel function
// to ensure API responses are properly mapped to state per the consistency fixes
func TestFromAPIModel_ConsistencyFixes(t *testing.T) {
	resource := &ApplicationResource{}

	tests := []struct {
		name        string
		app         *client.Application
		initialData *ApplicationResourceModel
		verify      func(t *testing.T, data *ApplicationResourceModel, description string)
		description string
	}{
		{
			name: "api response properly mapped to state",
			app: &client.Application{
				ID:              123,
				Name:            "api-app",
				Type:            "nodejs",
				StartCommand:    "npm run start:prod",
				NodeJSVersion:   "20",
				MemoryRequest:   "1Gi",
				HealthCheckPath: "/health",
				SchedulerEnabled: true,
				Replicas:        3,
				Status:          "running",
				URL:             "https://api-app.ploi.cloud",
				NeedsDeployment: false,
			},
			initialData: &ApplicationResourceModel{
				StartCommand: types.StringNull(),
				Runtime:      &RuntimeModel{},
				Settings:     &SettingsModel{},
			},
			verify: func(t *testing.T, data *ApplicationResourceModel, description string) {
				if !data.ID.Equal(types.Int64Value(123)) {
					t.Errorf("%s: Expected ID 123, got %v", description, data.ID)
				}
				if !data.StartCommand.Equal(types.StringValue("npm run start:prod")) {
					t.Errorf("%s: Expected StartCommand 'npm run start:prod', got %v", description, data.StartCommand)
				}
				if !data.Runtime.NodeJSVersion.Equal(types.StringValue("20")) {
					t.Errorf("%s: Expected NodeJSVersion '20', got %v", description, data.Runtime.NodeJSVersion)
				}
				if !data.Settings.MemoryRequest.Equal(types.StringValue("1Gi")) {
					t.Errorf("%s: Expected MemoryRequest '1Gi', got %v", description, data.Settings.MemoryRequest)
				}
				if !data.Settings.SchedulerEnabled.Equal(types.BoolValue(true)) {
					t.Errorf("%s: Expected SchedulerEnabled true, got %v", description, data.Settings.SchedulerEnabled)
				}
			},
			description: "API response values should be properly mapped to Terraform state",
		},
		{
			name: "value preservation when api returns different values",
			app: &client.Application{
				ID:            456,
				Name:          "preserve-app",
				Type:          "laravel",
				StartCommand:  "",
				MemoryRequest: "2Gi",
				Status:        "running",
			},
			initialData: &ApplicationResourceModel{
				StartCommand: types.StringValue("php artisan serve"),
				Settings: &SettingsModel{
					MemoryRequest: types.StringValue("1Gi"),
				},
			},
			verify: func(t *testing.T, data *ApplicationResourceModel, description string) {
				if data.StartCommand.IsNull() {
					t.Errorf("%s: StartCommand should be preserved when API returns empty", description)
				}
				if !data.Settings.MemoryRequest.Equal(types.StringValue("2Gi")) {
					t.Errorf("%s: Expected MemoryRequest to be updated to API value '2Gi', got %v", 
						description, data.Settings.MemoryRequest)
				}
			},
			description: "Planned values should be preserved when API returns empty, but API values take precedence when present",
		},
		{
			name: "null and empty field handling",
			app: &client.Application{
				ID:     789,
				Name:   "empty-fields-app",
				Type:   "nodejs",
				Status: "running",
			},
			initialData: &ApplicationResourceModel{
				StartCommand: types.StringNull(),
				Runtime:      &RuntimeModel{},
				Settings:     &SettingsModel{},
			},
			verify: func(t *testing.T, data *ApplicationResourceModel, description string) {
				if !data.StartCommand.IsNull() {
					t.Errorf("%s: StartCommand should remain null when API doesn't provide it", description)
				}
				if !data.Runtime.NodeJSVersion.IsNull() {
					t.Errorf("%s: NodeJSVersion should remain null when API doesn't provide it", description)
				}
				if !data.Settings.MemoryRequest.IsNull() {
					t.Errorf("%s: MemoryRequest should remain null when API doesn't provide it", description)
				}
			},
			description: "Null fields should remain null when API doesn't provide values",
		},
		{
			name: "memory request mismatch handling",
			app: &client.Application{
				ID:            999,
				Name:          "memory-mismatch-app",
				Type:          "nodejs",
				MemoryRequest: "1Gi",
				Status:        "running",
			},
			initialData: &ApplicationResourceModel{
				Settings: &SettingsModel{
					MemoryRequest: types.StringValue("512Mi"),
				},
			},
			verify: func(t *testing.T, data *ApplicationResourceModel, description string) {
				if !data.Settings.MemoryRequest.Equal(types.StringValue("1Gi")) {
					t.Errorf("%s: Expected MemoryRequest to be updated to API value '1Gi' (from planned '512Mi'), got %v", 
						description, data.Settings.MemoryRequest)
				}
			},
			description: "When there's a memory request mismatch, API value should take precedence",
		},
		{
			name: "application type specific version handling",
			app: &client.Application{
				ID:            111,
				Name:          "php-app",
				Type:          "laravel",
				PHPVersion:    "8.4",
				NodeJSVersion: "",
				Status:        "running",
			},
			initialData: &ApplicationResourceModel{
				Runtime: &RuntimeModel{},
			},
			verify: func(t *testing.T, data *ApplicationResourceModel, description string) {
				if !data.Runtime.PHPVersion.Equal(types.StringValue("8.4")) {
					t.Errorf("%s: Expected PHPVersion '8.4', got %v", description, data.Runtime.PHPVersion)
				}
				if !data.Runtime.NodeJSVersion.IsNull() {
					t.Errorf("%s: NodeJSVersion should be null for PHP apps, got %v", description, data.Runtime.NodeJSVersion)
				}
			},
			description: "PHP/Laravel apps should have NodeJSVersion cleared and PHPVersion set",
		},
		{
			name: "nodejs app version handling",
			app: &client.Application{
				ID:            222,
				Name:          "node-app", 
				Type:          "nodejs",
				NodeJSVersion: "22",
				PHPVersion:    "",
				Status:        "running",
			},
			initialData: &ApplicationResourceModel{
				Runtime: &RuntimeModel{},
			},
			verify: func(t *testing.T, data *ApplicationResourceModel, description string) {
				if !data.Runtime.NodeJSVersion.Equal(types.StringValue("22")) {
					t.Errorf("%s: Expected NodeJSVersion '22', got %v", description, data.Runtime.NodeJSVersion)
				}
				if !data.Runtime.PHPVersion.IsNull() {
					t.Errorf("%s: PHPVersion should be null for Node.js apps, got %v", description, data.Runtime.PHPVersion)
				}
			},
			description: "Node.js apps should have PHPVersion cleared and NodeJSVersion set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := *tt.initialData

			resource.fromAPIModel(tt.app, &data)

			tt.verify(t, &data, tt.description)
		})
	}
}

// TestFieldPreservation tests specific field preservation scenarios from the plan
func TestFieldPreservation(t *testing.T) {
	resource := &ApplicationResource{}

	tests := []struct {
		name         string
		plannedValue interface{}
		apiValue     interface{}
		fieldName    string
		shouldUpdate bool
		description  string
	}{
		{
			name:         "start_command preservation when api returns empty",
			plannedValue: "npm run prod",
			apiValue:     "",
			fieldName:    "start_command",
			shouldUpdate: false,
			description:  "Planned start_command should be preserved when API returns empty string",
		},
		{
			name:         "memory_request update when api returns different value",
			plannedValue: "512Mi",
			apiValue:     "1Gi",
			fieldName:    "memory_request",
			shouldUpdate: true,
			description:  "Memory request should be updated when API returns different value",
		},
		{
			name:         "nodejs_version preservation when api returns empty",
			plannedValue: "20",
			apiValue:     "",
			fieldName:    "nodejs_version",
			shouldUpdate: false,
			description:  "Planned nodejs_version should be preserved when API returns empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data ApplicationResourceModel
			var app client.Application

			switch tt.fieldName {
			case "start_command":
				data.StartCommand = types.StringValue(tt.plannedValue.(string))
				app.StartCommand = tt.apiValue.(string)
			case "memory_request":
				data.Settings = &SettingsModel{
					MemoryRequest: types.StringValue(tt.plannedValue.(string)),
				}
				app.MemoryRequest = tt.apiValue.(string)
			case "nodejs_version":
				data.Runtime = &RuntimeModel{
					NodeJSVersion: types.StringValue(tt.plannedValue.(string)),
				}
				app.Type = "nodejs"
				app.NodeJSVersion = tt.apiValue.(string)
			}

			app.ID = 1
			app.Name = "test-app"
			app.Status = "running"

			resource.fromAPIModel(&app, &data)

			var actualValue string
			switch tt.fieldName {
			case "start_command":
				actualValue = data.StartCommand.ValueString()
			case "memory_request":
				actualValue = data.Settings.MemoryRequest.ValueString()
			case "nodejs_version":
				actualValue = data.Runtime.NodeJSVersion.ValueString()
			}

			expectedValue := tt.plannedValue.(string)
			if tt.shouldUpdate {
				expectedValue = tt.apiValue.(string)
			}

			if actualValue != expectedValue {
				t.Errorf("%s: Expected %s = '%s', got '%s'",
					tt.description, tt.fieldName, expectedValue, actualValue)
			}
		})
	}
}

// TestNullHandling tests null/empty value scenarios per the plan
func TestNullHandling(t *testing.T) {
	resource := &ApplicationResource{}

	tests := []struct {
		name        string
		setupData   func() *ApplicationResourceModel
		setupAPI    func() *client.Application
		verify      func(t *testing.T, data *ApplicationResourceModel)
		description string
	}{
		{
			name: "null terraform values with non-empty api response",
			setupData: func() *ApplicationResourceModel {
				return &ApplicationResourceModel{
					StartCommand: types.StringNull(),
					Runtime:      &RuntimeModel{NodeJSVersion: types.StringNull()},
					Settings:     &SettingsModel{MemoryRequest: types.StringNull()},
				}
			},
			setupAPI: func() *client.Application {
				return &client.Application{
					ID:            1,
					Name:          "test-app",
					Type:          "nodejs",
					StartCommand:  "npm start",
					NodeJSVersion: "20",
					MemoryRequest: "1Gi",
					Status:        "running",
				}
			},
			verify: func(t *testing.T, data *ApplicationResourceModel) {
				if !data.StartCommand.Equal(types.StringValue("npm start")) {
					t.Errorf("Expected StartCommand to be set from API, got %v", data.StartCommand)
				}
				if !data.Runtime.NodeJSVersion.Equal(types.StringValue("20")) {
					t.Errorf("Expected NodeJSVersion to be set from API, got %v", data.Runtime.NodeJSVersion)
				}
				if !data.Settings.MemoryRequest.Equal(types.StringValue("1Gi")) {
					t.Errorf("Expected MemoryRequest to be set from API, got %v", data.Settings.MemoryRequest)
				}
			},
			description: "Null Terraform values should be set when API provides non-empty values",
		},
		{
			name: "empty api values with null terraform values",
			setupData: func() *ApplicationResourceModel {
				return &ApplicationResourceModel{
					StartCommand: types.StringNull(),
					Runtime:      &RuntimeModel{NodeJSVersion: types.StringNull()},
					Settings:     &SettingsModel{MemoryRequest: types.StringNull()},
				}
			},
			setupAPI: func() *client.Application {
				return &client.Application{
					ID:            1,
					Name:          "test-app",
					Type:          "nodejs",
					StartCommand:  "",
					NodeJSVersion: "",
					MemoryRequest: "",
					Status:        "running",
				}
			},
			verify: func(t *testing.T, data *ApplicationResourceModel) {
				if !data.StartCommand.IsNull() {
					t.Errorf("Expected StartCommand to remain null, got %v", data.StartCommand)
				}
				if !data.Runtime.NodeJSVersion.IsNull() {
					t.Errorf("Expected NodeJSVersion to remain null, got %v", data.Runtime.NodeJSVersion)
				}
				if !data.Settings.MemoryRequest.IsNull() {
					t.Errorf("Expected MemoryRequest to remain null, got %v", data.Settings.MemoryRequest)
				}
			},
			description: "Null Terraform values should remain null when API returns empty strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.setupData()
			app := tt.setupAPI()

			resource.fromAPIModel(app, data)

			tt.verify(t, data)
		})
	}
}

// TestConsistencyErrorScenarios tests specific scenarios that were causing consistency errors
func TestConsistencyErrorScenarios(t *testing.T) {
	resource := &ApplicationResource{}

	t.Run("missing start_command in update causes consistency error", func(t *testing.T) {
		data := &ApplicationResourceModel{
			StartCommand: types.StringValue("npm run production"),
		}

		updatePayload := resource.toUpdateAPIModel(data)

		if _, exists := updatePayload["start_command"]; !exists {
			t.Error("start_command must be included in update payload to prevent consistency errors")
		}

		if updatePayload["start_command"] != "npm run production" {
			t.Errorf("Expected start_command = 'npm run production', got %v", updatePayload["start_command"])
		}
	})

	t.Run("missing nodejs_version in update causes consistency error", func(t *testing.T) {
		data := &ApplicationResourceModel{
			Runtime: &RuntimeModel{
				NodeJSVersion: types.StringValue("20"),
			},
		}

		updatePayload := resource.toUpdateAPIModel(data)

		if _, exists := updatePayload["nodejs_version"]; !exists {
			t.Error("nodejs_version must be included in update payload to prevent consistency errors")
		}

		if updatePayload["nodejs_version"] != "20" {
			t.Errorf("Expected nodejs_version = '20', got %v", updatePayload["nodejs_version"])
		}
	})

	t.Run("missing memory_request in update causes consistency error", func(t *testing.T) {
		data := &ApplicationResourceModel{
			Settings: &SettingsModel{
				MemoryRequest: types.StringValue("512Mi"),
			},
		}

		updatePayload := resource.toUpdateAPIModel(data)

		if _, exists := updatePayload["memory_request"]; !exists {
			t.Error("memory_request must be included in update payload to prevent consistency errors")
		}

		if updatePayload["memory_request"] != "512Mi" {
			t.Errorf("Expected memory_request = '512Mi', got %v", updatePayload["memory_request"])
		}
	})
}

// TestIntegrationWorkflow tests the complete workflow to ensure no consistency errors occur
func TestIntegrationWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch {
		case r.URL.Path == "/applications" && r.Method == "POST":
			response := `{
				"success": true,
				"data": {
					"id": 100,
					"name": "integration-test-app",
					"application_type": "nodejs",
					"start_command": "npm run production",
					"nodejs_version": "20",
					"memory_request": "1Gi",
					"health_check_path": "/health",
					"scheduler_enabled": false,
					"replicas": 2,
					"cpu_request": "500m",
					"status": "creating",
					"url": "https://integration-test-app.ploi.cloud",
					"needs_deployment": true
				}
			}`
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(response))

		case r.URL.Path == "/applications/100" && r.Method == "PUT":
			response := `{
				"success": true,
				"data": {
					"id": 100,
					"name": "integration-test-app",
					"application_type": "nodejs",
					"start_command": "node --max-old-space-size=4096 server.js",
					"nodejs_version": "22",
					"memory_request": "2Gi",
					"health_check_path": "/status",
					"scheduler_enabled": false,
					"replicas": 3,
					"cpu_request": "1000m",
					"status": "running",
					"url": "https://integration-test-app.ploi.cloud",
					"needs_deployment": false
				}
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))

		case r.URL.Path == "/applications/100" && r.Method == "GET":
			response := `{
				"success": true,
				"data": {
					"id": 100,
					"name": "integration-test-app",
					"application_type": "nodejs",
					"start_command": "node --max-old-space-size=4096 server.js",
					"nodejs_version": "22",
					"memory_request": "2Gi",
					"health_check_path": "/status",
					"scheduler_enabled": false,
					"replicas": 3,
					"cpu_request": "1000m",
					"status": "running",
					"url": "https://integration-test-app.ploi.cloud",
					"needs_deployment": false
				}
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	c := client.NewClient("test-token", &server.URL)
	resource := &ApplicationResource{client: c}

	t.Run("create read update workflow", func(t *testing.T) {
		createData := &ApplicationResourceModel{
			Name:         types.StringValue("integration-test-app"),
			Type:         types.StringValue("nodejs"),
			StartCommand: types.StringValue("npm run production"),
			Runtime: &RuntimeModel{
				NodeJSVersion: types.StringValue("20"),
			},
			Settings: &SettingsModel{
				MemoryRequest:    types.StringValue("1Gi"),
				HealthCheckPath:  types.StringValue("/health"),
				SchedulerEnabled: types.BoolValue(false),
				Replicas:         types.Int64Value(2),
				CPURequest:       types.StringValue("500m"),
			},
		}

		apiModel := resource.toAPIModel(createData)
		created, err := c.CreateApplication(apiModel)
		if err != nil {
			t.Fatalf("Failed to create application: %v", err)
		}

		var createdData ApplicationResourceModel
		resource.fromAPIModel(created, &createdData)

		if createdData.ID.ValueInt64() != 100 {
			t.Errorf("Expected ID 100, got %d", createdData.ID.ValueInt64())
		}

		updateData := ApplicationResourceModel{
			ID:           createdData.ID,
			Name:         types.StringValue("integration-test-app"),
			Type:         types.StringValue("nodejs"),
			StartCommand: types.StringValue("node --max-old-space-size=4096 server.js"),
			Runtime: &RuntimeModel{
				NodeJSVersion: types.StringValue("22"),
			},
			Settings: &SettingsModel{
				MemoryRequest:    types.StringValue("2Gi"),
				HealthCheckPath:  types.StringValue("/status"),
				SchedulerEnabled: types.BoolValue(false),
				Replicas:         types.Int64Value(3),
				CPURequest:       types.StringValue("1000m"),
			},
		}

		updatePayload := resource.toUpdateAPIModel(&updateData)
		
		expectedFields := []string{"start_command", "nodejs_version", "memory_request", "health_check_path", "cpu_request", "replicas"}
		for _, field := range expectedFields {
			if _, exists := updatePayload[field]; !exists {
				t.Errorf("Expected field '%s' to be included in update payload", field)
			}
		}

		updated, err := c.UpdateApplication(createdData.ID.ValueInt64(), updatePayload)
		if err != nil {
			t.Fatalf("Failed to update application: %v", err)
		}

		var updatedData ApplicationResourceModel
		resource.fromAPIModel(updated, &updatedData)

		if !updatedData.StartCommand.Equal(types.StringValue("node --max-old-space-size=4096 server.js")) {
			t.Errorf("Expected updated StartCommand, got %v", updatedData.StartCommand)
		}
		if !updatedData.Runtime.NodeJSVersion.Equal(types.StringValue("22")) {
			t.Errorf("Expected updated NodeJSVersion, got %v", updatedData.Runtime.NodeJSVersion)
		}
		if !updatedData.Settings.MemoryRequest.Equal(types.StringValue("2Gi")) {
			t.Errorf("Expected updated MemoryRequest, got %v", updatedData.Settings.MemoryRequest)
		}
	})
}

// TestEdgeCasesAndErrorScenarios tests edge cases and error scenarios
func TestEdgeCasesAndErrorScenarios(t *testing.T) {
	resource := &ApplicationResource{}

	t.Run("empty model should not cause issues", func(t *testing.T) {
		data := &ApplicationResourceModel{}
		
		result := resource.toUpdateAPIModel(data)
		
		if len(result) != 0 {
			t.Errorf("Expected empty update payload for empty model, got %v", result)
		}
	})

	t.Run("nil runtime and settings blocks", func(t *testing.T) {
		data := &ApplicationResourceModel{
			Name: types.StringValue("test-app"),
		}
		
		result := resource.toUpdateAPIModel(data)
		
		if result["name"] != "test-app" {
			t.Errorf("Expected name = 'test-app', got %v", result["name"])
		}
		
		if _, exists := result["nodejs_version"]; exists {
			t.Error("nodejs_version should not be included when Runtime is nil")
		}
		if _, exists := result["memory_request"]; exists {
			t.Error("memory_request should not be included when Settings is nil")
		}
	})

	t.Run("api response with missing fields", func(t *testing.T) {
		app := &client.Application{
			ID:     123,
			Name:   "minimal-app",
			Type:   "nodejs",
			Status: "running",
		}
		
		var data ApplicationResourceModel
		data.Runtime = &RuntimeModel{}
		data.Settings = &SettingsModel{}
		
		resource.fromAPIModel(app, &data)
		
		if !data.StartCommand.IsNull() {
			t.Error("StartCommand should be null when API doesn't provide it")
		}
		if !data.Runtime.NodeJSVersion.IsNull() {
			t.Error("NodeJSVersion should be null when API doesn't provide it")
		}
		if !data.Settings.MemoryRequest.IsNull() {
			t.Error("MemoryRequest should be null when API doesn't provide it")
		}
	})
}

// Helper function for deep comparison of interface{} values
func deepEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch aVal := a.(type) {
	case []string:
		if bVal, ok := b.([]string); ok {
			if len(aVal) != len(bVal) {
				return false
			}
			for i, v := range aVal {
				if v != bVal[i] {
					return false
				}
			}
			return true
		}
		return false
	case string:
		bVal, ok := b.(string)
		return ok && aVal == bVal
	case int64:
		bVal, ok := b.(int64)
		return ok && aVal == bVal
	case bool:
		bVal, ok := b.(bool)
		return ok && aVal == bVal
	default:
		return a == b
	}
}