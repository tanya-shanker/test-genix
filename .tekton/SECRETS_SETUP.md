# Secrets Setup Guide for Tekton Pipeline

## Overview

The AI Test Generation task requires two secrets to function:
1. **Bob Shell API Key** - For AI-powered test generation
2. **GitHub Token** - For creating PRs and accessing repositories

## How Secrets Are Mounted

### Current Configuration (secretKeyRef Method)

Secrets are injected as environment variables using Kubernetes `secretKeyRef`:

```yaml
# In ai-test-generation-task.yaml
env:
  - name: BOBSHELL_API_KEY
    valueFrom:
      secretKeyRef:
        name: bobshell-api-key    # Kubernetes secret name
        key: api-key              # Key within the secret
        optional: true            # Pod starts even if secret missing
  
  - name: GITHUB_TOKEN
    valueFrom:
      secretKeyRef:
        name: github-token
        key: token
        optional: true
```

**How It Works:**
- Kubernetes automatically looks up the secret by name
- Extracts the specified key value
- Injects it as an environment variable in the container
- **No workspace mounting needed** - this is the standard Tekton/K8s approach

## Creating Secrets in Your Cluster

### Prerequisites
- Access to your Kubernetes cluster
- `kubectl` configured with appropriate permissions
- Your Bob Shell API key
- Your GitHub personal access token

### Step 1: Create Bob Shell API Key Secret

```bash
# Replace YOUR_BOB_API_KEY with your actual API key
kubectl create secret generic bobshell-api-key \
  --from-literal=api-key=YOUR_BOB_API_KEY \
  --namespace=YOUR_NAMESPACE

# Verify creation
kubectl get secret bobshell-api-key -n YOUR_NAMESPACE
kubectl describe secret bobshell-api-key -n YOUR_NAMESPACE
```

### Step 2: Create GitHub Token Secret

```bash
# Replace YOUR_GITHUB_TOKEN with your actual token
kubectl create secret generic github-token \
  --from-literal=token=YOUR_GITHUB_TOKEN \
  --namespace=YOUR_NAMESPACE

# Verify creation
kubectl get secret github-token -n YOUR_NAMESPACE
kubectl describe secret github-token -n YOUR_NAMESPACE
```

### Step 3: Verify Secrets Are Available

```bash
# List all secrets in namespace
kubectl get secrets -n YOUR_NAMESPACE

# Should show:
# NAME                TYPE     DATA   AGE
# bobshell-api-key    Opaque   1      1m
# github-token        Opaque   1      1m
```

## GitHub Token Permissions

Your GitHub token needs these permissions:
- `repo` - Full control of private repositories
- `workflow` - Update GitHub Action workflows
- `write:packages` - Upload packages to GitHub Package Registry (if needed)

### Creating a GitHub Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a descriptive name: "Tekton AI Test Generator"
4. Select scopes:
   - ✅ `repo` (all sub-scopes)
   - ✅ `workflow`
