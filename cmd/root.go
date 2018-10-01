package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kubespy <command>",
	Short: "Spy on your Kubernetes resources",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func parseObjID(objID string) (namespace, name string, _ error) {
	split := strings.Split(objID, "/")
	if l := len(split); l == 1 {
		return "default", split[0], nil
	} else if l == 2 {
		return split[0], split[1], nil
	}
	return "", "", fmt.Errorf(
		"Object ID must be of the form <name> or <namespace>/<name>, got: %s", objID)
}
