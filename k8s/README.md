# Krafti Vibe Kubernetes Deployment

This directory contains Kubernetes manifests for deploying Krafti Vibe API using Kustomize.

## Directory Structure

```
k8s/
├── base/                    # Base Kubernetes resources
│   ├── namespace.yaml       # Namespace definition
│   ├── configmap.yaml       # Application configuration
│   ├── secret.yaml          # Secrets (credentials, API keys)
│   ├── postgres-statefulset.yaml  # PostgreSQL database
│   ├── redis-statefulset.yaml     # Redis cache
│   ├── minio-deployment.yaml      # MinIO object storage
│   ├── api-deployment.yaml        # API application
│   ├── ingress.yaml              # Ingress configuration
│   ├── hpa.yaml                  # Horizontal Pod Autoscaler
│   └── kustomization.yaml        # Base kustomization
├── overlays/
│   ├── dev/                # Development environment
│   │   ├── kustomization.yaml
│   │   ├── patch-configmap.yaml
│   │   └── patch-api-deployment.yaml
│   └── prod/               # Production environment
│       ├── kustomization.yaml
│       └── patch-api-deployment.yaml
└── README.md
```

## Prerequisites

1. **Kubernetes cluster** (1.24+)
2. **kubectl** CLI tool
3. **kustomize** (built into kubectl 1.14+)
4. **NGINX Ingress Controller** (for ingress)
5. **cert-manager** (for TLS certificates)
6. **Metrics Server** (for HPA)

## Quick Start

### 1. Update Secrets

Before deploying, update the secrets in `base/secret.yaml`:

```bash
# Generate a strong password
openssl rand -base64 32
```

Update the following in `base/secret.yaml`:
- `DB_PASSWORD`
- `REDIS_PASSWORD`
- `MINIO_SECRET_KEY`
- `LOGTO_JWKS_URI`
- `LOGTO_ISSUER`
- `LOGTO_API_RESOURCE`

### 2. Build Docker Image

```bash
# Build and tag the image
docker build -t kraftivibe/api:latest .

# Push to your container registry
docker tag kraftivibe/api:latest your-registry/kraftivibe/api:latest
docker push your-registry/kraftivibe/api:latest
```

Update image references in `overlays/*/kustomization.yaml` if using a custom registry.

### 3. Deploy to Development

```bash
# Preview the manifests
kubectl kustomize k8s/overlays/dev

# Apply the manifests
kubectl apply -k k8s/overlays/dev

# Watch the deployment
kubectl get pods -n kraftivibe-dev -w
```

### 4. Deploy to Production

```bash
# Preview the manifests
kubectl kustomize k8s/overlays/prod

# Apply the manifests
kubectl apply -k k8s/overlays/prod

# Watch the deployment
kubectl get pods -n kraftivibe -w
```

## Configuration

### Environment-Specific Settings

**Development (`overlays/dev`)**:
- 1 replica
- Debug logging
- Lower resource limits
- SSL disabled for database

**Production (`overlays/prod`)**:
- 3 replicas (min)
- Info logging
- Higher resource limits
- SSL required for database
- Auto-scaling enabled (2-10 pods)

### Horizontal Pod Autoscaler

The HPA automatically scales the API deployment based on:
- **CPU**: Target 70% utilization
- **Memory**: Target 80% utilization
- **Min replicas**: 2
- **Max replicas**: 10

```bash
# Check HPA status
kubectl get hpa -n kraftivibe
```

### Persistent Storage

The following components use persistent storage:
- **PostgreSQL**: 10Gi (StatefulSet)
- **Redis**: 5Gi (StatefulSet)
- **MinIO**: 20Gi (PersistentVolumeClaim)

Ensure your cluster has a default StorageClass or specify one in the manifests.

## Accessing Services

### Port Forwarding (Development)

```bash
# API
kubectl port-forward -n kraftivibe-dev svc/dev-api-service 3000:3000

# PostgreSQL
kubectl port-forward -n kraftivibe-dev svc/dev-postgres-service 5432:5432

# Redis
kubectl port-forward -n kraftivibe-dev svc/dev-redis-service 6379:6379

# MinIO Console
kubectl port-forward -n kraftivibe-dev svc/dev-minio-service 9001:9001
```

### Ingress (Production)

Update DNS to point to your ingress controller's external IP:

```bash
# Get ingress IP
kubectl get ingress -n kraftivibe

# Update DNS
api.kraftivibe.com → <INGRESS_IP>
```

## Monitoring

### Health Checks

```bash
# Liveness probe
curl http://api.kraftivibe.com/health/live

# Readiness probe
curl http://api.kraftivibe.com/health/ready
```

### Logs

```bash
# API logs
kubectl logs -n kraftivibe -l app=kraftivibe-api -f

# PostgreSQL logs
kubectl logs -n kraftivibe -l app=postgres -f

# All pods
kubectl logs -n kraftivibe --all-containers -f
```

### Metrics

```bash
# Pod metrics
kubectl top pods -n kraftivibe

# Node metrics
kubectl top nodes
```

## Troubleshooting

### Pods Not Starting

```bash
# Describe pod
kubectl describe pod -n kraftivibe <pod-name>

# Check events
kubectl get events -n kraftivibe --sort-by='.lastTimestamp'

# Check logs
kubectl logs -n kraftivibe <pod-name> --previous
```

### Database Connection Issues

```bash
# Test database connectivity
kubectl exec -n kraftivibe -it <api-pod> -- /bin/sh
# Inside pod: wget -O- http://postgres-service:5432
```

### Storage Issues

```bash
# Check PVCs
kubectl get pvc -n kraftivibe

# Check PVs
kubectl get pv

# Describe PVC
kubectl describe pvc -n kraftivibe <pvc-name>
```

## Scaling

### Manual Scaling

```bash
# Scale API deployment
kubectl scale deployment -n kraftivibe kraftivibe-api --replicas=5

# Disable HPA first if enabled
kubectl delete hpa -n kraftivibe kraftivibe-api-hpa
```

### Update HPA

Edit `base/hpa.yaml` to adjust autoscaling behavior.

## Updating

### Rolling Update

```bash
# Update image tag
kubectl set image deployment/kraftivibe-api -n kraftivibe api=kraftivibe/api:v2.0.0

# Watch rollout
kubectl rollout status deployment/kraftivibe-api -n kraftivibe

# Rollback if needed
kubectl rollout undo deployment/kraftivibe-api -n kraftivibe
```

### Apply Configuration Changes

```bash
# Apply updated manifests
kubectl apply -k k8s/overlays/prod

# Restart deployment
kubectl rollout restart deployment/kraftivibe-api -n kraftivibe
```

## Cleanup

### Delete Development Environment

```bash
kubectl delete -k k8s/overlays/dev
```

### Delete Production Environment

```bash
kubectl delete -k k8s/overlays/prod
```

### Delete All Resources

```bash
kubectl delete namespace kraftivibe
kubectl delete namespace kraftivibe-dev
```

## Security Best Practices

1. **Secrets Management**: Use external secret managers (HashiCorp Vault, AWS Secrets Manager)
2. **Network Policies**: Implement network policies to restrict pod communication
3. **RBAC**: Configure role-based access control
4. **Image Scanning**: Scan container images for vulnerabilities
5. **Pod Security**: Use Pod Security Standards (restricted)
6. **TLS**: Enable TLS for all external and internal communication

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [cert-manager](https://cert-manager.io/)
