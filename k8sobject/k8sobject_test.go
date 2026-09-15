package k8sobject

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestOwnedBy(t *testing.T) {
	tests := []struct {
		name           string
		object         *unstructured.Unstructured
		apiVersion     interface{}
		kind           interface{}
		ownerName      interface{}
		expectedResult bool
	}{
		{
			name: "object owned by deployment",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "apps/v1",
								"kind":       "Deployment",
								"name":       "my-deployment",
							},
						},
					},
				},
			},
			apiVersion:     "apps/v1",
			kind:           "Deployment",
			ownerName:      "my-deployment",
			expectedResult: true,
		},
		{
			name: "object not owned by specified deployment",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "apps/v1",
								"kind":       "Deployment",
								"name":       "other-deployment",
							},
						},
					},
				},
			},
			apiVersion:     "apps/v1",
			kind:           "Deployment",
			ownerName:      "my-deployment",
			expectedResult: false,
		},
		{
			name: "object with no owner references",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{},
				},
			},
			apiVersion:     "apps/v1",
			kind:           "Deployment",
			ownerName:      "my-deployment",
			expectedResult: false,
		},
		{
			name: "object with invalid owner references",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": "invalid",
					},
				},
			},
			apiVersion:     "apps/v1",
			kind:           "Deployment",
			ownerName:      "my-deployment",
			expectedResult: false,
		},
		{
			name: "object with owner reference wrong kind",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "apps/v1",
								"kind":       "ReplicaSet",
								"name":       "my-deployment",
							},
						},
					},
				},
			},
			apiVersion:     "apps/v1",
			kind:           "Deployment",
			ownerName:      "my-deployment",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OwnedBy(tt.object, tt.apiVersion, tt.kind, tt.ownerName)
			if result != tt.expectedResult {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestPodConditions(t *testing.T) {
	tests := []struct {
		name               string
		pod                *unstructured.Unstructured
		expectedConditions int
	}{
		{
			name: "pod with conditions",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
							map[string]interface{}{
								"type":   "PodScheduled",
								"status": "True",
							},
						},
					},
				},
			},
			expectedConditions: 2,
		},
		{
			name: "pod with no conditions",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{},
				},
			},
			expectedConditions: 0,
		},
		{
			name: "pod with no status",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			expectedConditions: 0,
		},
		{
			name: "pod with invalid status",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": "invalid",
				},
			},
			expectedConditions: 0,
		},
		{
			name: "pod with invalid conditions",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": "invalid",
					},
				},
			},
			expectedConditions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conditions := PodConditions(tt.pod)
			if len(conditions) != tt.expectedConditions {
				t.Errorf("Expected %d conditions, got %d", tt.expectedConditions, len(conditions))
			}
		})
	}
}

func TestPodContainerStatuses(t *testing.T) {
	tests := []struct {
		name                     string
		pod                      *unstructured.Unstructured
		expectedContainerStatuses int
	}{
		{
			name: "pod with container statuses",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"containerStatuses": []interface{}{
							map[string]interface{}{
								"name":  "container1",
								"ready": true,
							},
							map[string]interface{}{
								"name":  "container2",
								"ready": false,
							},
						},
					},
				},
			},
			expectedContainerStatuses: 2,
		},
		{
			name: "pod with no container statuses",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{},
				},
			},
			expectedContainerStatuses: 0,
		},
		{
			name: "pod with no status",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			expectedContainerStatuses: 0,
		},
		{
			name: "pod with invalid status",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": "invalid",
				},
			},
			expectedContainerStatuses: 0,
		},
		{
			name: "pod with invalid container statuses",
			pod: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"containerStatuses": "invalid",
					},
				},
			},
			expectedContainerStatuses: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerStatuses := PodContainerStatuses(tt.pod)
			if len(containerStatuses) != tt.expectedContainerStatuses {
				t.Errorf("Expected %d container statuses, got %d", tt.expectedContainerStatuses, len(containerStatuses))
			}
		})
	}
}