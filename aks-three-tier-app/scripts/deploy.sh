#!/bin/bash
# Deployment helper script for AKS
# This script helps with common deployment tasks

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AZURE_REGION=${AZURE_REGION:-eastus}
AKS_CLUSTER_NAME=${AKS_CLUSTER_NAME:-three-tier-aks-cluster}
AKS_RESOURCE_GROUP=${AKS_RESOURCE_GROUP:-three-tier-aks-rg}
ACR_NAME=${ACR_NAME:-threetierappacr}

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Azure CLI is installed
    if ! command -v az &> /dev/null; then
        log_error "Azure CLI is not installed. Please install it first."
        exit 1
    fi
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install it first."
        exit 1
    fi
    
    # Check if terraform is installed
    if ! command -v terraform &> /dev/null; then
        log_error "Terraform is not installed. Please install it first."
        exit 1
    fi
    
    # Check if docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install it first."
        exit 1
    fi
    
    # Check if helm is installed
    if ! command -v helm &> /dev/null; then
        log_error "Helm is not installed. Please install it first."
        exit 1
    fi
    
    log_success "All prerequisites are installed"
}

setup_azure_credentials() {
    log_info "Checking Azure credentials..."
    
    if ! az account show &> /dev/null; then
        log_error "Azure credentials not configured. Please run 'az login' first."
        exit 1
    fi
    
    log_success "Azure credentials are configured"
}

get_aks_credentials() {
    log_info "Getting AKS credentials..."
    
    az aks get-credentials --resource-group $AKS_RESOURCE_GROUP --name $AKS_CLUSTER_NAME
    
    if kubectl cluster-info &> /dev/null; then
        log_success "Successfully connected to AKS cluster"
    else
        log_error "Failed to connect to AKS cluster"
        exit 1
    fi
}

build_and_push_images() {
    log_info "Building and pushing Docker images..."
    
    # Login to ACR
    az acr login --name $ACR_NAME
    
    # Build and push frontend
    log_info "Building frontend image..."
    docker build -t $ACR_NAME.azurecr.io/three-tier-app/frontend:latest ./app-src/frontend/
    docker push $ACR_NAME.azurecr.io/three-tier-app/frontend:latest
    
    # Build and push backend
    log_info "Building backend image..."
    docker build -t $ACR_NAME.azurecr.io/three-tier-app/backend:latest ./app-src/backend/
    docker push $ACR_NAME.azurecr.io/three-tier-app/backend:latest
    
    log_success "Docker images built and pushed successfully"
}

deploy_infrastructure() {
    log_info "Deploying infrastructure with Terraform..."
    
    cd terraform
    terraform init
    terraform plan -var-file="terraform.tfvars"
    terraform apply -auto-approve -var-file="terraform.tfvars"
    cd ..
    
    log_success "Infrastructure deployed successfully"
}

setup_ingress_controller() {
    log_info "Setting up NGINX Ingress Controller..."
    
    # Add NGINX Ingress Helm repository
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo update
    
    # Install NGINX Ingress Controller
    helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
        --namespace ingress-nginx \
        --create-namespace \
        --set controller.service.annotations."service\.beta\.kubernetes\.io/azure-load-balancer-health-probe-request-path"=/healthz
    
    # Wait for ingress controller to be ready
    kubectl wait --namespace ingress-nginx \
        --for=condition=ready pod \
        --selector=app.kubernetes.io/component=controller \
        --timeout=300s
    
    log_success "NGINX Ingress Controller deployed successfully"
}

deploy_applications() {
    log_info "Deploying applications to Kubernetes..."
    
    # Create namespaces
    kubectl create namespace frontend --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace backend --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace database --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
    
    # Deploy security configurations
    kubectl apply -f k8s-manifests/security/
    
    # Deploy database
    log_info "Deploying database..."
    kubectl apply -f k8s-manifests/database/
    kubectl rollout status statefulset/postgresql -n database --timeout=300s
    
    # Deploy backend
    log_info "Deploying backend..."
    kubectl apply -f k8s-manifests/backend/
    kubectl rollout status deployment/backend-api -n backend --timeout=300s
    
    # Deploy frontend
    log_info "Deploying frontend..."
    kubectl apply -f k8s-manifests/frontend/
    kubectl rollout status deployment/frontend-web -n frontend --timeout=300s
    
    # Deploy monitoring
    log_info "Deploying monitoring stack..."
    kubectl apply -f k8s-manifests/monitoring/
    kubectl rollout status deployment/prometheus -n monitoring --timeout=300s
    kubectl rollout status deployment/grafana -n monitoring --timeout=300s
    
    log_success "Applications deployed successfully"
}

