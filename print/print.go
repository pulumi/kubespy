package print

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/pulumi/kubespy/pods"
	"github.com/pulumi/pulumi-kubernetes/pkg/openapi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sWatch "k8s.io/apimachinery/pkg/watch"
)

const (
	v1Endpoints = "v1/Endpoints"
	v1Service   = "v1/Service"
	prefix      = "\n       - "
)

var (
	greenText    = color.New(color.FgGreen)
	yellowText   = color.New(color.FgYellow)
	cyanBoldText = color.New(color.FgCyan, color.Bold)
	cyanText     = color.New(color.FgCyan)
	redBoldText  = color.New(color.FgRed, color.Bold)
)

func SuccessStatusEvent(w io.Writer, message string) {
	fmt.Fprintf(w, "    ✅ %s\n", message)
}
func FailureStatusEvent(w io.Writer, message string) {
	fmt.Fprintf(w, "    ❌ %s\n", message)
}

func ServiceWatchTable(w *uilive.Writer, table map[string][]k8sWatch.Event) {
	var svcType string
	if events, hasSvc := table[v1Service]; hasSvc {
		o := events[0].Object.(*unstructured.Unstructured)
		watchEventHeader(w, events[0].Type, o)

		svcTypeI, _ := openapi.Pluck(o.Object, "spec", "type")
		var isString bool
		svcType, isString = svcTypeI.(string)
		if !isString {
			svcType = "ClusterIP"
		}

		switch svcType {
		case "ClusterIP":
			if eps, hasEndpoints := table[v1Endpoints]; hasEndpoints && eps[0].Type != k8sWatch.Deleted {
				SuccessStatusEvent(w,
					fmt.Sprintf("Successfully created Endpoints object '%s' to direct traffic to Pods",
						cyanText.Sprint(o.GetName())))
			} else {
				FailureStatusEvent(w, "Waiting for Endpoints object to be created, to direct traffic to Pods")
			}

			clusterIPI, _ := openapi.Pluck(o.Object, "spec", "clusterIP")
			if clusterIP, isString := clusterIPI.(string); isString && len(clusterIP) > 0 {
				SuccessStatusEvent(w,
					fmt.Sprintf("Successfully allocated a cluster-internal IP: %s",
						cyanText.Sprint(o.GetName())))
			} else {
				FailureStatusEvent(w, "Waiting for cluster-internal IP to be allocated")
			}
		case "LoadBalancer":
			if eps, hasEndpoints := table[v1Endpoints]; hasEndpoints && eps[0].Type != k8sWatch.Deleted {
				SuccessStatusEvent(w,
					fmt.Sprintf("Successfully created Endpoints object '%s' to direct traffic to Pods", o.GetName()))
			} else {
				FailureStatusEvent(w, "Waiting for Endpoints object to be created, to direct traffic to Pods")
			}

			ingressesI, _ := openapi.Pluck(o.Object, "status", "loadBalancer", "ingress")
			if ingresses, isMap := ingressesI.([]interface{}); isMap {
				ips := []string{}
				for _, ingressI := range ingresses {
					if ingress, isMap := ingressI.(map[string]interface{}); isMap {
						tmp := []string{}
						ipI, _ := ingress["ip"]
						ip, isString := ipI.(string)
						if isString {
							tmp = append(tmp, cyanText.Sprint(ip))
						}

						hostnameI, _ := ingress["hostname"]
						hostname, isString := hostnameI.(string)
						if isString {
							tmp = append(tmp, cyanText.Sprint(hostname))
						}

						ips = append(ips, strings.Join(tmp, "/"))
					}
				}

				if len(ips) > 0 {
					sort.Strings(ips)
					li := prefix + strings.Join(ips, prefix)
					SuccessStatusEvent(w, fmt.Sprintf("Service allocated the following IPs/hostnames:%s", li))
				}
			} else {
				FailureStatusEvent(w, "Waiting for public IP/host to be allocated")
			}

		case "ExternalName":
			externalNameI, _ := openapi.Pluck(o.Object, "spec", "externalName")
			if externalName, isString := externalNameI.(string); isString && len(externalName) > 0 {
				SuccessStatusEvent(w, fmt.Sprintf("Service proxying to %s", cyanText.Sprint(externalName)))
			} else {
				FailureStatusEvent(w, "Service not given a URI to proxy to in `.spec.externalName`")
			}
		}
	}

	fmt.Fprintln(w)

	if events, hasEPs := table[v1Endpoints]; hasEPs {
		o := events[0].Object.(*unstructured.Unstructured)
		watchEventHeader(w, events[0].Type, o)

		ready := pods.GetReady(o)
		sort.Strings(ready)
		unready := pods.GetUnready(o)
		sort.Strings(unready)
		pods := append(ready, unready...)

		if len(unready) > 0 {
			li := prefix + strings.Join(pods, prefix)
			FailureStatusEvent(w, fmt.Sprintf("Directs traffic to the following live Pods:%s", li))
		} else if len(ready) > 0 {
			li := prefix + strings.Join(pods, prefix)
			SuccessStatusEvent(w, fmt.Sprintf("Directs traffic to the following live Pods:%s", li))
		} else if len(unready) == 0 && len(ready) == 0 {
			FailureStatusEvent(w, fmt.Sprintf("Does not direct traffic to any Pods"))
		}

	} else if svcType != "ExternalName" {
		fmt.Fprintln(w, "❌ Waiting for live Pods to be targeted by service")
	}

	w.Flush()
}

func watchEventHeader(w io.Writer, eventType k8sWatch.EventType, o *unstructured.Unstructured) {
	var eventTypeS string
	if eventType == k8sWatch.Deleted {
		eventTypeS = redBoldText.Sprint(eventType)
	} else {
		eventTypeS = greenText.Sprint(eventType)
	}
	apiType := cyanBoldText.Sprintf("%s/%s", o.GetAPIVersion(), o.GetKind())
	fmt.Fprintf(w, "[%s %s]  %s/%s\n", eventTypeS, apiType, o.GetNamespace(), o.GetName())
}
