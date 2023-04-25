package print

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/mbrlabs/uilive"
	"github.com/pulumi/kubespy/k8sobject"
	"github.com/pulumi/kubespy/pods"
	"github.com/pulumi/pulumi-kubernetes/provider/v3/pkg/openapi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sWatch "k8s.io/apimachinery/pkg/watch"
)

const (
	v1Endpoints  = "v1/Endpoints"
	v1Service    = "v1/Service"
	v1Pod        = "v1/Pod"
	deployment   = "Deployment"
	v1ReplicaSet = "v1/ReplicaSet"

	prefix = "\n       - "

	deploymentRevisionKey = "deployment.kubernetes.io/revision"

	trueStatus      = "True"
	statusAvailable = "Available"
)

var (
	greenText      = color.New(color.FgGreen)
	faintGreenText = color.New(color.Faint, color.FgGreen)
	yellowText     = color.New(color.FgYellow)
	cyanBoldText   = color.New(color.FgCyan, color.Bold)
	cyanText       = color.New(color.FgCyan)
	redBoldText    = color.New(color.FgRed, color.Bold)
	whiteBoldText  = color.New(color.Bold)
	yellowBoldText = color.New(color.FgYellow, color.Bold)
	faintText      = color.New(color.Faint)
)

// SuccessStatusEvent prints a message using the formatting of success status.
func SuccessStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	fmt.Fprintf(w, "    ✅ %s\n", fmt.Sprintf(fmtstr, a...))
}

// FailureStatusEvent prints a message using the formatting of a failure status.
func FailureStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	fmt.Fprintf(w, "    ❌ %s\n", fmt.Sprintf(fmtstr, a...))
}

// PendingStatusEvent prints a message using the formatting of a pending status.
func PendingStatusEvent(w io.Writer, fmtstr string, a ...interface{}) {
	fmt.Fprintf(w, "    ⌛ %s\n", fmt.Sprintf(fmtstr, a...))
}

