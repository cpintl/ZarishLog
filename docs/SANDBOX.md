# ZarishLog Sandbox — Beginner Guide

This guide explains how to run the ZarishLog local sandbox with a GUI-friendly workflow.

## Overview

The sandbox runs the required infrastructure locally with Docker Compose and seeds the development database with sample/master data.

## Quick start (recommended)

1. Ensure Docker Desktop / Docker Engine is installed and running.
2. From the repository root:

```bash
bash scripts/sandbox-start.sh
```

3. Open the web UI: http://localhost:3000

Or use the GUI wrapper which starts the stack and opens browser tabs:

```bash
bash scripts/sandbox-start-gui.sh
```

## Stopping and resetting

- Stop services:

```bash
bash scripts/sandbox-stop.sh
```

- Reset local data (removes volumes/local caches):

```bash
bash scripts/sandbox-reset.sh
```

- Quick health check:

```bash
bash scripts/sandbox-health.sh
```

## VS Code tasks

Open Command Palette (Ctrl+Shift+P) → Tasks: Run Task → choose one of:

- `Sandbox: Start`
- `Sandbox: Stop`
- `Sandbox: Reset`
- `Sandbox: Health`

## Screenshots

Place screenshots here to help non-technical users. Suggested images:

- VS Code Tasks menu showing `Sandbox: Start`
- Browser showing the landing page at `http://localhost:3000`
- Keycloak admin console at `http://localhost:8080`

To add a screenshot:

1. Take a screenshot using your OS shortcut.
2. Save into `docs/images/` as `sandbox-01.png`, etc.
3. Add Markdown image links below.

Example:

![Landing page](./images/sandbox-landing.png)

## Troubleshooting

- If migrations fail, ensure Docker is running and Postgres container is healthy.
- If ports are in use, edit `.env` and `docker-compose.yml` accordingly.

## Notes for maintainers

- Large artifacts such as `apps/web/.next` cache and accidental binaries were archived to `.archive/cleanup-<date>/` to keep the repo tidy for contributors.

**_ End of guide _**
