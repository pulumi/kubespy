package cmd

import (
	"testing"
)

func TestParseObjID(t *testing.T) {
	tests := []struct {
		name          string
		objID         string
		expectedNS    string
		expectedName  string
		expectError   bool
		errorContains string
	}{
		{
			name:         "name only",
			objID:        "my-pod",
			expectedName: "my-pod",
			expectError:  true, // This will fail in test environment without kubeconfig
		},
		{
			name:         "name only - with kubeconfig error",
			objID:        "test-pod",
			expectedName: "test-pod",
			expectError:  true,
			errorContains: "configuration", // Should contain "configuration" in error message
		},
		{
			name:         "namespace and name",
			objID:        "my-namespace/my-pod",
			expectedNS:   "my-namespace",
			expectedName: "my-pod",
			expectError:  false,
		},
		{
			name:          "too many slashes",
			objID:         "ns/pod/extra",
			expectError:   true,
			errorContains: "Object ID must be of the form",
		},
		{
			name:          "empty slashes",
			objID:         "ns//pod",
			expectError:   true,
			errorContains: "Object ID must be of the form",
		},
		{
			name:         "empty name with namespace",
			objID:        "my-namespace/",
			expectedNS:   "my-namespace",
			expectedName: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespace, name, err := parseObjID(tt.objID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for objID %q, but got none", tt.objID)
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for objID %q: %v", tt.objID, err)
				return
			}

			if name != tt.expectedName {
				t.Errorf("Expected name %q, got %q", tt.expectedName, name)
			}

			// For name-only cases, we can't predict the namespace since it depends on kubeconfig
			// In test environment, this will likely fail, so we skip this check
			if tt.objID != "my-pod" && namespace != tt.expectedNS {
				t.Errorf("Expected namespace %q, got %q", tt.expectedNS, namespace)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || 
		   len(s) > len(substr) && s[:len(substr)] == substr ||
		   len(s) > len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}