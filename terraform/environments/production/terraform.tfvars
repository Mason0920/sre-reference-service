project    = "hslee"
aws_region = "ap-northeast-2"
github_org  = "Mason0920"
github_repo = "hslee"

# Restrict EKS API endpoint to office/VPN CIDRs only.
# Replace with actual office/VPN CIDRs before applying.
# RFC 5737 documentation IPs used as placeholders — not routable.
eks_public_access_cidrs = ["192.0.2.1/32", "198.51.100.0/24"]
