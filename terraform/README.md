# Terraform IaC

AWS infrastructure for the `hslee` project, managed as code.

## Structure

```
terraform/
├── bootstrap/              # One-time: creates S3 state bucket + DynamoDB lock table
├── modules/
│   ├── vpc/                # VPC, subnets, IGW, NAT gateways, route tables
│   ├── eks/                # EKS cluster, managed node group, OIDC provider
│   ├── ecr/                # ECR repositories with lifecycle policies
│   └── iam/                # GitHub Actions OIDC + CI/CD deploy role
└── environments/
    ├── staging/            # Staging: SPOT nodes, single NAT, t3.medium
    └── production/         # Production: ON_DEMAND nodes, HA NAT, t3.large
```

## First-Time Setup

### 1. Bootstrap remote state

```bash
cd terraform/bootstrap
terraform init
terraform apply -var="aws_account_id=$(aws sts get-caller-identity --query Account --output text)"
cd -   # return to repo root
```

### 2. Update backend bucket name

Replace `ACCOUNT_ID` in both `backend.tf` files with your AWS account ID:

```bash
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
# macOS: sed -i '' requires an empty-string backup extension
sed -i '' "s/ACCOUNT_ID/$AWS_ACCOUNT_ID/g" \
  terraform/environments/staging/backend.tf \
  terraform/environments/production/backend.tf
# Linux:
# sed -i "s/ACCOUNT_ID/$AWS_ACCOUNT_ID/g" \
#   terraform/environments/staging/backend.tf \
#   terraform/environments/production/backend.tf
```

### 3. Deploy an environment

```bash
cd terraform/environments/staging
terraform init
terraform plan
terraform apply
```

## GitHub Actions Integration

After `terraform apply`, add the `cicd_role_arn` output value to GitHub Secrets as `AWS_ROLE_ARN`.

Workflows use OIDC to assume this role — no long-lived credentials stored.

## PR Review Policy

All infrastructure changes require PR review before merge. The branch protection rule
on `main` enforces this. Never apply Terraform directly from a feature branch in production.

## State

| Environment | State file                        | Lock table              |
|-------------|-----------------------------------|-------------------------|
| Staging     | `s3://.../staging/terraform.tfstate`     | `hslee-terraform-lock` |
| Production  | `s3://.../production/terraform.tfstate`  | `hslee-terraform-lock` |
