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
      Environment = "production"
      ManagedBy   = "terraform"
      Repository  = "github.com/${var.github_org}/${var.github_repo}"
    }
  }
}

# GitHub Actions OIDC provider is account-wide; created once in bootstrap/.
data "aws_iam_openid_connect_provider" "github_actions" {
  url = "https://token.actions.githubusercontent.com"
}

module "vpc" {
  source = "../../modules/vpc"

  project            = var.project
  environment        = "production"
  vpc_cidr           = var.vpc_cidr
  az_count           = 2
  single_nat_gateway = false # HA: one NAT per AZ
}

module "ecr" {
  source = "../../modules/ecr"

  project          = var.project
  repository_names = ["${var.project}-production"]
}

module "eks" {
  source = "../../modules/eks"

  project               = var.project
  environment           = "production"
  vpc_id                = module.vpc.vpc_id
  private_subnet_ids    = module.vpc.private_subnet_ids
  kubernetes_version    = var.kubernetes_version
  public_access_enabled = true
  public_access_cidrs   = var.eks_public_access_cidrs

  node_instance_types = ["t3.large"]
  node_capacity_type  = "ON_DEMAND"
  node_desired_size   = 2
  node_min_size       = 2
  node_max_size       = 6
}

module "iam" {
  source = "../../modules/iam"

  project                  = var.project
  environment              = "production"
  github_org               = var.github_org
  github_repo              = var.github_repo
  github_oidc_provider_arn = data.aws_iam_openid_connect_provider.github_actions.arn
  ecr_repository_arns      = values(module.ecr.repository_arns)
}
