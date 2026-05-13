# ArgoCD Installation

Install ArgoCD on the EKS cluster. Run these commands against the **staging** cluster where ArgoCD will be hosted.

## 1. Create namespace and install ArgoCD

```bash
kubectl create namespace argocd

# Install ArgoCD (stable release)
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

## 2. Install ArgoCD CLI

```bash
# macOS
brew install argocd

# Linux
curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
chmod +x argocd-linux-amd64 && sudo mv argocd-linux-amd64 /usr/local/bin/argocd
```

## 3. Expose ArgoCD API server (VPN/SSO, internal only)

```bash
# Option A: LoadBalancer (restrict to VPN CIDR via security group)
kubectl patch svc argocd-server -n argocd -p '{"spec":{"type":"LoadBalancer"}}'

# Option B: Port-forward for initial setup
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

## 4. Get initial admin password

```bash
argocd admin initial-password -n argocd
```

## 5. Login and change password

```bash
argocd login <ARGOCD_SERVER> --username admin --password <initial-password>
argocd account update-password
```

## 6. Register production cluster

ArgoCD is installed in staging cluster. Register the production EKS cluster:

```bash
# Switch to production cluster context first, then add it to ArgoCD
kubectl config use-context <production-cluster-context>
argocd cluster add <production-cluster-context> --name production
kubectl config use-context <staging-cluster-context>
```

## 7. Apply Application CRDs

```bash
kubectl apply -f ../applications/staging.yaml
kubectl apply -f ../applications/production.yaml
```

## 8. Store GitHub Actions secrets

Add these to GitHub repository secrets/vars:
- `ARGOCD_AUTH_TOKEN` — service account token (see step below)
- `ARGOCD_SERVER` — ArgoCD server hostname (without https://)

```bash
# Create CI service account with read+sync permission
argocd account generate-token --account ci-deploy
```

Configure the `ci-deploy` account in the ArgoCD ConfigMap:
```yaml
# kubectl edit configmap argocd-cm -n argocd
data:
  accounts.ci-deploy: apiKey, login
```

Apply RBAC:
```yaml
# kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-rbac-cm
  namespace: argocd
data:
  policy.csv: |
    p, role:ci-deploy, applications, get, */*, allow
    p, role:ci-deploy, applications, sync, */*, allow
    p, role:ci-deploy, applications, action/*, */*, allow
    g, ci-deploy, role:ci-deploy
EOF
```