5. Click "Generate token"
6. **Copy the token immediately** (you won't see it again)

## IBM Cloud Specific Setup

If using IBM Cloud Kubernetes Service:

```bash
# Login to IBM Cloud
ibmcloud login

# Target your cluster
ibmcloud ks cluster config --cluster YOUR_CLUSTER_NAME

# Get your namespace
kubectl get namespaces

# Create secrets in the correct namespace
kubectl create secret generic bobshell-api-key \
  --from-literal=api-key=YOUR_BOB_API_KEY \
  --namespace=YOUR_TEKTON_NAMESPACE

kubectl create secret generic github-token \
  --from-literal=token=YOUR_GITHUB_TOKEN \
  --namespace=YOUR_TEKTON_NAMESPACE
```

## Updating Secrets

If you need to update a secret:

```bash
# Delete old secret
kubectl delete secret bobshell-api-key -n YOUR_NAMESPACE

# Create new secret with updated value
kubectl create secret generic bobshell-api-key \
  --from-literal=api-key=NEW_API_KEY \
  --namespace=YOUR_NAMESPACE
```

Or update in place:

```bash
# Encode your new value
echo -n "NEW_API_KEY" | base64

# Edit the secret
kubectl edit secret bobshell-api-key -n YOUR_NAMESPACE
# Update the base64 encoded value in the editor
```

## Troubleshooting

### Secret Not Found Error

**Symptom:** Task fails with "secret not found" error

**Solution:**
1. Verify secret exists: `kubectl get secret SECRET_NAME -n NAMESPACE`
2. Check secret is in correct namespace
3. Verify secret name matches task parameter (default: `bobshell-api-key`, `github-token`)

### Permission Denied

**Symptom:** Task runs but API calls fail with 401/403 errors

**Solution:**
1. Verify API key/token is correct
2. Check token hasn't expired
3. Verify token has required permissions
4. Test token manually:
   ```bash
   # Test GitHub token
   curl -H "Authorization: token YOUR_TOKEN" https://api.github.com/user
   
   # Test Bob API key (adjust URL as needed)
   curl -H "Authorization: Bearer YOUR_API_KEY" https://bob-api-endpoint
   ```

### Secret Value Is Empty

**Symptom:** Environment variable is set but empty

**Solution:**
1. Check secret key name matches: `api-key` for Bob, `token` for GitHub
2. Verify secret data:
   ```bash
   kubectl get secret SECRET_NAME -n NAMESPACE -o yaml
   ```
3. Decode and verify value:
   ```bash
   kubectl get secret SECRET_NAME -n NAMESPACE -o jsonpath='{.data.api-key}' | base64 -d
   ```

## Security Best Practices

1. **Never commit secrets to Git**
   - Secrets should only exist in Kubernetes
   - Use `.gitignore` for any local secret files

2. **Use RBAC to restrict access**
   ```bash
   # Only allow specific service accounts to access secrets
   kubectl create role secret-reader \
     --verb=get,list \
     --resource=secrets \
     --namespace=YOUR_NAMESPACE
   ```

3. **Rotate secrets regularly**
   - Update GitHub tokens every 90 days
   - Update API keys when team members leave

4. **Use optional: true for graceful degradation**
   - Pipeline can still run without secrets (for testing)
   - Tasks should handle missing secrets gracefully

5. **Monitor secret access**
   ```bash
   # Check which pods are using secrets
   kubectl get pods -n YOUR_NAMESPACE -o json | \
     jq '.items[] | select(.spec.containers[].env[]?.valueFrom.secretKeyRef != null) | .metadata.name'
   ```

## Verification Script

Save this as `verify-secrets.sh`:

```bash
#!/bin/bash

NAMESPACE="${1:-default}"

echo "Checking secrets in namespace: $NAMESPACE"
echo "=========================================="

# Check Bob Shell API key
if kubectl get secret bobshell-api-key -n "$NAMESPACE" &>/dev/null; then
  echo "✅ bobshell-api-key exists"
  KEY_LENGTH=$(kubectl get secret bobshell-api-key -n "$NAMESPACE" -o jsonpath='{.data.api-key}' | base64 -d | wc -c)
  echo "   Key length: $KEY_LENGTH characters"
else
  echo "❌ bobshell-api-key NOT FOUND"
fi

# Check GitHub token
if kubectl get secret github-token -n "$NAMESPACE" &>/dev/null; then
  echo "✅ github-token exists"
  TOKEN_LENGTH=$(kubectl get secret github-token -n "$NAMESPACE" -o jsonpath='{.data.token}' | base64 -d | wc -c)
  echo "   Token length: $TOKEN_LENGTH characters"
else
  echo "❌ github-token NOT FOUND"
fi

echo ""
echo "To create missing secrets, see SECRETS_SETUP.md"
```

Run with:
```bash
chmod +x verify-secrets.sh
./verify-secrets.sh YOUR_NAMESPACE
```

## Summary

✅ **Secrets ARE mounted** via `secretKeyRef` in environment variables
✅ **No workspace needed** - Kubernetes handles mounting automatically
✅ **Optional flag** allows pipeline to run even if secrets are missing
✅ **Standard approach** - follows Tekton and Kubernetes best practices

The secrets are properly configured in the task definition. You just need to create them in your Kubernetes cluster using the commands above.