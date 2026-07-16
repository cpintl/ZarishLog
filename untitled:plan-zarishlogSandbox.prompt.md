## Plan: Harden ZarishLog into a GUI-friendly local sandbox

### TL;DR

The workspace already has a strong foundation, but it is not yet a reliable one-click sandbox. The main problems are inconsistent versions, weak bootstrap automation, missing dependency-aware startup flow, fragmented documentation, and a few noisy or stale artifacts. The plan is to turn this into a Linux-first, GUI-friendly dev sandbox with one-click startup, clear docs, and safer defaults.

### Findings

- The bootstrap script and Make targets are present, but they do not create a fully reliable startup path from a cold state.
- A live verification showed that database seeding fails when PostgreSQL is not already running, which confirms the current flow is incomplete.
- The workspace mixes version expectations across files: the setup script pins Go 1.26.4 and pnpm 11, while CI uses Go 1.26 and pnpm 9, and the top-level package.json expects pnpm 9.12.0.
- The VS Code tasks exist, but the sandbox experience is still command-line oriented and lacks a simple “start everything” workflow and clear post-start guidance.
- The MCP setup is present but depends on external packages and environment variables and is not documented as a first-class sandbox feature.
- The repository contains several stale or noisy artifacts, including a root-level binary-like file named api and a large dependency tree under the workspace root that should be normalized.

### Implementation phases

1. Stabilize the bootstrap path

- Standardize the pinned tool versions across scripts/zarishlog-setup.sh, Makefile, package.json, and .github/workflows/ci.yml.
- Make the bootstrap flow idempotent and self-healing: detect missing tools, install or warn clearly, start Docker if possible, and then run database setup automatically.
- Add a top-level “sandbox start” workflow that performs prerequisites, Docker startup, migrations, seeding, and service readiness checks in sequence.

2. Make the workspace GUI-friendly

- Add a simple one-click entrypoint for non-technical users via VS Code tasks and a dedicated startup script.
- Improve .vscode/tasks.json and .vscode/launch.json with tasks for “Start sandbox”, “Stop sandbox”, “Open services”, and “Reset sandbox”.
- Ensure the VS Code environment points to the correct workspace folder, uses the right terminal profile, and exposes the local service URLs clearly.

3. Harden local service orchestration

- Review docker-compose.yml for a clean local service lifecycle, health checks, and stable port mappings.
- Make database initialization more robust by waiting for PostgreSQL readiness and retrying migrations instead of assuming the DB is already ready.
- Add a clear service health summary (DB, API, web, Keycloak, MinIO, Meilisearch) at the end of startup.

4. Improve documentation and onboarding

- Replace scattered setup guidance with a single beginner-friendly sandbox guide covering prerequisites, first run, common troubleshooting, and service URLs.
- Update README.md, SETUP.md, and CONFIGURE.md to reflect the new one-click flow and remove contradictory instructions.
- Document the MCP configuration, environment variables, and expected dependencies in a concise “sandbox features” section.

5. Clean up repository noise

- Remove or archive stale artifacts that are not needed for day-to-day development.
- Normalize dependency installation so the workspace does not carry large duplicated install trees in the repo root.
- Review hidden config files in .opencode and .vscode and keep only what is actively needed for the sandbox experience.

### Relevant files

- Makefile
- scripts/zarishlog-setup.sh
- docker-compose.yml
- .env.example
- .vscode/tasks.json
- .vscode/launch.json
- .vscode/extensions.json
- .opencode/mcp.json
- .github/workflows/ci.yml
- README.md
- SETUP.md
- CONFIGURE.md

### Verification

- Run the bootstrap flow from a clean machine state and confirm that the sandbox reaches a healthy state without manual intervention.
- Verify that the one-click startup flow works end to end: Docker starts, migrations run, seed data loads, API and web boot, and the health endpoints respond.
- Validate that documentation matches the actual commands and output.
