terraform {
  backend "s3" {
    bucket       = "en-tf-state"
    key          = "eks-three-tier-app/mspy.tfstate"
    region       = "us-east-1"
    use_lockfile = true
    encrypt      = true
  }
}
# Get current AWS account info
data "aws_caller_identity" "current" {}

# Get available availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# Configure kubectl provider data sources
# data "aws_eks_cluster" "cluster" {
#   name       = module.eks.cluster_name
#   depends_on = [module.eks]
# }

# data "aws_eks_cluster_auth" "cluster" {
#   name       = module.eks.cluster_name
#   depends_on = [module.eks]
# }