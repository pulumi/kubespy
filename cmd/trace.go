package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/pulumi/kubespy/watch"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sWatch "k8s.io/apimachinery/pkg/watch"
)

const (
	v1Endpoints = "v1/Endpoints"
	v1Service   = "v1/Service"
)

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
		watch.PrintWatchTable(writer, table)
	}

}
