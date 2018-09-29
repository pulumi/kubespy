package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/pulumi/kubespy/watch"
	"github.com/spf13/cobra"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func init() {
	rootCmd.AddCommand(changesCmd)
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
