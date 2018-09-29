package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubespy <command>",
	Short: "Spy on your Kubernetes resources",
}

func init() {
	// Turn off timestamp prefix for `log.Fatal*`.
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
