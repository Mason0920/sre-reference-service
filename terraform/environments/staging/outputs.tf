output "vpc_id" {
  value = module.vpc.vpc_id
}

output "eks_cluster_name" {
  value = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "ecr_urls" {
  value = module.ecr.repository_urls
}

output "cicd_role_arn" {
  description = "Add this ARN to GitHub Actions as AWS_ROLE_ARN secret"
  value       = module.iam.cicd_role_arn
}
