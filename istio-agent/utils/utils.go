package utils

import (
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetEnv(env, defVal string) string {
	if value := os.Getenv(env); value != "" {
		return value
	}
	return defVal
}

func ReadConfig(path string) *rest.Config {

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		log.Println(err)
	}

	config.Burst = 1000
	config.QPS = 1000

	return config
}

func NewK8sClient(restConfigs *rest.Config) *kubernetes.Clientset {
	client, err := kubernetes.NewForConfig(restConfigs)
	if err != nil {
		log.Println(err)
	}

	return client
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
