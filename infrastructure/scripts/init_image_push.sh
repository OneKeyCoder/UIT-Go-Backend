#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Move to project root
cd $SCRIPT_DIR/../..

if command -v podman &> /dev/null; then
    cmd=podman
elif command -v docker &> /dev/null; then
    cmd=docker
else
    echo "Error: Neither podman nor docker found in PATH"
    exit 1
fi

# Check if ACR name is provided
if [ -z "$1" ]; then
    echo "Error: ACR name is required"
    echo "Usage: $0 <acr-name> [tag]"
    exit 1
fi

ACR_NAME=$1
TAG=${2:-latest}

# Login to ACR
echo "Logging in to Azure Container Registry: $ACR_NAME"
DOCKER_COMMAND=$cmd az acr login --name $ACR_NAME

$cmd compose build

services=("api-gateway" "logger-service" "authentication-service" "location-service" "trip-service" "user-service")

# Push each service to ACR
for service in "${services[@]}"; do
    echo "Processing $service..."
    
    # Tag the image for ACR
    $cmd tag $service:latest $ACR_NAME.azurecr.io/$service:$TAG
    
    # Push to ACR
    echo "Pushing $service to ACR..."
    $cmd push $ACR_NAME.azurecr.io/$service:$TAG
    
    echo "Successfully pushed $service:$TAG"
done

echo "All services pushed to $ACR_NAME.azurecr.io"