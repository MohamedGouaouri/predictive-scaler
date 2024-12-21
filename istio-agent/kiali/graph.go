package kiali

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

func GetWorkloadGraph(url string) WorkloadGraph {

	// Make the GET request
	response, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to make GET request: %v", err)
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var schema WorkloadGraph
	err = json.Unmarshal([]byte(body), &schema)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	return schema
}

func InjectFeatures(graph WorkloadGraph, k8sClient *kubernetes.Clientset, metricsClient *metricsclientset.Clientset) WorkloadGraph {

	for i := range graph.Elements.Nodes {
		node := &graph.Elements.Nodes[i]
		ns := node.Data.Namespace
		wl := node.Data.Workload
		dep, err := k8sClient.AppsV1().Deployments(ns).Get(context.TODO(), wl, v1.GetOptions{})

		if err != nil {
			fmt.Println(err)
			continue
		}
		labelSelector := metav1.FormatLabelSelector(dep.Spec.Selector)
		pods, err := k8sClient.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			fmt.Println(err)
			continue
		}
		var avgDeployCpuUsage float64 = 0.0
		var avgDeployMemUsage float64 = 0.0
		for _, pod := range pods.Items {
			metrics, err := metricsClient.MetricsV1beta1().PodMetricses(ns).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Println(err)
				continue
			}
			// scheduledOn, err := k8sClient.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
			// if err != nil {
			// 	fmt.Println(err)
			// 	continue
			// }
			var avgPodCpu float64 = 0.0
			var avgPodMem float64 = 0.0
			for _, containerMetrics := range metrics.Containers {
				avgPodCpu += float64(containerMetrics.Usage.Cpu().MilliValue())
				avgPodMem += float64(containerMetrics.Usage.Memory().MilliValue())
			}

			// Normalize avgPodCpu and avgPodMem
			// avgPodCpuUsage := float64(avgPodCpu * 100 / float64(scheduledOn.Status.Allocatable.Cpu().MilliValue()))
			// avgPodMemUsage := float64(avgPodMem * 100 / float64(scheduledOn.Status.Allocatable.Memory().MilliValue()))

			avgDeployCpuUsage += avgPodCpu
			avgDeployMemUsage += avgPodMem

		}
		// avgDeployCpuUsage /= float64(len(pods.Items))
		// avgDeployMemUsage /= float64(len(pods.Items))
		node.Data.Cpu = avgDeployCpuUsage
		node.Data.Mem = avgDeployMemUsage
		node.Data.Replicas = int(*dep.Spec.Replicas)
	}
	return graph
}

func (graph *WorkloadGraph) Convert() *ReshapedNode {
	gatewayNode := &ReshapedNode{}
	nodesMap := make(map[string]*ReshapedNode)
	for _, node := range graph.Elements.Nodes {
		if node.Data.IsGateway != nil {
			// This is a root node
			gatewayNode.ID = node.Data.ID
			gatewayNode.Children = make([]*ReshapedNode, 0)
			gatewayNode.Edges = make([]ReshapedEdge, 0)
			gatewayNode.IsGateway = true
			gatewayNode.Workload = node.Data.Workload
			gatewayNode.Cpu = node.Data.Cpu
			gatewayNode.Mem = node.Data.Mem
			gatewayNode.Replicas = node.Data.Replicas
			nodesMap[node.Data.ID] = gatewayNode
			continue
		}
		nodesMap[node.Data.ID] = &ReshapedNode{
			ID:        node.Data.ID,
			Children:  make([]*ReshapedNode, 0),
			Edges:     make([]ReshapedEdge, 0),
			IsGateway: false,
			Workload:  node.Data.Workload,
			Cpu:       node.Data.Cpu,
			Mem:       node.Data.Mem,
			Replicas:  node.Data.Replicas,
		}
	}

	// Iterate over the graph edges to create new reshaped edges and link nodes
	for _, edge := range graph.Elements.Edges {
		sourceNode := nodesMap[edge.Data.Source]
		targetNode := nodesMap[edge.Data.Target]
		if sourceNode != nil && targetNode != nil {
			rate := 0.0
			if edge.Data.Traffic.Protocol == "http" {
				rate = parseFloat(edge.Data.Traffic.Rates["http"])
			} else if edge.Data.Traffic.Protocol == "grpc" {
				rate = parseFloat(edge.Data.Traffic.Rates["grpc"])
			}
			// TODO: Handle raw tcp
			newEdge := ReshapedEdge{
				ResponseTime: parseFloat(edge.Data.ResponseTime),
				RequestRate:  rate,
			}
			sourceNode.Edges = append(sourceNode.Edges, newEdge)
			sourceNode.Children = append(sourceNode.Children, targetNode)
		}
	}

	return gatewayNode
}

