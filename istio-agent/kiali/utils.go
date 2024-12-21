package kiali

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type StringBuilder struct {
	baseString string
}

func NewStringBuilder(baseString string) *StringBuilder {
	return &StringBuilder{
		baseString: baseString,
	}
}

func (sb *StringBuilder) Add(segment string) *StringBuilder {
	sb.baseString += segment
	return sb
}

func (sb *StringBuilder) Build() string {
	return sb.baseString
}

func AddUrlQuery(query, value string) string {
	return fmt.Sprintf("&%s=%s", query, value)
}

func BuildKialiGraphUrl(
	baseUrl string,
	namespace string,
	duration string,
) string {
	sb := NewStringBuilder(baseUrl)
	sb.Add("/kiali/api/namespaces/graph?")
	sb.Add(AddUrlQuery("duration", duration))
	sb.Add(AddUrlQuery("graphType", "workload"))
	sb.Add(AddUrlQuery("includeIdleEdges", "false"))
	sb.Add(AddUrlQuery("injectServiceNodes", "false"))
	sb.Add(AddUrlQuery("responseTime", "avg"))
	sb.Add(AddUrlQuery("appenders", "deadNode,istio,serviceEntry,meshCheck,workloadEntry,health,responseTime"))
	sb.Add(AddUrlQuery("rateGrpc", "requests"))
	sb.Add(AddUrlQuery("rateHttp", "requests"))
	sb.Add(AddUrlQuery("rateTcp", "sent"))
	sb.Add(AddUrlQuery("namespaces", namespace))
	fmt.Println(sb.Build())
	return sb.Build()
}

func Scale(k8sClient *kubernetes.Clientset, namespace string, microservice string, replicas int32) {
	// Get the deployment
	deployment, err := k8sClient.AppsV1().Deployments(namespace).Get(context.TODO(), microservice, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Error retrieving deployment: %s\n", err.Error())
		return
	}
	// // Update the number of replicas
	// deployment.Spec.Replicas = &replicas

	// // Apply the updated deployment
	// updatedDeployment, err := k8sClient.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
	// if err != nil {
	// 	fmt.Printf("Error updating deployment: %s\n", err.Error())
	// 	return
	// }
	// fmt.Printf("Deployment %s scaled to %d replicas\n", updatedDeployment.Name, replicas)
	cmd := exec.Command("kubectl", "scale", fmt.Sprintf("--replicas=%d", replicas), fmt.Sprintf("deploy/%s", microservice))

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error scaling: %s\n", err.Error())
		return
	}
	// Wait for the deployment to have all replicas ready
	if err := waitForDeploymentReady(k8sClient, namespace, deployment.Name, replicas, 5*time.Minute); err != nil {
		fmt.Printf("deployment %s did not reach the desired state: %v", deployment.Name, err)
		return
	}

	fmt.Printf("Deployment %s is ready with %d replicas\n", deployment.Name, replicas)
}

// waitForDeploymentReady waits until the deployment's ready replicas equal the desired replicas.
func waitForDeploymentReady(k8sClient *kubernetes.Clientset, namespace string, deploymentName string, replicas int32, timeout time.Duration) error {
	timeoutTimer := time.After(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutTimer:
			return fmt.Errorf("timeout exceeded while waiting for deployment %s to become ready", deploymentName)
		case <-ticker.C:
			deployment, err := k8sClient.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error retrieving deployment: %v", err)
			}

			if deployment.Status.ReadyReplicas == replicas {
				return nil
			}

			fmt.Printf("Waiting for deployment %s: %d/%d replicas ready\n", deploymentName, deployment.Status.ReadyReplicas, replicas)
		}
	}
}
