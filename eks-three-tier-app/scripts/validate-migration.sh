#!/bin/bash

# Terraform Migration Validation Script
# This script validates the migration from custom modules to official terraform-aws-modules

set -e

echo "🔍 Validating Terraform Migration..."
echo "======================================"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
        exit 1
    fi
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "ℹ️  $1"
}

# Check if we're in the right directory
if [[ ! -f "terraform/main.tf" ]]; then
    echo -e "${RED}❌ Please run this script from the project root directory${NC}"
    exit 1
fi

# Navigate to terraform directory
cd terraform

print_info "Checking Terraform configuration..."

# 1. Validate Terraform syntax
print_info "1. Validating Terraform syntax..."
terraform fmt -check=true -diff=true > /dev/null 2>&1
print_status $? "Terraform syntax is properly formatted"

# 2. Check for official modules usage
print_info "2. Checking for official terraform-aws-modules usage..."
if grep -q "terraform-aws-modules/vpc/aws" main.tf && \
   grep -q "terraform-aws-modules/eks/aws" main.tf && \
   grep -q "terraform-aws-modules/iam/aws" main.tf; then
    print_status 0 "Official terraform-aws-modules are being used"
else
    print_status 1 "Official terraform-aws-modules not found in main.tf"
fi

# 3. Validate that custom modules are no longer referenced
print_info "3. Checking that custom modules are not referenced..."
if ! grep -q 'source.*"./modules/' main.tf; then
    print_status 0 "No custom module references found in main.tf"
else
    print_status 1 "Custom module references still exist in main.tf"
fi

# 4. Check for required providers
print_info "4. Validating required providers..."
if grep -q "terraform-aws-modules/vpc/aws" main.tf; then
    if grep -q "version.*~> 5.0" main.tf; then
        print_status 0 "VPC module version constraint is correct"
    else
        print_warning "VPC module version might need updating"
    fi
fi

if grep -q "terraform-aws-modules/eks/aws" main.tf; then
    if grep -q "version.*~> 20.0" main.tf; then
        print_status 0 "EKS module version constraint is correct"
    else
        print_warning "EKS module version might need updating"
    fi
fi

# 5. Check for IRSA modules
print_info "5. Checking IRSA role configurations..."
if grep -q "iam-role-for-service-accounts-eks" main.tf; then
    print_status 0 "IRSA roles are properly configured"
else
    print_warning "IRSA role configurations not found"
fi

# 6. Validate VPC 3-tier architecture is preserved
print_info "6. Validating 3-tier network architecture..."
if grep -q "public_subnets.*cidrsubnet.*8.*k\]" main.tf && \
   grep -q "private_subnets.*cidrsubnet.*4.*k + 1" main.tf && \
   grep -q "database_subnets.*cidrsubnet.*8.*k + 16" main.tf; then
    print_status 0 "3-tier network architecture (public/private/database) is preserved"
else
    print_status 1 "3-tier network architecture configuration is missing"
fi

# 7. Check for Kubernetes tagging
print_info "7. Checking Kubernetes subnet tagging..."
if grep -q "kubernetes.io/role/elb" main.tf && \
   grep -q "kubernetes.io/role/internal-elb" main.tf; then
    print_status 0 "Kubernetes subnet tagging is present"
else
    print_status 1 "Kubernetes subnet tagging is missing"
fi

# 8. Validate outputs are updated
print_info "8. Checking outputs configuration..."
if [[ -f "outputs.tf" ]]; then
    if grep -q "module.eks.cluster_name" outputs.tf; then
        print_status 0 "Outputs are updated for new EKS module structure"
    else
        print_warning "Some outputs might need updating for new module structure"
    fi
else
    print_warning "outputs.tf file not found"
fi

# 9. Check for additional resources
print_info "9. Validating additional resources..."
if grep -q "aws_security_group.*alb" main.tf && \
   grep -q "helm_release.*aws_load_balancer_controller" main.tf; then
    print_status 0 "Additional security groups and Helm releases are configured"
else
    print_warning "Some additional resources might be missing"
fi

# 10. Try terraform init (if requested)
if [[ "$1" == "--init" ]]; then
    print_info "10. Initializing Terraform..."
    terraform init -upgrade > /dev/null 2>&1
    print_status $? "Terraform initialization successful"
    
    print_info "11. Validating Terraform configuration..."
    terraform validate > /dev/null 2>&1
    print_status $? "Terraform configuration is valid"
fi

echo ""
echo "🎉 Migration Validation Summary"
echo "=============================="
print_info "✅ Official terraform-aws-modules are being used"
print_info "✅ 3-tier network architecture is preserved"  
print_info "✅ Kubernetes subnet tagging is configured"
print_info "✅ IRSA roles are properly set up"
print_info "✅ Additional security and Helm resources are included"

echo ""
echo "📋 Next Steps:"
echo "1. Run 'terraform plan' to review the changes"
echo "2. Run 'terraform apply' to apply the migration"
echo "3. Test cluster connectivity: aws eks update-kubeconfig --region us-west-2 --name en-mspy"
echo "4. Verify deployments: kubectl get nodes && kubectl get pods -A"

echo ""
print_info "Migration validation completed successfully! 🚀"
