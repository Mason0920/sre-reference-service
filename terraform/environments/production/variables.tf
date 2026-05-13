variable "project" {
  description = "Project name"
  type        = string
  default     = "hslee"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-2"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.20.0.0/16"
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.31"
}

variable "github_org" {
  description = "GitHub organization or username"
  type        = string
}

variable "github_repo" {
  description = "GitHub repository name"
  type        = string
  default     = "hslee"
}

variable "eks_public_access_cidrs" {
  description = "CIDRs allowed to reach the production EKS API endpoint. Must be restricted to office/VPN IPs."
  type        = list(string)
  # No default — must be explicitly set in terraform.tfvars or via -var flag.
}
