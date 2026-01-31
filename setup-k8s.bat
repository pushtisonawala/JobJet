@echo off
REM Quick setup script for JobJet K8s integration (Windows)

echo.
echo 🚀 JobJet Kubernetes Setup
echo ==========================
echo.

REM Check if kubectl is available
where kubectl >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ kubectl not found. Please install kubectl first.
    exit /b 1
)

REM Check if connected to cluster
kubectl cluster-info >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Not connected to a Kubernetes cluster.
    echo    Please configure your kubeconfig first.
    exit /b 1
)

echo ✓ Connected to cluster
echo.

REM Install CRD
echo 📦 Installing JobDefinition CRD...
kubectl apply -f jobdefinition-crd.yaml

REM Wait for CRD to be ready
echo ⏳ Waiting for CRD to be established...
kubectl wait --for condition=established --timeout=60s crd/jobdefinitions.jobjet.dev

echo ✓ CRD installed successfully
echo.

REM Show instructions
echo 🎯 Next steps:
echo.
echo 1. Start the controller:
echo    go run cmd/controller/main.go
echo.
echo 2. In another terminal, create a job:
echo    kubectl apply -f job.yaml
echo.
echo 3. Watch the status:
echo    kubectl get jd -w
echo.
echo 4. See full details:
echo    kubectl get jobdefinitions -o yaml
echo.
echo ✨ Setup complete!
