#!/bin/bash
# Deployment helper script for EKS
# This script helps with common deployment tasks

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AWS_REGION=${AWS_REGION:-us-west-2}
EKS_CLUSTER_NAME=${EKS_CLUSTER_NAME:-three-tier-eks-cluster}
ECR_REPOSITORY_FRONTEND=${ECR_REPOSITORY_FRONTEND:-three-tier-app/frontend}
ECR_REPOSITORY_BACKEND=${ECR_REPOSITORY_BACKEND:-three-tier-app/backend}

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
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI is not installed. Please install it first."
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
    
    log_success "All prerequisites are installed"
}

setup_aws_credentials() {
    log_info "Checking AWS credentials..."
    
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS credentials not configured. Please run 'aws configure' first."
        exit 1
    fi
    
    log_success "AWS credentials are configured"
}

update_kubeconfig() {
    log_info "Updating kubeconfig for EKS cluster..."
    
    aws eks update-kubeconfig --region $AWS_REGION --name $EKS_CLUSTER_NAME
    
    if kubectl cluster-info &> /dev/null; then
        log_success "Successfully connected to EKS cluster"
    else
        log_error "Failed to connect to EKS cluster"
        exit 1
    fi
}

build_and_push_images() {
    log_info "Building and pushing Docker images..."
    
    # Get ECR login token
    aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $(aws sts get-caller-identity --query Account --output text).dkr.ecr.$AWS_REGION.amazonaws.com
    
    # Build and push frontend
    log_info "Building frontend image..."
    docker build -t $ECR_REPOSITORY_FRONTEND:latest ./app-src/frontend/
    docker tag $ECR_REPOSITORY_FRONTEND:latest $(aws sts get-caller-identity --query Account --output text).dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY_FRONTEND:latest
    docker push $(aws sts get-caller-identity --query Account --output text).dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY_FRONTEND:latest
    
    # Build and push backend
    log_info "Building backend image..."
    docker build -t $ECR_REPOSITORY_BACKEND:latest ./app-src/backend/
    docker tag $ECR_REPOSITORY_BACKEND:latest $(aws sts get-caller-identity --query Account --output text).dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY_BACKEND:latest
    docker push $(aws sts get-caller-identity --query Account --output text).dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPOSITORY_BACKEND:latest
    
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
    
    LOAD_BALANCER_URL=$(kubectl get ingress frontend-web-ingress -n frontend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || echo "")
    
    if [[ -n "$LOAD_BALANCER_URL" ]]; then
        log_success "Application URL: https://$LOAD_BALANCER_URL"
    else
        log_warning "Load balancer URL not ready yet. Please wait a few minutes and try again."
    fi
}

cleanup_resources() {
    log_warning "This will delete all resources. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        log_info "Deleting Kubernetes resources..."
        kubectl delete -f k8s-manifests/ --ignore-not-found=true
        
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
        setup_aws_credentials
        deploy_infrastructure
        update_kubeconfig
        build_and_push_images
        deploy_applications
        check_cluster_health
        get_application_url
        ;;
    "build")
        check_prerequisites
        setup_aws_credentials
        build_and_push_images
        ;;
    "infra")
        check_prerequisites
        setup_aws_credentials
        deploy_infrastructure
        ;;
    "apps")
        check_prerequisites
        setup_aws_credentials
        update_kubeconfig
        deploy_applications
        ;;
    "health")
        check_prerequisites
        setup_aws_credentials
        update_kubeconfig
        check_cluster_health
        ;;
    "url")
        check_prerequisites
        setup_aws_credentials
        update_kubeconfig
        get_application_url
        ;;
    "cleanup")
        check_prerequisites
        setup_aws_credentials
        update_kubeconfig
        cleanup_resources
        ;;
    *)
        echo "Usage: $0 {deploy|build|infra|apps|health|url|cleanup}"
        echo ""
        echo "Commands:"
        echo "  deploy   - Full deployment (infrastructure + applications)"
        echo "  build    - Build and push Docker images"
        echo "  infra    - Deploy infrastructure only"
        echo "  apps     - Deploy applications only"
        echo "  health   - Check cluster health"
        echo "  url      - Get application URL"
        echo "  cleanup  - Delete all resources"
        exit 1
        ;;
esac