// ServiceWatchTable prints the status of a Service, as represented in a table.
func ServiceWatchTable(w *uilive.Writer, table map[string][]k8sWatch.Event) {
	var svcType string
	if events, hasSvc := table[v1Service]; hasSvc {
		o := events[0].Object.(*unstructured.Unstructured)
		watchEventHeader(w, events[0].Type, o)

		svcTypeI, _ := openapi.Pluck(o.Object, "spec", "type")
		var isString bool
		svcType, isString = svcTypeI.(string)
		if !isString {
			svcType = "ClusterIP"
		}

		switch svcType {
		case "ClusterIP":
			if eps, hasEndpoints := table[v1Endpoints]; hasEndpoints && eps[0].Type != k8sWatch.Deleted {
				SuccessStatusEvent(w, "Successfully created Endpoints object '%s' to direct traffic to Pods",
					cyanText.Sprint(o.GetName()))
			} else {
				FailureStatusEvent(w, "Waiting for Endpoints object to be created, to direct traffic to Pods")
			}

			clusterIPI, _ := openapi.Pluck(o.Object, "spec", "clusterIP")
			if clusterIP, isString := clusterIPI.(string); isString && len(clusterIP) > 0 {
				SuccessStatusEvent(w, "Successfully allocated a cluster-internal IP: %s",
					cyanText.Sprint(o.GetName()))
			} else {
				FailureStatusEvent(w, "Waiting for cluster-internal IP to be allocated")
			}
		case "LoadBalancer":
			if eps, hasEndpoints := table[v1Endpoints]; hasEndpoints && eps[0].Type != k8sWatch.Deleted {
				SuccessStatusEvent(w, "Successfully created Endpoints object '%s' to direct traffic to Pods", o.GetName())
			} else {
				FailureStatusEvent(w, "Waiting for Endpoints object to be created, to direct traffic to Pods")
			}

			ingressesI, _ := openapi.Pluck(o.Object, "status", "loadBalancer", "ingress")
			if ingresses, isMap := ingressesI.([]interface{}); isMap {
				ips := []string{}
				for _, ingressI := range ingresses {
					if ingress, isMap := ingressI.(map[string]interface{}); isMap {
						tmp := []string{}
						ipI, _ := ingress["ip"]
						ip, isString := ipI.(string)
						if isString {
							tmp = append(tmp, cyanText.Sprint(ip))
						}

						hostnameI, _ := ingress["hostname"]
						hostname, isString := hostnameI.(string)
						if isString {
							tmp = append(tmp, cyanText.Sprint(hostname))
						}

						ips = append(ips, strings.Join(tmp, "/"))
					}
				}

				if len(ips) > 0 {
					sort.Strings(ips)
					li := prefix + strings.Join(ips, prefix)
					SuccessStatusEvent(w, "Service allocated the following IPs/hostnames:%s", li)
				}
			} else {
				FailureStatusEvent(w, "Waiting for public IP/host to be allocated")
			}

		case "ExternalName":
			externalNameI, _ := openapi.Pluck(o.Object, "spec", "externalName")
			if externalName, isString := externalNameI.(string); isString && len(externalName) > 0 {
				SuccessStatusEvent(w, "Service proxying to %s", cyanText.Sprint(externalName))
			} else {
				FailureStatusEvent(w, "Service not given a URI to proxy to in `.spec.externalName`")
			}
		}
	}

	fmt.Fprintln(w)

	if events, hasEPs := table[v1Endpoints]; hasEPs {
		o := events[0].Object.(*unstructured.Unstructured)
		watchEventHeader(w, events[0].Type, o)

		ready := pods.GetReady(o)
		sort.Strings(ready)
		unready := pods.GetUnready(o)
		sort.Strings(unready)
		pods := append(ready, unready...)

		if len(unready) > 0 {
			li := prefix + strings.Join(pods, prefix)
			FailureStatusEvent(w, "Directs traffic to the following live Pods:%s", li)
		} else if len(ready) > 0 {
			li := prefix + strings.Join(pods, prefix)
			SuccessStatusEvent(w, "Directs traffic to the following live Pods:%s", li)
		} else if len(unready) == 0 && len(ready) == 0 {
			FailureStatusEvent(w, "Does not direct traffic to any Pods")
		}

	} else if svcType != "ExternalName" {
		fmt.Fprintln(w, "❌ Waiting for live Pods to be targeted by service")
	}

	w.Flush()
}

