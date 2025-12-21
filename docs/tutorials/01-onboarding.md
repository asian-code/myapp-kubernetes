# Tutorial: Day 1 Onboarding Guide

**Goal**: Go from "zero" to running the full platform locally.
**Time**: ~30 minutes

---

## Prerequisites

Ensure you have the following tools installed:
*   [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or Rancher Desktop)
*   [Go 1.24+](https://go.dev/dl/)
*   [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker)
*   [Helm](https://helm.sh/docs/intro/install/)
*   [kubectl](https://kubernetes.io/docs/tasks/tools/)

---

## Step 1: Clone the Repository

```bash
git clone https://github.com/your-org/myapp-kubernetes.git
cd myapp-kubernetes
```

## Step 2: Start a Local Cluster

We use `kind` to spin up a local Kubernetes cluster that mimics our production EKS environment.

```bash
kind create cluster --name myhealth-local
kubectl cluster-info --context kind-myhealth-local
```

## Step 3: Build Docker Images

Since we are running locally, we need to build the images and load them into the Kind cluster (so it doesn't try to pull from ECR).

```bash
# Build images
docker build -t api-service:local ./services/api-service
docker build -t data-processor:local ./services/data-processor

# Load into Kind
kind load docker-image api-service:local --name myhealth-local
kind load docker-image data-processor:local --name myhealth-local
```

## Step 4: Deploy with Helm

We use the same Helm chart for local development as we do for production, just with different values.

```bash
cd helm/myhealth

# Install the chart
helm install myhealth . \
  --set global.environment=local \
  --set apiService.image.repository=api-service \
  --set apiService.image.tag=local \
  --set dataProcessor.image.repository=data-processor \
  --set dataProcessor.image.tag=local
```

## Step 5: Verify Deployment

Check if the pods are running:

```bash
kubectl get pods
```

You should see:
*   `api-service-xxx` (Running)
*   `data-processor-xxx` (Running)
*   `postgres-xxx` (Running - if enabled locally)

## Step 6: Access the API

Port-forward the API service to your local machine:

```bash
kubectl port-forward svc/api-service 8080:8080
```

Now open your browser or curl:
`http://localhost:8080/health`

---

## 🎉 Congratulations!
You have successfully deployed the myHealth platform locally. 

**Next Steps:**
*   Read the [Architecture Overview](../explanation/architecture-overview.md)
*   Check out the [Service Catalog](../reference/service-catalog.md)