func (rn *ReshapedNode) toJSON() string {
	jsonData, err := json.Marshal(rn)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return ""
	}
	return string(jsonData)
}

// Helper function to parse string to float64
func parseFloat(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return value
}

func calculateCPUUsageForPod(metricsClient *metricsclientset.Clientset, pod *coreV1.Pod) float64 {
	ctx := context.TODO()

	// Get metrics for the specified pod
	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(pod.Namespace).Get(ctx, pod.Name, v1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting metrics for pod %s: %v\n", pod.Name, err)
		return 0.0
	}

	// Sum CPU usage across all containers in the pod
	var totalCPUUsageMillicores int64 = 0
	for _, container := range podMetrics.Containers {
		cpuUsage := container.Usage[coreV1.ResourceCPU]
		totalCPUUsageMillicores += cpuUsage.MilliValue()
	}

	// Sum CPU requests for all containers in the pod
	var totalCPURequestsMillicores int64 = 0
	for _, container := range pod.Spec.Containers {
		if request, ok := container.Resources.Requests[coreV1.ResourceCPU]; ok {
			totalCPURequestsMillicores += request.MilliValue()
		}
	}

	// Calculate CPU usage percentage
	if totalCPURequestsMillicores == 0 {
		fmt.Printf("Pod %s has zero CPU requests.\n", pod.Name)
		return 0.0
	}

	cpuUsagePercentage := (float64(totalCPUUsageMillicores) / float64(totalCPURequestsMillicores)) * 100.0
	return cpuUsagePercentage
}

func calculateCPUUsageForNode(k8sClient *kubernetes.Clientset, metricsClient *metricsclientset.Clientset, node *coreV1.Node) float64 {
	ctx := context.TODO()

	// Get metrics for the specified node
	nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, v1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting metrics for node %s: %v\n", node.Name, err)
		return 0.0
	}

	// Extract CPU usage in millicores from node metrics
	cpuUsage := nodeMetrics.Usage[coreV1.ResourceCPU]
	cpuUsageMillicores := cpuUsage.MilliValue()

	// Get node capacity for CPU in millicores
	cpuCapacity := node.Status.Capacity[coreV1.ResourceCPU]
	cpuCapacityMillicores := cpuCapacity.MilliValue()

	// Calculate CPU usage percentage
	if cpuCapacityMillicores == 0 {
		fmt.Printf("Node %s has zero CPU capacity.\n", node.Name)
		return 0.0
	}

	cpuUsagePercentage := (float64(cpuUsageMillicores) / float64(cpuCapacityMillicores)) * 100.0
	return cpuUsagePercentage
}

func calculateMemoryUsageForNode(k8sClient *kubernetes.Clientset, metricsClient *metricsclientset.Clientset, node *coreV1.Node) float64 {
	ctx := context.TODO()

	// Get metrics for the specified node
	nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, v1.GetOptions{})
	if err != nil {
		fmt.Printf("Error getting metrics for node %s: %v\n", node.Name, err)
		return 0.0
	}

	// Extract memory usage in bytes from node metrics
	memoryUsage := nodeMetrics.Usage[coreV1.ResourceMemory]
	memoryUsageBytes := memoryUsage.Value()

	// Get node capacity for memory in bytes
	memoryCapacity := node.Status.Capacity[coreV1.ResourceMemory]
	memoryCapacityBytes := memoryCapacity.Value()

	// Calculate memory usage percentage
	if memoryCapacityBytes == 0 {
		fmt.Printf("Node %s has zero memory capacity.\n", node.Name)
		return 0.0
	}

	memoryUsagePercentage := (float64(memoryUsageBytes) / float64(memoryCapacityBytes)) * 100.0
	return memoryUsagePercentage
}