// DeploymentWatchTable prints the status of a Deployment, as represented in a table.
func DeploymentWatchTable(w *uilive.Writer, table map[string][]k8sWatch.Event) {
	const (
		waitingForControllerCreate = "Waiting for controller to create Deployment"
		rolloutNotStarted          = "Deployment has not begun to roll out the change"
		appUnavailable             = "Deployment does not have minimum replicas (%d out of %d)"
	)

	var currentRevision int
	if events, hasDepl := table[deployment]; hasDepl {
		o := events[0].Object.(*unstructured.Unstructured)
		var err error
		currentRevision, err = parseRevision(o)
		if err != nil {
			FailureStatusEvent(w, waitingForControllerCreate)
		}
	}

	// Get current ReplicaSet.
	newReplicaSetAvailable := true
	var currRepSet *unstructured.Unstructured
	var currRepSetEventType k8sWatch.EventType
	type repSet struct {
		eventType k8sWatch.EventType
		revision  int
		replicas  int64
		namespace string
		name      string
	}
	var prevRepSet *unstructured.Unstructured
	var prevRepSetSpec *repSet
	if events, hasRs := table[v1ReplicaSet]; hasRs {
		for _, e := range events {
			rs := e.Object.(*unstructured.Unstructured)
			repSetRevision, err := parseRevision(rs)
			if err != nil {
				continue
			}
			if currentRevision == repSetRevision {
				currRepSet = rs
				currRepSetEventType = e.Type
			} else if e.Type != k8sWatch.Deleted {
				replicasI, _ := openapi.Pluck(rs.Object, "status", "replicas")
				replicas, _ := replicasI.(int64)

				revision, err := parseRevision(rs)
				if replicas > 0 && err == nil {
					prevRepSet = rs
					prevRepSetSpec = &repSet{
						eventType: e.Type,
						revision:  revision,
						replicas:  replicas,
						namespace: rs.GetNamespace(),
						name:      rs.GetName()}
				}
			}
		}
	}

	// Display `Deployment` status.
	if events, hasDepl := table[deployment]; hasDepl {
		o := events[0].Object.(*unstructured.Unstructured)
		watchEventHeader(w, events[0].Type, o)

		specReplicasI, _ := openapi.Pluck(o.Object, "spec", "replicas")
		specReplicas, isInt := specReplicasI.(int)
		if !isInt {
			specReplicas = 1
		}

		availableReplicasI, _ := openapi.Pluck(o.Object, "status", "availableReplicas")
		availableReplicas, isInt := availableReplicasI.(int)
		if !isInt {
			specReplicas = 0
		}

		// Check Deployments conditions to see whether new ReplicaSet is available. If it is, we are
		// successful.
		conditionsI, _ := openapi.Pluck(o.Object, "status", "conditions")
		conditions, isSlice := conditionsI.([]interface{})
		if !isSlice {
			FailureStatusEvent(w, appUnavailable, 0, specReplicas)
			FailureStatusEvent(w, rolloutNotStarted)
		} else {
			whiteBoldText.Fprintf(w, "    Rolling out Deployment revision %d\n", currentRevision)

			var isProgressing bool
			var progressingReason string

			var deploymentAvailable bool
			var availableReason string

			var rolloutSuccessful bool

			// Success occurs when the ReplicaSet of the `currentGeneration` is marked as available, and
			// when the deployment is available.
			for _, rawCondition := range conditions {
				condition, isMap := rawCondition.(map[string]interface{})
				if !isMap {
					continue
				}

				if condition["type"] == "Progressing" {
					isProgressing = condition["status"] == trueStatus

					reasonI, _ := condition["reason"]
					reason, isString := reasonI.(string)
					if !isString {
						continue
					}
					messageI, _ := condition["message"]
					message, isString := messageI.(string)
					if !isString {
						continue
					}
					progressingReason = fmt.Sprintf("[%s] %s", reason, message)

					rolloutSuccessful = condition["reason"] == "NewReplicaSetAvailable" && newReplicaSetAvailable
				}

				if condition["type"] == statusAvailable {
					deploymentAvailable = condition["status"] == trueStatus
					reasonI, _ := condition["reason"]
					reason, isString := reasonI.(string)
					if !isString {
						continue
					}
					messageI, _ := condition["message"]
					message, isString := messageI.(string)
					if !isString {
						continue
					}
					availableReason = fmt.Sprintf("[%s] %s", reason, message)
				}
			}

			if !deploymentAvailable {
				FailureStatusEvent(w, "Deployment is failing; %d out of %d Pods are available: %s",
					availableReplicas, specReplicas, availableReason)
			} else {
				SuccessStatusEvent(w, "Deployment is currently available")
			}

			if !isProgressing {
				FailureStatusEvent(w, "Rollout has failed; controller is no longer rolling forward: %s",
					progressingReason)
			} else if rolloutSuccessful {
				SuccessStatusEvent(w, "Rollout successful: new ReplicaSet marked 'available'")
			} else {
				PendingStatusEvent(w, "Rollout proceeding: %s", progressingReason)
			}
		}
	}

	fmt.Fprintln(w)

	// Display `ReplicaSet` status.
	if currRepSet != nil {
		cyanBoldText.Fprintln(w, "ROLLOUT STATUS:")
		fmt.Fprintf(w, "- [%s | Revision %d] ", yellowBoldText.Sprint("Current rollout"), currentRevision)

		specReplicasI, _ := openapi.Pluck(currRepSet.Object, "spec", "replicas")
		specReplicas, isInt := specReplicasI.(int64)
		if !isInt {
			specReplicas = 1
		}

		availableReplicasI, _ := openapi.Pluck(currRepSet.Object, "status", "availableReplicas")
		availableReplicas, isInt := availableReplicasI.(int64)
		if !isInt {
			availableReplicas = 0
		}

		currReplicaSetStatus(w, currRepSetEventType, currentRevision, currRepSet)
		if availableReplicas < specReplicas {
			PendingStatusEvent(w,
				"Waiting for ReplicaSet to attain minimum available Pods (%d available of a %d minimum)",
				availableReplicas, specReplicas)
		} else {
			SuccessStatusEvent(w,
				"ReplicaSet is available [%d Pods available of a %d minimum]",
				availableReplicas, specReplicas)
		}

		printPodStatus(w,
			func(w io.Writer, f string, a ...interface{}) { fmt.Fprintf(w, f, a...) }, currRepSet, table)
	} else {
		fmt.Fprintln(w, "⌛ Waiting for Deployment controller to create ReplicaSet")
	}

	if prevRepSetSpec != nil {
		fmt.Fprintln(w)

		faintText.Fprintf(w, "- [%s", whiteBoldText.Sprint("Previous ReplicaSet"))
		faintText.Fprintf(w, " | Revision %d] [%s", prevRepSetSpec.revision,
			greenText.Sprint(prevRepSetSpec.eventType))
		faintText.Fprintf(w, "]  %s/%s\n", prevRepSetSpec.namespace, prevRepSetSpec.name)
		faintText.Fprintf(w, "    ⌛ Waiting for ReplicaSet to scale to 0 Pods (%d currently exist)\n",
			prevRepSetSpec.replicas)

		printPodStatus(w, faintText.FprintfFunc(), prevRepSet, table)

	}

	w.Flush()
}

