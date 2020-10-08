package watch

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pulumi/kubespy/k8sobject"
	"github.com/pulumi/pulumi-kubernetes/provider/v2/pkg/clients"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

	// Load auth plugins. Removing this will likely cause compilation error.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	greenText    = color.New(color.FgGreen)
	yellowText   = color.New(color.FgYellow)
	cyanBoldText = color.New(color.FgCyan, color.Bold)
	cyanText     = color.New(color.FgCyan)
	redBoldText  = color.New(color.FgRed, color.Bold)
)

type watchType string

const (
	watchByName  watchType = "watchByName"
	watchByOwner watchType = "watchByOwner"
	watchAll     watchType = "watchAll"
)

// All configures a watch to look for all objects of a type in a namespace.
func All(namespace string) Opts {
	return Opts{watchType: watchAll, namespace: namespace}
}

// ThisObject configures a watch to look for an object specified by a name and a namespace.
func ThisObject(namespace, name string) Opts {
	return Opts{watchType: watchByName, namespace: namespace, name: name}
}

// ObjectsOwnedBy specifies a watch should look for objects owned by `ownerName` in `namespace`.
func ObjectsOwnedBy(namespace, ownerName string) Opts {
	return Opts{watchType: watchByOwner, namespace: namespace, ownerName: ownerName}
}

// Opts specifies which objects to watch for (e.g., "called this" or "owned by x").
type Opts struct {
	watchType watchType

	// (Optional) name of object to watch for.
	name string

	// (Optional) namespace in which to watch for objects.
	namespace string

	// (Optional) ID of object that owns the object we're watching for (e.g., ReplicaSet owned by
	// some Deployment).
	ownerName string
}

func (opts *Opts) Check(o *unstructured.Unstructured) bool {
	switch opts.watchType {
	case watchByName:
		return o.GetName() == opts.name
	case watchByOwner:
		return k8sobject.OwnedBy(o, "extensions/v1beta1", "Deployment", opts.ownerName) ||
			k8sobject.OwnedBy(o, "apps/v1beta1", "Deployment", opts.ownerName) ||
			k8sobject.OwnedBy(o, "apps/v1beta1", "Deployment", opts.ownerName) ||
			k8sobject.OwnedBy(o, "apps/v1", "Deployment", opts.ownerName)
	case watchAll:
		return true
	default:
		panic("Unknown watch type " + opts.watchType)
	}
}

// Forever will watch a resource forever, emitting `watch.Event` until it is killed.
func Forever(apiVersion, kind string, opts Opts) (<-chan watch.Event, error) {
	client, mapper, err := makeClient()
	if err != nil {
		return nil, err
	}

	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}

	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gv.Group, Kind: strings.Title(kind)}, gv.Version)
	if err != nil {
		return nil, err
	}

	clientForResource := client.Resource(mapping.Resource).Namespace(opts.namespace)
	watcher, err := clientForResource.Watch(context.TODO(), metav1.ListOptions{})
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
				if opts.Check(o) {
					out <- e
				}
			}
		}
	}()

	return out, nil
}

func makeClient() (dynamic.Interface, meta.RESTMapper, error) {
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

	client, err := dynamic.NewForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	drm := restmapper.NewDeferredDiscoveryRESTMapper(clients.NewMemCacheClient(discoveryClient))
	return client, drm, nil
}
