package watch

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/pulumi/pulumi-kubernetes/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	// Load auth plugins. Removing this will likely cause compilation error.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
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
