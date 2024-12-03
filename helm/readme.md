# Rekor Helm Chart

## Prerequisites
- Kubernetes 1.19+
- Helm 3.x
- Persistent Volume support in the cluster

## Installation

### Add Helm Repository (Optional)
```bash
helm repo add rekor https://sigstore.github.io/helm-charts
helm repo update
```

### Install the Chart
```bash
helm install rekor ./rekor \
  --namespace rekor \
  --create-namespace
```

### Customize Deployment
Create a `values.yaml` to override default configurations:

```yaml
mysql:
  rootPassword: your-secure-password
  
redis:
  password: your-redis-password

rekorServer:
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
```

Then install with:
```bash
helm install rekor ./rekor -f values.yaml
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `mysql.rootPassword` | MySQL root password | `zaphod` |
| `redis.password` | Redis authentication password | `test` |
| `rekorServer.resources` | Kubernetes resource requests/limits | See `values.yaml` |

## Uninstallation
```bash
helm uninstall rekor -n rekor
```

## Troubleshooting
- Ensure sufficient cluster resources
- Check pod logs with `kubectl logs`
- Verify persistent volume configurations