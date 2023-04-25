package k8sconfig

import (
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

// New creates a ClientConfig for kubernetes
func New() clientcmd.ClientConfig {
	// Use client-go to resolve the final configuration values for the client. Typically these
	// values would would reside in the $KUBECONFIG file, but can also be altered in several
	// places, including in env variables, client-go default values, and (if we allowed it) CLI
	// flags.
	overrides := &clientcmd.ConfigOverrides{}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, overrides, os.Stdin)
}
