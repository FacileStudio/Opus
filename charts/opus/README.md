# Opus Helm chart

Deploys [Opus](https://opus.facile.studio) on Kubernetes: one Deployment running the Opus image
(the Go API and the built React client in one container, port 80), its Service, and an optional
bundled PostgreSQL 16. Ingress and Gateway API routes are optional.

Production on opus.facile.studio runs the Docker Compose file, not this chart. See
[docs/deployment.md](../../docs/deployment.md).

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- A PV provisioner, when persistence is enabled

## Install

```bash
helm install opus ./charts/opus --namespace opus --create-namespace \
  --set opus.env.clientUrl=https://opus.example.com
kubectl port-forward svc/opus-opus 8080:80 -n opus
```

Sign-in goes through a Registre (OIDC) issuer only. Opus has no password or sign-up route, so
without the `OIDC_*` variables nobody can sign in. Pass them through `opus.extraEnv`:

```yaml
opus:
  env:
    clientUrl: https://opus.example.com
  extraEnv:
    - name: OIDC_ISSUER
      value: https://sso.example.com
    - name: OIDC_CLIENT_ID
      value: opus
    - name: OIDC_REDIRECT_URL
      value: https://opus.example.com/api/auth/oidc/callback
    - name: OIDC_CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: opus-oidc
          key: client-secret
```

Every variable the app reads is in [docs/configuration.md](../../docs/configuration.md).

Uninstall with `helm uninstall opus -n opus`.

## What the chart sets on the container

| Variable | From |
|---|---|
| `OPUS_CLIENT_URL` | `opus.env.clientUrl`. The entrypoint derives `OPUS_API_URL` as `<clientUrl>/api` |
| `CORS_ORIGINS` | `opus.env.corsOrigins` (string or list), else `opus.env.clientUrl` |
| `PORT` | `opus.service.targetPort` |
| `DATABASE_URL` | The bundled PostgreSQL, or `opus.env.database.external.*` |
| anything else | `opus.extraEnv`, appended last, so a duplicate name overrides |

Migrations run when the app starts. With a role that may create databases, the app also creates
the database named in `DATABASE_URL` when it is missing.

## Parameters

### Global

| Name | Description | Default |
|---|---|---|
| `nameOverride` | Partially override the fullname | `""` |
| `fullnameOverride` | Fully override the fullname | `""` |
| `replicaCount` | Replicas, ignored with autoscaling | `1` |
| `autoscaling.enabled` | Create a HorizontalPodAutoscaler | `false` |
| `autoscaling.minReplicas` / `maxReplicas` | Replica bounds | `1` / `10` |
| `autoscaling.targetCPUUtilizationPercentage` | CPU target | `80` |
| `serviceAccount.create` / `name` / `annotations` | Service account | `true` / `""` / `{}` |
| `podAnnotations`, `podSecurityContext`, `nodeSelector`, `tolerations`, `affinity` | Pod settings | empty |

More than one replica needs `REDIS_URL` in `opus.extraEnv` for realtime updates, and shared
storage for uploads.

### Opus container

| Name | Description | Default |
|---|---|---|
| `opus.image.repository` | Image | `ghcr.io/facilestudio/opus` |
| `opus.image.tag` | Tag, falls back to the chart's `appVersion` when empty | `latest` |
| `opus.image.pullPolicy` | Pull policy | `IfNotPresent` |
| `opus.securityContext` | Container security context | `{}` |
| `opus.service.type` / `port` / `targetPort` | Service | `ClusterIP` / `80` / `80` |
| `opus.resources` | Requests and limits | `{}` |
| `opus.persistence.enabled` | Mount a PVC at `/app/data/uploads` for logos and local assets | `false` |
| `opus.persistence.size` / `storageClass` / `accessMode` | Uploads PVC | `1Gi` / `""` / `ReadWriteOnce` |
| `opus.env.clientUrl` | Public URL of the instance | `""` |
| `opus.env.corsOrigins` | Allowed origins, string or list | `[]` |
| `opus.env.database.external.enabled` | Use an external PostgreSQL (set `postgresql.enabled=false`) | `false` |
| `opus.env.database.external.host` / `port` / `database` / `username` / `password` | External database | `""` / `5432` / `opus` / `opus_user` / `""` |
| `opus.env.database.external.existingSecret.enabled` / `name` / `passwordKey` | Read the whole connection URI from a secret | `false` / `""` / `postgres_uri` |
| `opus.extraEnv` | Extra `EnvVar` entries | `[]` |
| `opus.livenessProbe` / `readinessProbe` | Probes, both on `/api/health` | see `values.yaml` |

### Bundled PostgreSQL

| Name | Description | Default |
|---|---|---|
| `postgresql.enabled` | Deploy PostgreSQL in the release | `true` |
| `postgresql.image.repository` / `tag` / `pullPolicy` | Image | `postgres` / `16-alpine` / `IfNotPresent` |
| `postgresql.auth.database` / `username` / `password` | Credentials. Change the password | `opus` / `opus_user` / `opus_password` |
| `postgresql.auth.existingSecret` | Secret holding the PostgreSQL password, under `secretKeys.userPasswordKey` (`password`). The app's `DATABASE_URL` is still built from `postgresql.auth.password`, so keep the two equal | `""` |
| `postgresql.persistence.enabled` / `size` / `storageClass` / `accessMode` | Data PVC | `true` / `8Gi` / `""` / `ReadWriteOnce` |
| `postgresql.service.type` / `port` | Service | `ClusterIP` / `5432` |
| `postgresql.resources` | Requests and limits | `{}` |

### Ingress and Gateway API

| Name | Description | Default |
|---|---|---|
| `ingress.enabled` | Create an Ingress | `false` |
| `ingress.className` | Ingress class | `nginx` |
| `ingress.annotations` / `tls` | Annotations and TLS | `{}` / `[]` |
| `ingress.hosts` | Hosts and paths. `service: opus` expands to the release's Opus Service | `opus.local`, `/` |
| `gateway.enabled` | Create an `HTTPRoute` | `false` |
| `gateway.parentRefs` / `hostnames` | Gateway listeners and hostnames | `[]` / `[]` |
| `gateway.annotations` / `labels` | Route metadata | `{}` / `{}` |
| `gateway.rules` | Rules. `service: opus` expands the same way | `/` to the Opus Service |

Route every path to the Opus Service on port 80. The binary serves `/api`, `/mcp`, `/ws` and the
client itself, so the chart needs no path rewrite and no second backend.

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: opus.example.com
      paths:
        - path: /
          pathType: Prefix
          service: opus
          port: 80
  tls:
    - secretName: opus-tls
      hosts:
        - opus.example.com
```

## Backups

```bash
kubectl exec deployment/opus-postgresql -n opus -- pg_dump -U opus_user --format=custom opus > opus.dump
```
