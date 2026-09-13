---
description: Start the development environment with Docker services and dev servers
agent: build
---

Start the ZarishLog development environment:

1. Start Docker services (PostgreSQL, Redis, MinIO, Keycloak, Meilisearch)
2. Wait for services to be healthy
3. Start the Go API server on port 8080
4. Start the Next.js web server on port 3000
5. Display access URLs and health status

Use `make dev` to start all services, or follow the step-by-step approach if specific services need attention.

After starting, verify:
- PostgreSQL is accessible on localhost:5432
- Redis is accessible on localhost:6379
- API server responds on localhost:8080
- Web server responds on localhost:3000

If any service fails to start, check logs with `docker compose logs <service>` and report the issue.