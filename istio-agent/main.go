package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MohammedGouaouri/get-pod-metrics/constants"
	"github.com/MohammedGouaouri/get-pod-metrics/kiali"
	"github.com/MohammedGouaouri/get-pod-metrics/queue"
	"github.com/MohammedGouaouri/get-pod-metrics/utils"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	ConfPath   = utils.GetEnv("KubeConfig", "./config")
	KubeConfig = flag.String("config", ConfPath, "")
)

func main() {

	cf := &genericclioptions.ConfigFlags{
		KubeConfig: KubeConfig,
	}
	restConfig, err := cf.ToRESTConfig()
	if err != nil {
		fmt.Printf("Rest cnf error: %v\n", err)
		return
	}

	k8s, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		fmt.Printf("kubernetes cnf error: %v\n", err)
		return
	}
	restConfig.WarningHandler = rest.NoWarnings{}
	restConfig.QPS = 1000
	restConfig.Burst = 1000

	// fmt.Printf("Deployment %s scaled to %d replicas\n", updatedDeployment.Name, a)
	metricsClient, err := metricsclientset.NewForConfig(restConfig)
	if err != nil {
		fmt.Printf("Metrics cnf error: %v\n", err)
		return
	}

	// Define the URL
	url := kiali.BuildKialiGraphUrl(constants.KIALI_URL, "default", "60s")
	pub := queue.NewPublisher("scaler")

	// Declare queue consumer to receive commands
	consumer := queue.NewCosnumer("commands", "istio-agent")
	commandChan := make(chan string)
	// Consume is a blocking function
	go consumer.Consume(commandChan)

	// Goroutine that continously reads graph data from kiali and publishes it
	stopGraphCollectionChan := make(chan bool)
	stopCommandConsumption := make(chan bool)

	var mu sync.RWMutex

	go func() {
		for {
			select {
			case <-stopGraphCollectionChan:
				log.Println("Stopping collection")
				return

			default:
				mu.Lock()
				s := kiali.GetWorkloadGraph(url)
				s = kiali.InjectFeatures(s, k8s, metricsClient)

				root := s.Convert()
				jsonData, err := json.Marshal(root)
				if err != nil {
					fmt.Println("Error marshalling to JSON:", err)
					mu.Unlock()
					continue
				}
				if root.ID != "" {
					// If there's data to send
					pub.Publish(string(jsonData))
				}
				time.Sleep(time.Second * time.Duration(constants.GRAPH_COLLECTION_PERIOD))
				mu.Unlock()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-stopCommandConsumption:
				log.Println("Stopping consumption")
				return
			case command := <-commandChan:
				log.Printf("Received command %v\n", command)
				var cmd queue.QueueCommand
				err := json.Unmarshal([]byte(command), &cmd)
				if err != nil {
					fmt.Println("Error unmarshaling JSON:", err)
					return
				}
				switch cmd.Command {
				case "SCALE":
					var wg sync.WaitGroup
					// pause <- true
					mu.Lock()
					scalingAction := cmd.Args
					for m, r := range scalingAction {
						wg.Add(1) // Increment the WaitGroup counter for each goroutine
						go func(m string, r int32) {
							defer wg.Done() // Decrement the counter when the goroutine completes
							kiali.Scale(k8s, "default", m, r)
						}(m, r)
					}
					wg.Wait()
					// time.Sleep(5 * time.Second)
					mu.Unlock()
					// resume <- true
				}
			}
		}
	}()

	waitForTermination(stopGraphCollectionChan, stopCommandConsumption)
}

func waitForTermination(waiters ...chan bool) {
	// Channel that is notified when we are done and should exit
	doneChan := make(chan bool)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range signalChan {
			log.Println("Termination Signal Received")
			doneChan <- true
		}
	}()

	<-doneChan
	for _, waiter := range waiters {
		waiter <- true
	}
}
