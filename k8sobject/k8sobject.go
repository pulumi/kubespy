package k8sobject

import (
	"github.com/pulumi/pulumi-kubernetes/pkg/openapi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func OwnedBy(o *unstructured.Unstructured, apiVersion, kind, ownerName interface{}) bool {
	ownerReferencesI, _ := openapi.Pluck(o.Object, "metadata", "ownerReferences")
	ownerReferences, isSlice := ownerReferencesI.([]interface{})
	if !isSlice {
		return false
	}

	for _, refI := range ownerReferences {
		ref, isMap := refI.(map[string]interface{})
		if !isMap {
			continue
		}

		apiVersion := ref["apiVersion"]
		if ref["kind"] == kind && apiVersion == ref["apiVersion"] && ref["name"] == ownerName {
			return true
		}
	}

	return false
}

func PodConditions(pod *unstructured.Unstructured) []interface{} {
	statusI, _ := openapi.Pluck(pod.Object, "status")
	status, isMap := statusI.(map[string]interface{})
	if !isMap {
		return []interface{}{}
	}

	conditionsI, _ := status["conditions"]
	conditions, isSlice := conditionsI.([]interface{})
	if !isSlice {
		return []interface{}{}
	}

	return conditions
}

func PodContainerStatuses(pod *unstructured.Unstructured) []interface{} {
	statusI, _ := openapi.Pluck(pod.Object, "status")
	status, isMap := statusI.(map[string]interface{})
	if !isMap {
		return []interface{}{}
	}

	containerStatusesI, _ := status["containerStatuses"]
	containerStatuses, isSlice := containerStatusesI.([]interface{})
	if !isSlice {
		return []interface{}{}
	}

	return containerStatuses
}
