# Deployment

## Container

Build the server image from the repository root:

```sh
make docker-build
```

The image contains both binaries and defaults to `phatodo-server`. The Makefile tags it as `$(REGISTRY)/phatodo:$(TAG)`, which defaults to `10.80.0.85:30500/phatodo:latest`.

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
- Traefik Ingress over plain HTTP by default

TLS is optional. The checked-in k3s bundle does not require cert-manager or a certificate issuer to be present. If you want HTTPS later, add a separate TLS overlay or manifest rather than changing the default deployment path.

Run database migrations separately after Postgres is available. The current server image includes `migrations/`, but no migration runner has been implemented yet.

The checked-in k3s bundle targets the `phatodo` namespace and pulls `localhost:30500/phatodo:$(TAG)` by default. The Makefile still builds and pushes the image to `$(REGISTRY)/phatodo:$(TAG)` for the registry side of the workflow. If you change `REGISTRY`, `TAG`, or `KUBE_NAMESPACE`, use the same values when deploying.

Apply and wait for rollout with:

```sh
make deploy-k3s
```

The deploy target creates `$(KUBE_NAMESPACE)` first with an idempotent `kubectl create namespace ... --dry-run=client -o yaml | kubectl apply -f -` step, then applies the k3s bundle, updates the `phatodo-server` image to `$(KUBE_IMAGE)`, and waits for the rollout to finish.