func printPodStatus(w io.Writer, fprintf func(w io.Writer, f string, a ...interface{}),
	rs *unstructured.Unstructured, table map[string][]k8sWatch.Event) {
	if podEvents, exists := table[v1Pod]; exists {
		for _, e := range podEvents {
			pod := e.Object.(*unstructured.Unstructured)
			if !k8sobject.OwnedBy(pod, rs.GetAPIVersion(), rs.GetKind(), rs.GetName()) {
				continue
			}

			conditions := k8sobject.PodConditions(pod)
			for _, conditionI := range conditions {
				condition, isMap := conditionI.(map[string]interface{})
				if !isMap {
					continue
				}

				if condition["type"] == "PodScheduled" {
					if condition["status"] != trueStatus {
						reason, message := errorFromCondition(condition)
						printPodContainerError(w, fprintf, pod, reason, message)
					}
				}

				if condition["type"] == "Initialized" {
					if condition["status"] != trueStatus {
						reason, message := errorFromCondition(condition)
						printPodContainerError(w, fprintf, pod, reason, message)
					}
				}

				if condition["type"] == "Ready" {
					if condition["status"] != trueStatus {
						reason, message := errorFromCondition(condition)
						printPodContainerError(w, fprintf, pod, reason, message)
					} else {
						fprintf(w, "       - [%s", greenText.Sprint("Ready"))
						fprintf(w, "] %s\n", cyanText.Sprint(pod.GetName()))
					}
				}
			}

			// Collect the errors from any containers that are failing.
			containerStatuses := k8sobject.PodContainerStatuses(pod)
			for _, rawContainerStatus := range containerStatuses {
				containerStatus, isMap := rawContainerStatus.(map[string]interface{})
				if !isMap || containerStatus["ready"] == true {
					continue
				}

				// Process container that's waiting.
				rawWaiting, isWaiting := openapi.Pluck(containerStatus, "state", "waiting")
				waiting, isMap := rawWaiting.(map[string]interface{})
				if isWaiting && rawWaiting != nil && isMap {
					reason, message := checkWaitingContainer(waiting)
					printPodContainerError(w, fprintf, pod, reason, message)
				}

				// Process container that's terminated.
				rawTerminated, isTerminated := openapi.Pluck(containerStatus, "state", "terminated")
				terminated, isMap := rawTerminated.(map[string]interface{})
				if isTerminated && rawTerminated != nil && isMap {
					reason, message := checkTerminatedContainer(terminated)
					printPodContainerError(w, fprintf, pod, reason, message)
				}
			}

			// Exhausted our knowledge of possible error states for Pods. Return.
		}
	}
}

