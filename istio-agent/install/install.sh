#!/bin/bash

curl -L https://istio.io/downloadIstio | sh -


istioctl install --set profile=demo
minikube addons enable metrics-server # or kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

kubectl label namespace default istio-injection=enabled

## Deploy sample apps
kubectl apply -f samples/addons/
kubectl apply -f samples/bookinfo/networking/bookinfo-gateway.yaml
kubectl apply -f samples/bookinfo/platform/kube/bookinfo.yaml
