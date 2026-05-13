project    = "hslee"
aws_region = "ap-northeast-2"
github_org  = "Mason0920"
github_repo = "hslee"

# Restrict EKS API endpoint to office/VPN CIDRs only.
# Replace with actual office and VPN CIDR ranges before applying.
eks_public_access_cidrs = ["REPLACE_WITH_OFFICE_CIDR/32", "REPLACE_WITH_VPN_CIDR/24"]
