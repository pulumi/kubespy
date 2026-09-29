package watch

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestAll(t *testing.T) {
	namespace := "test-namespace"
	opts := All(namespace)
	
	if opts.watchType != watchAll {
		t.Errorf("Expected watchType to be %v, got %v", watchAll, opts.watchType)
	}
	
	if opts.namespace != namespace {
		t.Errorf("Expected namespace to be %q, got %q", namespace, opts.namespace)
	}
}

func TestThisObject(t *testing.T) {
	namespace := "test-namespace"
	name := "test-object"
	opts := ThisObject(namespace, name)
	
	if opts.watchType != watchByName {
		t.Errorf("Expected watchType to be %v, got %v", watchByName, opts.watchType)
	}
	
	if opts.namespace != namespace {
		t.Errorf("Expected namespace to be %q, got %q", namespace, opts.namespace)
	}
	
	if opts.name != name {
		t.Errorf("Expected name to be %q, got %q", name, opts.name)
	}
}

func TestObjectsOwnedBy(t *testing.T) {
	namespace := "test-namespace"
	ownerName := "test-owner"
	opts := ObjectsOwnedBy(namespace, ownerName)
	
	if opts.watchType != watchByOwner {
		t.Errorf("Expected watchType to be %v, got %v", watchByOwner, opts.watchType)
	}
	
	if opts.namespace != namespace {
		t.Errorf("Expected namespace to be %q, got %q", namespace, opts.namespace)
	}
	
	if opts.ownerName != ownerName {
		t.Errorf("Expected ownerName to be %q, got %q", ownerName, opts.ownerName)
	}
}

func TestOptsCheck(t *testing.T) {
	tests := []struct {
		name           string
		opts           Opts
		object         *unstructured.Unstructured
		expectedResult bool
	}{
		{
			name: "watchByName - matching name",
			opts: Opts{
				watchType: watchByName,
				name:      "test-pod",
			},
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-pod",
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "watchByName - non-matching name",
			opts: Opts{
				watchType: watchByName,
				name:      "test-pod",
			},
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "other-pod",
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "watchAll - always true",
			opts: Opts{
				watchType: watchAll,
			},
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "any-object",
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "watchByOwner - apps/v1 deployment owner match",
			opts: Opts{
				watchType: watchByOwner,
				ownerName: "test-deployment",
			},
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "apps/v1",
								"kind":       "Deployment",
								"name":       "test-deployment",
							},
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "watchByOwner - extensions/v1beta1 deployment owner match",
			opts: Opts{
				watchType: watchByOwner,
				ownerName: "test-deployment",
			},
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "extensions/v1beta1",
								"kind":       "Deployment",
								"name":       "test-deployment",
							},
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "watchByOwner - apps/v1beta1 deployment owner match",
			opts: Opts{
				watchType: watchByOwner,
				ownerName: "test-deployment",
			},
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"ownerReferences": []interface{}{
							map[string]interface{}{
								"apiVersion": "apps/v1beta1",
								"kind":       "Deployment",
								"name":       "test-deployment",
							},
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "watchByOwner - no owner match",
			opts: Opts{
				watchType: watchByOwner,
				ownerName: "test-deployment",
			},
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
			expectedResult: false,
		},
		{
			name: "watchByOwner - no owner references",
			opts: Opts{
				watchType: watchByOwner,
				ownerName: "test-deployment",
			},
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{},
				},
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.Check(tt.object)
			if result != tt.expectedResult {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestOptsCheckPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for unknown watch type, but didn't panic")
		}
	}()

	opts := Opts{
		watchType: "invalid",
	}
	object := &unstructured.Unstructured{}
	opts.Check(object)
}