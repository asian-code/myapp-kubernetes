terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 6.23"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}
