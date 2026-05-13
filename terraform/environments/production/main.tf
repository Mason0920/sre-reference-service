terraform {
  required_version = ">= 1.6"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

module "eks_cluster" {
  source = "../../modules/eks-cluster"

  project    = "hslee"
  environment = "production"
  aws_region  = var.aws_region

  vpc_cidr           = "10.2.0.0/16"
  az_count           = 3
  single_nat_gateway = false

  kubernetes_version  = "1.29"
  node_instance_types = ["t3.large"]
  node_capacity_type  = "ON_DEMAND"
  node_min_size       = 2
  node_max_size       = 6
  node_desired_size   = 3

  tags = {
    Environment = "production"
    Project     = "hslee"
  }
}