setup_azure_monitor() {
    log_info "Setting up Azure Monitor integration..."
    
    # Enable monitoring addon
    az aks enable-addons \
        --resource-group $AKS_RESOURCE_GROUP \
        --name $AKS_CLUSTER_NAME \
        --addons monitoring
    
    # Create Log Analytics workspace
    az monitor log-analytics workspace create \
        --resource-group $AKS_RESOURCE_GROUP \
        --workspace-name three-tier-app-logs \
        --location $AZURE_REGION \
        --query id -o tsv
    
    log_success "Azure Monitor integration configured"
}

setup_key_vault_integration() {
    log_info "Setting up Azure Key Vault integration..."
    
    # Deploy Secret Store CSI Driver
    helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
    helm repo update
    helm upgrade --install csi-secrets-store secrets-store-csi-driver/secrets-store-csi-driver \
        --namespace kube-system \
        --set syncSecret.enabled=true \
        --set enableSecretRotation=true
    
    # Deploy Azure Key Vault Provider
    kubectl apply -f https://raw.githubusercontent.com/Azure/secrets-store-csi-driver-provider-azure/master/deployment/provider-azure-installer.yaml
    
    log_success "Azure Key Vault integration configured"
}

check_cluster_health() {
    log_info "Checking cluster health..."
    
    echo "Cluster nodes:"
    kubectl get nodes -o wide
    
    echo -e "\nCluster pods:"
    kubectl get pods -A
    
    echo -e "\nCluster services:"
    kubectl get services -A
    
    echo -e "\nIngress resources:"
    kubectl get ingress -A
    
    log_success "Cluster health check completed"
}

get_application_url() {
    log_info "Getting application URL..."
    
    EXTERNAL_IP=$(kubectl get service ingress-nginx-controller -n ingress-nginx -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    
    if [[ -n "$EXTERNAL_IP" ]]; then
        log_success "Application URL: http://$EXTERNAL_IP"
        log_info "Grafana URL: http://$EXTERNAL_IP/grafana"
        log_info "Prometheus URL: http://$EXTERNAL_IP/prometheus"
    else
        log_warning "External IP not ready yet. Please wait a few minutes and try again."
    fi
}

cleanup_resources() {
    log_warning "This will delete all resources. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "Deleting Kubernetes resources..."
        kubectl delete -f k8s-manifests/ --ignore-not-found=true
        
        log_info "Deleting ingress controller..."
        helm uninstall ingress-nginx -n ingress-nginx
        kubectl delete namespace ingress-nginx
        
        log_info "Destroying Terraform infrastructure..."
        cd terraform
        terraform destroy -auto-approve -var-file="terraform.tfvars"
        cd ..
        
        log_success "Cleanup completed"
    else
        log_info "Cleanup cancelled"
    fi
}

# Main execution
case "${1:-}" in
    "deploy")
        check_prerequisites
        setup_azure_credentials
        deploy_infrastructure
        get_aks_credentials
        setup_ingress_controller
        build_and_push_images
        deploy_applications
        setup_azure_monitor
        setup_key_vault_integration
        check_cluster_health
        get_application_url
        ;;
    "build")
        check_prerequisites
        setup_azure_credentials
        build_and_push_images
        ;;
    "infra")
        check_prerequisites
        setup_azure_credentials
        deploy_infrastructure
        ;;
    "apps")
        check_prerequisites
        setup_azure_credentials
        get_aks_credentials
        deploy_applications
        ;;
    "ingress")
        check_prerequisites
        setup_azure_credentials
        get_aks_credentials
        setup_ingress_controller
        ;;
    "monitor")
        check_prerequisites
        setup_azure_credentials
        setup_azure_monitor
        ;;
    "keyvault")
        check_prerequisites
        setup_azure_credentials
        get_aks_credentials
        setup_key_vault_integration
        ;;
    "health")
        check_prerequisites
        setup_azure_credentials
        get_aks_credentials
        check_cluster_health
        ;;
    "url")
        check_prerequisites
        setup_azure_credentials
        get_aks_credentials
        get_application_url
        ;;
    "cleanup")
        check_prerequisites
        setup_azure_credentials
        get_aks_credentials
        cleanup_resources
        ;;
    *)
        echo "Usage: $0 {deploy|build|infra|apps|ingress|monitor|keyvault|health|url|cleanup}"
        echo ""
        echo "Commands:"
        echo "  deploy   - Full deployment (infrastructure + applications)"
        echo "  build    - Build and push Docker images"
        echo "  infra    - Deploy infrastructure only"
        echo "  apps     - Deploy applications only"
        echo "  ingress  - Setup NGINX Ingress Controller"
        echo "  monitor  - Setup Azure Monitor integration"
        echo "  keyvault - Setup Azure Key Vault integration"
        echo "  health   - Check cluster health"
        echo "  url      - Get application URLs"
        echo "  cleanup  - Delete all resources"
        exit 1
        ;;
esac
