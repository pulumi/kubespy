package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pulumi/kubespy/k8sconfig"
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
		kubeconfig := k8sconfig.New()
		ns, _, err := kubeconfig.Namespace()
		if err != nil {
			return "", "", err
		}
		return ns, split[0], nil
	} else if l == 2 {
		return split[0], split[1], nil
	}
	return "", "", fmt.Errorf(
		"Object ID must be of the form <name> or <namespace>/<name>, got: %s", objID)
}
