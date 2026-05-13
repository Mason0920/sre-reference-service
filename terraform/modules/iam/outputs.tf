output "cicd_role_arn" {
  description = "ARN of the GitHub Actions CI/CD IAM role"
  value       = aws_iam_role.cicd.arn
}

output "cicd_role_name" {
  description = "Name of the GitHub Actions CI/CD IAM role"
  value       = aws_iam_role.cicd.name
}

