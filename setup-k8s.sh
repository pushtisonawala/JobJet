#!/bin/bash
# Quick setup script for JobJet K8s integration

set -e

echo "🚀 JobJet Kubernetes Setup"
echo "=========================="
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Please install kubectl first."
    exit 1
fi

# Check if connected to cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Not connected to a Kubernetes cluster."
    echo "   Please configure your kubeconfig first."
    exit 1
fi

echo "✓ Connected to cluster"
echo ""

# Install CRD
echo "📦 Installing JobDefinition CRD..."
kubectl apply -f jobdefinition-crd.yaml

# Wait for CRD to be ready
echo "⏳ Waiting for CRD to be established..."
kubectl wait --for condition=established --timeout=60s crd/jobdefinitions.jobjet.dev

echo "✓ CRD installed successfully"
echo ""

# Show instructions
echo "🎯 Next steps:"
echo ""
echo "1. Start the controller:"
echo "   go run cmd/controller/main.go"
echo ""
echo "2. In another terminal, create a job:"
echo "   kubectl apply -f job.yaml"
echo ""
echo "3. Watch the status:"
echo "   kubectl get jd -w"
echo ""
echo "4. See full details:"
echo "   kubectl get jobdefinitions -o yaml"
echo ""
echo "✨ Setup complete!"
