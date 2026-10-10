package print

import (
	"bytes"
	"testing"
)

func TestSuccessStatusEvent(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple message",
			format:   "Test completed successfully",
			args:     nil,
			expected: "    ✅ Test completed successfully\n",
		},
		{
			name:     "message with formatting",
			format:   "Pod %s is ready",
			args:     []interface{}{"test-pod"},
			expected: "    ✅ Pod test-pod is ready\n",
		},
		{
			name:     "message with multiple args",
			format:   "Service %s created in namespace %s",
			args:     []interface{}{"my-service", "default"},
			expected: "    ✅ Service my-service created in namespace default\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			SuccessStatusEvent(&buf, tt.format, tt.args...)
			
			result := buf.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFailureStatusEvent(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple message",
			format:   "Test failed",
			args:     nil,
			expected: "    ❌ Test failed\n",
		},
		{
			name:     "message with formatting",
			format:   "Pod %s failed to start",
			args:     []interface{}{"test-pod"},
			expected: "    ❌ Pod test-pod failed to start\n",
		},
		{
			name:     "message with multiple args",
			format:   "Service %s failed in namespace %s",
			args:     []interface{}{"my-service", "default"},
			expected: "    ❌ Service my-service failed in namespace default\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			FailureStatusEvent(&buf, tt.format, tt.args...)
			
			result := buf.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPendingStatusEvent(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "simple message",
			format:   "Test pending",
			args:     nil,
			expected: "    ⌛ Test pending\n",
		},
		{
			name:     "message with formatting",
			format:   "Pod %s is pending",
			args:     []interface{}{"test-pod"},
			expected: "    ⌛ Pod test-pod is pending\n",
		},
		{
			name:     "message with multiple args",
			format:   "Service %s pending in namespace %s",
			args:     []interface{}{"my-service", "default"},
			expected: "    ⌛ Service my-service pending in namespace default\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PendingStatusEvent(&buf, tt.format, tt.args...)
			
			result := buf.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}