package watch

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/pulumi/kubespy/pods"
	"github.com/pulumi/kubespy/print"
	"github.com/pulumi/pulumi-kubernetes/pkg/client"
	"github.com/pulumi/pulumi-kubernetes/pkg/openapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	k8sWatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	// Load auth plugins. Removing this will likely cause compilation error.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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

// Forever will watch a resource forever, emitting `watch.Event` until it is killed.
func Forever(apiVersion, kind, objID string) (<-chan watch.Event, error) {
	disco, pool, err := makeClient()
	if err != nil {
		return nil, err
	}

	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}

	gvk := schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: strings.Title(kind)}

	namespace, name, err := parseObjID(objID)
	if err != nil {
		return nil, err
	}

	clientForResource, err := client.FromGVK(pool, disco, gvk, namespace)
	if err != nil {
		return nil, err
	}

	watcher, err := clientForResource.Watch(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make(chan watch.Event)
	go func() {
		for {
			select {
			case e := <-watcher.ResultChan():
				o, isUnst := e.Object.(*unstructured.Unstructured)
				if !isUnst {
					break
				}
				if o.GetName() == name {
					out <- e
				}
			}
		}
	}()

	return out, nil
}

func PrintWatchTable(w *uilive.Writer, table map[string][]k8sWatch.Event) {
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
				print.SuccessStatusEvent(w,
					fmt.Sprintf("Successfully created Endpoints object '%s' to direct traffic to Pods",
						cyanText.Sprint(o.GetName())))
			} else {
				print.FailureStatusEvent(w, "Waiting for Endpoints object to be created, to direct traffic to Pods")
			}

			clusterIPI, _ := openapi.Pluck(o.Object, "spec", "clusterIP")
			if clusterIP, isString := clusterIPI.(string); isString && len(clusterIP) > 0 {
				print.SuccessStatusEvent(w,
					fmt.Sprintf("Successfully allocated a cluster-internal IP: %s",
						cyanText.Sprint(o.GetName())))
			} else {
				print.FailureStatusEvent(w, "Waiting for cluster-internal IP to be allocated")
			}
		case "LoadBalancer":
			if eps, hasEndpoints := table[v1Endpoints]; hasEndpoints && eps[0].Type != k8sWatch.Deleted {
				print.SuccessStatusEvent(w,
					fmt.Sprintf("Successfully created Endpoints object '%s' to direct traffic to Pods", o.GetName()))
			} else {
				print.FailureStatusEvent(w, "Waiting for Endpoints object to be created, to direct traffic to Pods")
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
					print.SuccessStatusEvent(w, fmt.Sprintf("Service allocated the following IPs/hostnames:%s", li))
				}
			} else {
				print.FailureStatusEvent(w, "Waiting for public IP/host to be allocated")
			}

		case "ExternalName":
			externalNameI, _ := openapi.Pluck(o.Object, "spec", "externalName")
			if externalName, isString := externalNameI.(string); isString && len(externalName) > 0 {
				print.SuccessStatusEvent(w, fmt.Sprintf("Service proxying to %s", cyanText.Sprint(externalName)))
			} else {
				print.FailureStatusEvent(w, "Service not given a URI to proxy to in `.spec.externalName`")
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
			print.FailureStatusEvent(w, fmt.Sprintf("Directs traffic to the following live Pods:%s", li))
		} else if len(ready) > 0 {
			li := prefix + strings.Join(pods, prefix)
			print.SuccessStatusEvent(w, fmt.Sprintf("Directs traffic to the following live Pods:%s", li))
		} else if len(unready) == 0 && len(ready) == 0 {
			print.FailureStatusEvent(w, fmt.Sprintf("Does not direct traffic to any Pods"))
		}

	} else if svcType != "ExternalName" {
		fmt.Fprintln(w, "❌ Waiting for live Pods to be targeted by service")
	}

	w.Flush()
}

func makeClient() (discovery.CachedDiscoveryInterface, dynamic.ClientPool, error) {
	// Use client-go to resolve the final configuration values for the client. Typically these
	// values would would reside in the $KUBECONFIG file, but can also be altered in several
	// places, including in env variables, client-go default values, and (if we allowed it) CLI
	// flags.
	overrides := &clientcmd.ConfigOverrides{}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	kubeconfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, overrides, os.Stdin)

	// Configure the discovery client.
	conf, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to read kubectl config: %v", err)
	}

	disco, err := discovery.NewDiscoveryClientForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	// Cache the discovery information (OpenAPI schema, etc.) so we don't have to retrieve it for
	// every request.
	discoCache := client.NewMemcachedDiscoveryClient(disco)
	mapper := discovery.NewDeferredDiscoveryRESTMapper(discoCache, dynamic.VersionInterfaces)
	pathresolver := dynamic.LegacyAPIPathResolverFunc

	// Create client pool, reusing one client per API group (e.g., one each for core, extensions,
	// apps, etc.)
	pool := dynamic.NewClientPool(conf, mapper, pathresolver)
	return discoCache, pool, nil
}

func parseObjID(objID string) (string, string, error) {
	split := strings.Split(objID, "/")
	if l := len(split); l == 1 {
		return "default", split[0], nil
	} else if l == 2 {
		return split[0], split[1], nil
	}
	return "", "", fmt.Errorf(
		"Object ID must be of the form <name> or <namespace>/<name>, got: %s", objID)
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
