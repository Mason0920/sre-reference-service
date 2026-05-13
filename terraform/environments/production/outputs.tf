output "cluster_name" {
  description = "EKS cluster name"
  value       = module.eks_cluster.cluster_name
}

output "cluster_endpoint" {
  description = "EKS API server endpoint"
  value       = module.eks_cluster.cluster_endpoint
}

output "vpc_id" {
  description = "VPC ID"
  value       = module.eks_cluster.vpc_id
}

output "oidc_provider_arn" {
  description = "OIDC provider ARN (for IRSA)"
  value       = module.eks_cluster.oidc_provider_arn
}

output "kubeconfig_command" {
  description = "Command to configure kubectl"
  value       = module.eks_cluster.kubeconfig_command
}
