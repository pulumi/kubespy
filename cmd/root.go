package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/pulumi/pulumi-kubernetes/pkg/openapi"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubespy <command>",
	Short: "Spy on your Kubernetes resources",
}

func init() {
	// Turn off timestamp prefix for `log.Fatal*`.
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	_, _ = openapi.Pluck(map[string]interface{}{}, "spec", "type")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