func printPodContainerError(w io.Writer, fprintf func(w io.Writer, f string, a ...interface{}),
	pod *unstructured.Unstructured, reason,
	message string) {
	if reason == "" || message == "" {
		return
	}
	fprintf(w, "       - [%s", redBoldText.Sprint(reason))
	fprintf(w, "] %s", cyanText.Sprint(pod.GetName()))
	fprintf(w, " %s\n", message)
}

func errorFromCondition(condition map[string]interface{}) (string, string) {
	reasonI, _ := condition["reason"]
	reason, isString := reasonI.(string)
	if !isString {
		return "", ""
	}
	messageI, _ := condition["message"]
	message, isString := messageI.(string)
	if !isString {
		return "", ""
	}

	return reason, message
}

func checkWaitingContainer(waiting map[string]interface{}) (string, string) {
	rawReason, hasReason := waiting["reason"]
	reason, isString := rawReason.(string)
	if !hasReason || !isString || reason == "" || reason == "ContainerCreating" {
		return "", ""
	}

	rawMessage, hasMessage := waiting["message"]
	message, isString := rawMessage.(string)
	if !hasMessage || !isString {
		return "", ""
	}

	// Image pull error has a bunch of useless junk at the beginning of the error message. Try to
	// remove it.
	imagePullJunk := "rpc error: code = Unknown desc = Error response from daemon: "
	message = strings.TrimPrefix(message, imagePullJunk)

	return reason, message
}

func checkTerminatedContainer(terminated map[string]interface{}) (string, string) {
	reasonI, _ := terminated["reason"]
	reason, isString := reasonI.(string)
	if !isString || reason == "" {
		return "", ""
	}

	messageI, _ := terminated["message"]
	message, isString := messageI.(string)
	if !isString {
		message = fmt.Sprintf("Container completed with exit code %d", terminated["exitCode"])
	}

	return reason, message
}

func watchEventHeader(w io.Writer, eventType k8sWatch.EventType, o *unstructured.Unstructured) {
	var eventTypeS string
	if eventType == k8sWatch.Deleted {
		eventTypeS = redBoldText.Sprint(eventType)
	} else {
		eventTypeS = greenText.Sprint(eventType)
	}
	apiType := cyanBoldText.Sprintf("%s/%s", o.GetAPIVersion(), o.GetKind())
	fmt.Fprintf(w, "[%s %s]  %s/%s\n", eventTypeS, apiType, o.GetNamespace(), o.GetName())
}

func currReplicaSetStatus(
	w io.Writer, eventType k8sWatch.EventType, revision int, o *unstructured.Unstructured,
) {
	var eventTypeS string
	if eventType == k8sWatch.Deleted {
		eventTypeS = redBoldText.Sprint(eventType)
	} else {
		eventTypeS = greenText.Sprint(eventType)
	}
	fmt.Fprintf(w, "[%s]  %s/%s\n", eventTypeS, o.GetNamespace(), o.GetName())
}

func parseRevision(o *unstructured.Unstructured) (int, error) {
	revisionI, _ := openapi.Pluck(o.Object, "metadata", "annotations", deploymentRevisionKey)
	revisionS, _ := revisionI.(string)
	return strconv.Atoi(revisionS)
}
