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

  backend "s3" {
    bucket         = "hslee-terraform-state-ACCOUNT_ID"
    key            = "staging/terraform.tfstate"
    region         = "ap-northeast-2"
    dynamodb_table = "hslee-terraform-lock"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region
}

module "eks_cluster" {
  source = "../../modules/eks-cluster"

  project    = "hslee"
  environment = "staging"
  aws_region  = var.aws_region

  vpc_cidr           = "10.1.0.0/16"
  az_count           = 2
  single_nat_gateway = true

  kubernetes_version  = "1.29"
  node_instance_types = ["t3.medium"]
  node_capacity_type  = "ON_DEMAND"
  node_min_size       = 1
  node_max_size       = 4
  node_desired_size   = 2

  tags = {
    Environment = "staging"
    Project     = "hslee"
  }
}
