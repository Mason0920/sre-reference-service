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

  default_tags {
    tags = {
      Project     = var.project
      Environment = "staging"
      ManagedBy   = "terraform"
      Repository  = "github.com/${var.github_org}/${var.github_repo}"
    }
  }
}

module "vpc" {
  source = "../../modules/vpc"

  project            = var.project
  environment        = "staging"
  vpc_cidr           = var.vpc_cidr
  az_count           = 2
  single_nat_gateway = true # cost saving for staging
}

module "ecr" {
  source = "../../modules/ecr"

  project          = var.project
  repository_names = [var.project]
}

module "eks" {
  source = "../../modules/eks"

  project               = var.project
  environment           = "staging"
  vpc_id                = module.vpc.vpc_id
  private_subnet_ids    = module.vpc.private_subnet_ids
  kubernetes_version    = var.kubernetes_version
  public_access_enabled = true

  node_instance_types = ["t3.medium"]
  node_capacity_type  = "SPOT"
  node_desired_size   = 1
  node_min_size       = 1
  node_max_size       = 3
}

module "iam" {
  source = "../../modules/iam"

  project             = var.project
  github_org          = var.github_org
  github_repo         = var.github_repo
  ecr_repository_arns = values(module.ecr.repository_arns)
}
