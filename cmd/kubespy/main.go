package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/pulumi/kubespy/watch"
	"github.com/pulumi/pulumi-kubernetes/pkg/openapi"
	"github.com/spf13/cobra"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sWatch "k8s.io/apimachinery/pkg/watch"
)

func main() {
	rootCmd.Execute()
}

func init() {
	// Turn off timestamp prefix for `log.Fatal*`.
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	_, _ = openapi.Pluck(map[string]interface{}{}, "spec", "type")

	rootCmd.AddCommand(changesCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(traceCmd)
}

var rootCmd = &cobra.Command{
	Use:   "kubespy <command>",
	Short: "Spy on your Kubernetes resources",
}

var changesCmd = &cobra.Command{
	Use:   "changes <apiVersion> <kind> [<namespace>/]<name>",
	Short: "Displays changes made to a Kubernetes resource in real time. Emitted as JSON diffs",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		events, err := watch.Forever(args[0], args[1], args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(color.GreenString("Watching for changes on %s %s %s", args[0], args[1], args[2]))

		heading := color.New(color.FgBlue, color.Bold)

		var last *unstructured.Unstructured
		for {
			select {
			case e := <-events:
				o := e.Object.(*unstructured.Unstructured)
				if last == nil {
					heading.Println("CREATED")

					ojson, err := json.MarshalIndent(o.Object, "", "  ")
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(color.GreenString(string(ojson)))
				} else {
					heading.Println(string(e.Type))

					diff := gojsondiff.New().CompareObjects(last.Object, o.Object)
					if diff.Modified() {
						fcfg := formatter.AsciiFormatterConfig{Coloring: true}
						formatter := formatter.NewAsciiFormatter(last.Object, fcfg)
						text, err := formatter.Format(diff)
						if err != nil {
							log.Fatal(err)
						}
						fmt.Println(text)
					}
				}
				last = o
			}
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status <apiVersion> <kind> [<namespace>/]<name>",
	Short: "Displays changes to a Kubernetes resources's status in real time. Emitted as JSON diffs",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		events, err := watch.Forever(args[0], args[1], args[2])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(color.GreenString("Watching status of %s %s %s", args[0], args[1], args[2]))

		heading := color.New(color.FgBlue, color.Bold)

		var lastStatus map[string]interface{}
		for {
			select {
			case e := <-events:
				o := e.Object.(*unstructured.Unstructured)
				var currStatus map[string]interface{}
				if status, hasStatus := o.Object["status"]; !hasStatus {
					currStatus = make(map[string]interface{})
				} else if status, isMap := status.(map[string]interface{}); !isMap {
					currStatus = make(map[string]interface{})
				} else {
					currStatus = status
				}

				if lastStatus == nil {
					heading.Println("CREATED")

					ojson, err := json.MarshalIndent(currStatus, "", "  ")
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(color.GreenString(string(ojson)))
				} else {
					heading.Println(string(e.Type))

					diff := gojsondiff.New().CompareObjects(lastStatus, currStatus)
					if diff.Modified() {
						fcfg := formatter.AsciiFormatterConfig{Coloring: true}
						formatter := formatter.NewAsciiFormatter(lastStatus, fcfg)
						text, err := formatter.Format(diff)
						if err != nil {
							log.Fatal(err)
						}
						fmt.Println(text)
					}
				}
				lastStatus = currStatus
			}
		}
	},
}

var traceCmd = &cobra.Command{
	Use:   "trace <type> [<namespace>/]<name>",
	Short: "Traces status of complex API objects",
	Long: `Traces status of complex API objects. Accepted types are:
  - service (aliases: {service, svc})`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		switch t := strings.ToLower(args[0]); t {
		case "service", "svc":
			traceService(args[1])
		default:
			msg := "Unknown resource type '%s'. The following resources are available:\n" +
				"  - service (aliases: {service, svc})"
			log.Fatalf(msg, t)
		}
	},
}

const (
	v1Endpoints = "v1/Endpoints"
	v1Service   = "v1/Service"

	prefix = "\n       - "
)

var (
	greenText    = color.New(color.FgGreen)
	yellowText   = color.New(color.FgYellow)
	cyanBoldText = color.New(color.FgCyan, color.Bold)
	cyanText     = color.New(color.FgCyan)
	redBoldText  = color.New(color.FgRed, color.Bold)
)

func traceService(objID string) {
	serviceEvents, err := watch.Forever("v1", "Service", objID)
	if err != nil {
		log.Fatal(err)
	}

	endpointEvents, err := watch.Forever("v1", "Endpoints", objID)
	if err != nil {
		log.Fatal(err)
	}

	writer := uilive.New()
	writer.RefreshInterval = time.Minute * 1
	writer.Start()      // Start listening for updates, render.
	defer writer.Stop() // Flush buffers, stop rendering.

	// Initial message.
	fmt.Fprintln(writer, color.New(color.FgCyan, color.Bold).Sprintf("Waiting for Service '%s'", objID))
	writer.Flush()

	table := map[string][]k8sWatch.Event{}

	for {
		select {
		case e := <-serviceEvents:
			if e.Type == k8sWatch.Deleted {
				o := e.Object.(*unstructured.Unstructured)
				delete(o.Object, "spec")
				delete(o.Object, "status")
			}
			table[v1Service] = []k8sWatch.Event{e}
		case e := <-endpointEvents:
			if e.Type == k8sWatch.Deleted {
				o := e.Object.(*unstructured.Unstructured)
				delete(o.Object, "spec")
				delete(o.Object, "status")
				delete(o.Object, "subsets")
			}
			table[v1Endpoints] = []k8sWatch.Event{e}
		}
		printWatchTable(writer, table)
	}

}

func printWatchTable(w *uilive.Writer, table map[string][]k8sWatch.Event) {
	var svcType string
	if events, hasSvc := table[v1Service]; hasSvc {
		o := events[0].Object.(*unstructured.Unstructured)
		printWatchEventHeader(w, events[0].Type, o)

		svcTypeI, _ := openapi.Pluck(o.Object, "spec", "type")
		var isString bool
		svcType, isString = svcTypeI.(string)
		if !isString {
			svcType = "ClusterIP"
		}

		switch svcType {
		case "ClusterIP":
			if eps, hasEndpoints := table[v1Endpoints]; hasEndpoints && eps[0].Type != k8sWatch.Deleted {
				printSuccessStatusEvent(w,
					fmt.Sprintf("Successfully created Endpoints object '%s' to direct traffic to Pods",
						cyanText.Sprint(o.GetName())))
			} else {
				printFailureStatusEvent(w, "Waiting for Endpoints object to be created, to direct traffic to Pods")
			}

			clusterIPI, _ := openapi.Pluck(o.Object, "spec", "clusterIP")
			if clusterIP, isString := clusterIPI.(string); isString && len(clusterIP) > 0 {
				printSuccessStatusEvent(w,
					fmt.Sprintf("Successfully allocated a cluster-internal IP: %s",
						cyanText.Sprint(o.GetName())))
			} else {
				printFailureStatusEvent(w, "Waiting for cluster-internal IP to be allocated")
			}
		case "LoadBalancer":
			if eps, hasEndpoints := table[v1Endpoints]; hasEndpoints && eps[0].Type != k8sWatch.Deleted {
				printSuccessStatusEvent(w,
					fmt.Sprintf("Successfully created Endpoints object '%s' to direct traffic to Pods", o.GetName()))
			} else {
				printFailureStatusEvent(w, "Waiting for Endpoints object to be created, to direct traffic to Pods")
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
					printSuccessStatusEvent(w, fmt.Sprintf("Service allocated the following IPs/hostnames:%s", li))
				}
			} else {
				printFailureStatusEvent(w, "Waiting for public IP/host to be allocated")
			}

		case "ExternalName":
			externalNameI, _ := openapi.Pluck(o.Object, "spec", "externalName")
			if externalName, isString := externalNameI.(string); isString && len(externalName) > 0 {
				printSuccessStatusEvent(w, fmt.Sprintf("Service proxying to %s", cyanText.Sprint(externalName)))
			} else {
				printFailureStatusEvent(w, "Service not given a URI to proxy to in `.spec.externalName`")
			}
		}
	}

	fmt.Fprintln(w)

	if events, hasEPs := table[v1Endpoints]; hasEPs {
		o := events[0].Object.(*unstructured.Unstructured)
		printWatchEventHeader(w, events[0].Type, o)

		ready := getReadyPods(o)
		sort.Strings(ready)
		unready := getUnreadyPods(o)
		sort.Strings(unready)
		pods := append(ready, unready...)

		if len(unready) > 0 {
			li := prefix + strings.Join(pods, prefix)
			printFailureStatusEvent(w, fmt.Sprintf("Directs traffic to the following live Pods:%s", li))
		} else if len(ready) > 0 {
			li := prefix + strings.Join(pods, prefix)
			printSuccessStatusEvent(w, fmt.Sprintf("Directs traffic to the following live Pods:%s", li))
		} else if len(unready) == 0 && len(ready) == 0 {
			printFailureStatusEvent(w, fmt.Sprintf("Does not direct traffic to any Pods"))
		}

	} else if svcType != "ExternalName" {
		fmt.Fprintln(w, "❌ Waiting for live Pods to be targeted by service")
	}

	w.Flush()
}

func getReadyPods(o *unstructured.Unstructured) []string {
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

func getUnreadyPods(o *unstructured.Unstructured) []string {
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
func printWatchEventHeader(w io.Writer, eventType k8sWatch.EventType, o *unstructured.Unstructured) {
	var eventTypeS string
	if eventType == k8sWatch.Deleted {
		eventTypeS = redBoldText.Sprint(eventType)
	} else {
		eventTypeS = greenText.Sprint(eventType)
	}
	apiType := cyanBoldText.Sprintf("%s/%s", o.GetAPIVersion(), o.GetKind())
	fmt.Fprintf(w, "[%s %s]  %s/%s\n", eventTypeS, apiType, o.GetNamespace(), o.GetName())
}

func printSuccessStatusEvent(w io.Writer, message string) {
	fmt.Fprintf(w, "    ✅ %s\n", message)
}
func printFailureStatusEvent(w io.Writer, message string) {
	fmt.Fprintf(w, "    ❌ %s\n", message)
}
