package pods

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetReady(t *testing.T) {
	tests := []struct {
		name           string
		endpoints      *unstructured.Unstructured
		expectedCount  int
		expectedFormat bool // Whether to check if output contains expected formatting
	}{
		{
			name: "endpoints with ready addresses",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{
							"addresses": []interface{}{
								map[string]interface{}{
									"ip": "192.168.1.10",
									"targetRef": map[string]interface{}{
										"name": "pod-1",
									},
								},
								map[string]interface{}{
									"ip": "192.168.1.11",
									"targetRef": map[string]interface{}{
										"name": "pod-2",
									},
								},
							},
						},
					},
				},
			},
			expectedCount:  2,
			expectedFormat: true,
		},
		{
			name: "endpoints with no subsets",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with invalid subsets",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": "invalid",
				},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with no addresses",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{},
					},
				},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with invalid addresses",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{
							"addresses": "invalid",
						},
					},
				},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with address missing targetRef",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{
							"addresses": []interface{}{
								map[string]interface{}{
									"ip": "192.168.1.10",
								},
							},
						},
					},
				},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with address missing IP",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{
							"addresses": []interface{}{
								map[string]interface{}{
									"targetRef": map[string]interface{}{
										"name": "pod-1",
									},
								},
							},
						},
					},
				},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ready := GetReady(tt.endpoints)
			if len(ready) != tt.expectedCount {
				t.Errorf("Expected %d ready pods, got %d", tt.expectedCount, len(ready))
			}

			if tt.expectedFormat && len(ready) > 0 {
				// Check that the format contains expected elements
				for _, readyPod := range ready {
					if !containsString(readyPod, "Ready") {
						t.Errorf("Expected ready pod format to contain 'Ready', got: %s", readyPod)
					}
					if !containsString(readyPod, "@") {
						t.Errorf("Expected ready pod format to contain '@', got: %s", readyPod)
					}
				}
			}
		})
	}
}

func TestGetUnready(t *testing.T) {
	tests := []struct {
		name           string
		endpoints      *unstructured.Unstructured
		expectedCount  int
		expectedFormat bool // Whether to check if output contains expected formatting
	}{
		{
			name: "endpoints with unready addresses",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{
							"notReadyAddresses": []interface{}{
								map[string]interface{}{
									"ip": "192.168.1.10",
									"targetRef": map[string]interface{}{
										"name": "pod-1",
									},
								},
								map[string]interface{}{
									"ip": "192.168.1.11",
									"targetRef": map[string]interface{}{
										"name": "pod-2",
									},
								},
							},
						},
					},
				},
			},
			expectedCount:  2,
			expectedFormat: true,
		},
		{
			name: "endpoints with no subsets",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with invalid subsets",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": "invalid",
				},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with no notReadyAddresses",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{},
					},
				},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with invalid notReadyAddresses",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{
							"notReadyAddresses": "invalid",
						},
					},
				},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with address missing targetRef",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{
							"notReadyAddresses": []interface{}{
								map[string]interface{}{
									"ip": "192.168.1.10",
								},
							},
						},
					},
				},
			},
			expectedCount: 0,
		},
		{
			name: "endpoints with address missing IP",
			endpoints: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"subsets": []interface{}{
						map[string]interface{}{
							"notReadyAddresses": []interface{}{
								map[string]interface{}{
									"targetRef": map[string]interface{}{
										"name": "pod-1",
									},
								},
							},
						},
					},
				},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unready := GetUnready(tt.endpoints)
			if len(unready) != tt.expectedCount {
				t.Errorf("Expected %d unready pods, got %d", tt.expectedCount, len(unready))
			}

			if tt.expectedFormat && len(unready) > 0 {
				// Check that the format contains expected elements
				for _, unreadyPod := range unready {
					if !containsString(unreadyPod, "Not live") {
						t.Errorf("Expected unready pod format to contain 'Not live', got: %s", unreadyPod)
					}
					if !containsString(unreadyPod, "@") {
						t.Errorf("Expected unready pod format to contain '@', got: %s", unreadyPod)
					}
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}