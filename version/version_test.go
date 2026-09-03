package version

import "testing"

func TestVersion(t *testing.T) {
	// Test default version
	if Version == "" {
		t.Error("Version should have a default value")
	}

	// The default version should be "dev"
	expectedDefault := "dev"
	if Version != expectedDefault {
		t.Errorf("Expected default version to be %q, got %q", expectedDefault, Version)
	}
}

func TestVersionIsString(t *testing.T) {
	// Ensure Version is a string type
	if _, ok := interface{}(Version).(string); !ok {
		t.Error("Version should be of type string")
	}
}