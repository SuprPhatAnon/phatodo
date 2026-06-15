# Deployment

## Container

Build the server image from the repository root:

```sh
make docker-build
```

The image contains both binaries and defaults to `phatodo-server`. The Makefile tags it as `$(REGISTRY)/phatodo:$(TAG)`, which defaults to `10.80.0.85:30500/phatodo:latest`. Use `TAG=...` or `REGISTRY=...` to override that if needed.

Push the image to the cluster-local registry with:

```sh
make docker-push
```

Configure the server with:

- `PHATODO_ADDR`, default `:8080`
- `PHATODO_DATABASE_URL`, a Postgres connection string

## Local Compose

Run the API server and Postgres locally:

```sh
make compose-up
```

Compose exposes the server at `http://localhost:8080` and Postgres at `localhost:5432`. The Postgres container loads SQL files from `migrations/` when the data volume is first created.

## k3s

The manifests in `deploy/k3s/` deploy:

- `phatodo-server` Deployment and Service
- Postgres StatefulSet and Service
- Postgres Secret
- Traefik Ingress
- cert-manager `Issuer` for Let's Encrypt TLS

Run database migrations separately after Postgres is available. The current server image includes `migrations/`, but no migration runner has been implemented yet.

The checked-in k3s bundle targets the `phatodo` namespace and the default image tag from the Makefile. If you change `REGISTRY`, `TAG`, or `KUBE_NAMESPACE`, use the same values when deploying.

Apply and wait for rollout with:

```sh
make deploy-k3s
```
