#!/bin/bash

check_minikube_status() {
  minikube status &> /dev/null
  if [ $? -eq 0 ]; then
    echo "Minikube is already running."
  else
    echo "Starting Minikube..."
    minikube start
  fi
}

check_istioctl() {
  if ! command -v istioctl &> /dev/null; then
    echo "istioctl not found. Downloading and installing Istio..."
    curl -L https://istio.io/downloadIstio | sh -
    cd istio-*
    export PATH=$PWD/bin:$PATH
    echo "istioctl installed and added to PATH."
  else
    echo "istioctl is already installed."
  fi
}

enable_istio_injection() {
  echo "Labeling the default namespace for Istio injection..."
  kubectl label namespace default istio-injection=enabled --overwrite
}

# Deploy Istio and setup sample applications
deploy_istio() {
  echo "Installing Istio with the demo profile..."
  istioctl install --set profile=demo -y

  echo "Enabling the metrics server in Minikube..."
  minikube addons enable metrics-server || \
  kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

  echo "Deploying sample apps..."
  kubectl apply -f addons/
  kubectl apply -f online-boutique/networking/*
  kubectl apply -f online-boutique/apps/*
}


start_kiali() {
    echo "Starting Kiali"
    nohup istioctl dashboard kiali --address=0.0.0.0 &
}

start_proxy() {
  echo "Starting port forward"

  kubectl port-forward -n istio-system svc/istio-ingressgateway 8080:80
}

# Main script execution
main() {
  check_minikube_status
  check_istioctl
  enable_istio_injection
  deploy_istio
  start_kiali
  start_proxy
}

# Run the main function
main
