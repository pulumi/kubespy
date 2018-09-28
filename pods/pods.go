package pods

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/pulumi/pulumi-kubernetes/pkg/openapi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	greenText    = color.New(color.FgGreen)
	yellowText   = color.New(color.FgYellow)
	cyanBoldText = color.New(color.FgCyan, color.Bold)
	cyanText     = color.New(color.FgCyan)
	redBoldText  = color.New(color.FgRed, color.Bold)
)

func GetReady(o *unstructured.Unstructured) []string {
	ready := []string{}
	subsetsI, _ := openapi.Pluck(o.Object, "subsets")
	subsets, isSlice := subsetsI.([]interface{})
	if !isSlice {
		return ready
	}

	for _, subsetI := range subsets {
		subset, isMap := subsetI.(map[string]interface{})
		if !isMap {
			continue
		}

		addressesI, _ := subset["addresses"]
		addresses, isSlice := addressesI.([]interface{})
		if !isSlice {
			continue
		}

		for _, addressI := range addresses {
			address, isMap := addressI.(map[string]interface{})
			if !isMap {
				continue
			}

			nameI, _ := openapi.Pluck(address, "targetRef", "name")
			name, isString := nameI.(string)
			if !isString {
				continue
			}

			ipI, _ := address["ip"]
			ip, isString := ipI.(string)
			if !isString {
				continue
			}
			ready = append(ready, fmt.Sprintf("[%s] %s @ %s", greenText.Sprint("Ready"), cyanText.Sprint(name), yellowText.Sprint(ip)))
		}
	}
	return ready
}

func GetUnready(o *unstructured.Unstructured) []string {
	ready := []string{}
	subsetsI, _ := openapi.Pluck(o.Object, "subsets")
	subsets, isSlice := subsetsI.([]interface{})
	if !isSlice {
		return ready
	}

	for _, subsetI := range subsets {
		subset, isMap := subsetI.(map[string]interface{})
		if !isMap {
			continue
		}

		addressesI, _ := subset["notReadyAddresses"]
		addresses, isSlice := addressesI.([]interface{})
		if !isSlice {
			continue
		}

		for _, addressI := range addresses {
			address, isMap := addressI.(map[string]interface{})
			if !isMap {
				continue
			}

			nameI, _ := openapi.Pluck(address, "targetRef", "name")
			name, isString := nameI.(string)
			if !isString {
				continue
			}

			ipI, _ := address["ip"]
			ip, isString := ipI.(string)
			if !isString {
				continue
			}
			ready = append(ready, fmt.Sprintf("[%s] %s @ %s", redBoldText.Sprint("Not live"), cyanText.Sprint(name), yellowText.Sprint(ip)))
		}
	}
	return ready
}
