#!/bin/sh

TEMP_KUBECONFIG_FILE=/tmp/kubeconfig

TEMP_FILE=$(mktemp)

# Replace paths in the kubeconfig file and save to a temporary file
sed "s|/home/mohammeddhiyaeddinegouaouri/.minikube/ca.crt|/config/ca.crt|g; \
     s|/home/mohammeddhiyaeddinegouaouri/.minikube/profiles/minikube/client.crt|/config/client.crt|g; \
     s|/home/mohammeddhiyaeddinegouaouri/.minikube/profiles/minikube/client.key|/config/client.key|g" \
     $TEMP_KUBECONFIG_FILE > $TEMP_FILE

# Safely overwrite the original kubeconfig file
cp "$TEMP_FILE" /config/kubeconfig

echo "Kubeconfig paths updated successfully:"