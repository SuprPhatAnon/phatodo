# Deployment

## Container

Build the server image from the repository root:

```sh
make docker-build
```

The image contains both binaries, but runs `phatodo-server` by default. Configure it with:

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

Before applying, edit:

- `deploy/k3s/app.yaml` image name
- `deploy/k3s/ingress.yaml` host
- `deploy/k3s/cert-issuer.yaml` email
- `deploy/k3s/postgres-secret.yaml` password and URL

Apply with:

```sh
make k3s-render
make deploy-k3s
```
