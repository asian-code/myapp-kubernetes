terraform {
  required_version = "~> 1.13.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
    # for managing Helm charts and apps
    helm = {
      source  = "hashicorp/helm"
      version = "~> 3.0"
    }
    # for managing Kubernetes resources
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}
