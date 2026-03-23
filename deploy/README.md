# VPS Deployment

This repository now verifies the Go and Vue projects, builds Docker images in GitHub Actions, pushes them to Docker Hub, and then deploys on the VPS by pulling those images.

## Required GitHub Secrets

- `DOCKERHUB_USERNAME`: Docker Hub login username
- `DOCKERHUB_TOKEN`: Docker Hub access token
- `VPS_HOST`: VPS IP or domain
- `VPS_USER`: SSH user on the VPS
- `VPS_SSH_KEY`: private key for the deploy user
- `VPS_PORT`: optional, defaults to `22`
- `VPS_DEPLOY_PATH`: optional, defaults to `/opt/xiaomaipro`

## Optional GitHub Variables

- `DOCKERHUB_NAMESPACE`: Docker Hub namespace or organization name
- `DEPLOY_PROFILE`: set to `1g` for the lightweight VPS profile

If `DOCKERHUB_NAMESPACE` is not set, the workflow uses `DOCKERHUB_USERNAME`.

## VPS Requirements

- Docker and Docker Compose installed
- `rsync` available on the VPS
- The deploy user can run Docker commands

## Runtime Configuration

1. First deployment creates `deploy/.env` from `deploy/.env.example`
2. Update `deploy/.env` on the VPS with real runtime secrets and domain values
3. GitHub Actions only syncs the `deploy/` directory to the VPS
4. Each deployment pulls the tagged application images from Docker Hub and restarts the stack

## 1G VPS

For a `1 vCPU / 1 GB RAM` VPS, use the lightweight override:

- `deploy/docker-compose.1g.yml`
- `deploy/scripts/deploy-1g.sh`

This profile disables Kafka and RabbitMQ, keeps Redis/PostgreSQL on smaller settings, and relies on the order service local async fallback instead of Kafka.

## Database Initialization

- Put initial PostgreSQL schema or seed files into `deploy/postgres/init/`
- PostgreSQL only runs those scripts when the data volume is created for the first time
