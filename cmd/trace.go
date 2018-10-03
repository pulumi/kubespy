package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/pulumi/kubespy/print"
	"github.com/pulumi/kubespy/watch"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sWatch "k8s.io/apimachinery/pkg/watch"
)

const (
	v1Endpoints                 = "v1/Endpoints"
	v1Service                   = "v1/Service"
	v1Pod                       = "v1/Pod"
	deployment                  = "Deployment"
	extensionsV1Beta1ReplicaSet = "extensions/v1beta1/ReplicaSet"
)

func init() {
	rootCmd.AddCommand(traceCmd)
}

var traceCmd = &cobra.Command{
	Use:   "trace <type> [<namespace>/]<name>",
	Short: "Traces status of complex API objects",
	Long: `Traces status of complex API objects. Accepted types are:
  - service (aliases: {svc})
  - deployment (aliases: {deploy})`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		namespace, name, err := parseObjID(args[1])
		if err != nil {
			log.Fatal(err)
		}

		switch t := strings.ToLower(args[0]); t {
		case "service", "svc":
			traceService(namespace, name)
		case "deployment", "deploy":
			traceDeployment(namespace, name)
		default:
			msg := "Unknown resource type '%s'. The following resources are available:\n" +
				"  - service (aliases: {svc})\n" +
				"  - deployment (aliases: {deploy})"
			log.Fatalf(msg, t)
		}
	},
}

func traceService(namespace, name string) {
	serviceEvents, err := watch.Forever("v1", "Service", watch.ThisObject(namespace, name))
	if err != nil {
		log.Fatal(err)
	}

	// NOTE: We can use the same watch opts here because the `Endpoints` object will have the same
	// name and be in the same namespace.
	endpointEvents, err := watch.Forever("v1", "Endpoints", watch.ThisObject(namespace, name))
	if err != nil {
		log.Fatal(err)
	}

	writer := uilive.New()
	writer.RefreshInterval = time.Minute * 1
	writer.Start()      // Start listening for updates, render.
	defer writer.Stop() // Flush buffers, stop rendering.

	// Initial message.
	fmt.Fprintln(writer, color.New(color.FgCyan, color.Bold).Sprintf("Waiting for Service '%s/%s'",
		namespace, name))
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
		print.ServiceWatchTable(writer, table)
	}
}

func traceDeployment(namespace, name string) {
	// API server should rewrite this to apps/v1beta2, apps/v1beta2, or apps/v1 as appropriate.
	deploymentEvents, err := watch.Forever("extensions/v1beta1", "Deployment",
		watch.ThisObject(namespace, name))
	if err != nil {
		log.Fatal(err)
	}

	replicaSetEvents, err := watch.Forever("extensions/v1beta1", "ReplicaSet",
		watch.ObjectsOwnedBy(name))
	if err != nil {
		log.Fatal(err)
	}

	podEvents, err := watch.Forever("v1", "Pod", watch.All(namespace))
	if err != nil {
		log.Fatal(err)
	}

	writer := uilive.New()
	writer.RefreshInterval = time.Minute * 1
	writer.Start()      // Start listening for updates, render.
	defer writer.Stop() // Flush buffers, stop rendering.

	// Initial message.
	fmt.Fprintln(writer, color.New(color.FgCyan, color.Bold).Sprintf("Waiting for Deployment '%s/%s'",
		namespace, name))
	writer.Flush()

	table := map[string][]k8sWatch.Event{} // apiVersion/Kind -> []k8sWatch.Event
	repSets := map[string]k8sWatch.Event{} // Deployment name -> Pod
	pods := map[string]k8sWatch.Event{}    // ReplicaSet name -> Pod

	for {
		select {
		case e := <-deploymentEvents:
			if e.Type == k8sWatch.Deleted {
				o := e.Object.(*unstructured.Unstructured)
				delete(o.Object, "spec")
				delete(o.Object, "status")
			}
			table[deployment] = []k8sWatch.Event{e}
		case e := <-replicaSetEvents:
			o := e.Object.(*unstructured.Unstructured)
			if e.Type == k8sWatch.Deleted {
				delete(repSets, o.GetName())
			} else {
				repSets[o.GetName()] = e
			}
			table[extensionsV1Beta1ReplicaSet] = []k8sWatch.Event{}
			for _, rsEvent := range repSets {
				table[extensionsV1Beta1ReplicaSet] = append(table[extensionsV1Beta1ReplicaSet], rsEvent)
			}
		case e := <-podEvents:
			o := e.Object.(*unstructured.Unstructured)
			if e.Type == k8sWatch.Deleted {
				delete(pods, o.GetName())
			} else {
				pods[o.GetName()] = e
			}

			table[v1Pod] = []k8sWatch.Event{}
			for _, podEvent := range pods {
				table[v1Pod] = append(table[v1Pod], podEvent)
			}
		}
		print.DeploymentWatchTable(writer, table)
	}
}
