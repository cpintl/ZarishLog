# ZarishLog Branch Audit

> **Status:** Historical record. This document captures a completed audit from 2026-09-09 and is maintained for reference only.

**Audit date:** 2026-09-09
**Final mainline:** `62481f4` — `chore: update Next.js 15 and PostCSS`

## Executive result

The original set of 18 dependency branches was reviewed individually. **Twelve were merged after passing both backend and frontend CI**, while **six were closed as superseded or unsuitable for automatic merging**. The underlying work was not discarded: compatible changes were consolidated into clean pull requests, and breaking changes were documented for a separate migration effort.

All temporary remote branches associated with the reviewed pull requests have been removed. The repository now has only the `main` remote branch, with no open pull requests.

## Original dependency-branch disposition

| PRs | Change | Disposition | Reason |
|---|---|---|---|
| #5, #7, #9, #11 | GitHub Actions and pnpm action upgrades | Merged | Patch/tooling updates passed backend and frontend CI after the pnpm build-approval fix |
| #6, #8 | Go JWT and PostgreSQL driver upgrades | Merged | Compatible updates passed backend and frontend CI |
| #10 | Go validator `10.30.4` | Closed as superseded | Included in merged PR #24 with the current Go dependency graph |
| #12 | Go Testify `1.12.1` | Closed as superseded | Included in merged PR #24 with the current Go dependency graph |
| #13 | React Native `0.87.1` | Merged | Current CI passed; mobile dependency update does not alter the web runtime |
| #14 | Vitest `5.0.0` | Merged | Current unit tests and CI passed |
| #15 | Zod `4.5.4` | Merged | Current type-check, tests, build, and CI passed |
| #16 | Dexie `4.4.5` | Merged | Patch update passed CI |
| #17 | Grouped Next/PostCSS/Vitest update | Closed as superseded | Overlapped with separately merged Vitest work and the later safe Next/PostCSS consolidation |
| #18 | Testing Library React `16.3.3` | Merged | Current CI passed |
| #19 | Tailwind CSS `4.3.3` | Closed, migration required | Major upgrade requires a Tailwind configuration and utility-class audit; current Tailwind 3 stack remains verified |
| #20 | Dexie React Hooks `4.4.0` | Merged | Current CI passed after the mainline dependency updates |
| #21 | React DOM `19.2.8` | Closed as superseded | The isolated branch failed because `react` remained `19.2.7`; merged PR #24 upgraded React and React DOM together |
| #22 | Next.js `16.3.4` | Closed, migration required | Major upgrade requires a dedicated `next-pwa`, ESLint, and app-router compatibility review; verified Next.js 15 line retained |

## Consolidation pull requests

| PR | Purpose | Result |
|---|---|---|
| #23 | Approve `unrs-resolver` build scripts in pnpm CI installs | Merged |
| #24 | Consolidate compatible Go updates and aligned React/React DOM `19.2.8` | Merged |
| #26 | Consolidate Next.js `15.5.24` and PostCSS `8.5.23` | Merged |

## Final verified dependency baseline

| Area | Final line |
|---|---|
| Next.js | `15.5.25` resolved by the `^15.5.24` declaration |
| React / React DOM | `19.2.8` aligned |
| Vitest | `5.0.0` |
| Zod | `4.5.4` |
| Dexie | `4.4.5` |
| Dexie React Hooks | `4.4.0` |
| PostCSS | `8.5.23` declaration; current lockfile resolution verified |
| Go validator | `10.30.4` |
| Go Testify | `1.12.1` |
| Go JWT | `5.3.1` |
| PostgreSQL driver | `lib/pq 1.12.3` |
| CI actions | checkout `v7`, setup-go `v7`, setup-node `v7`, pnpm/action-setup `v6` |

## Final verification

The consolidated mainline passed the repository CI jobs for the final upgrade pull requests. The frontend quality gate passed linting, TypeScript checking, four Vitest tests, and a production Next.js build. The backend CI job passed its PostgreSQL-backed vet, race-enabled tests, and build checks. The sandbox itself does not have the Go toolchain installed, so backend verification was intentionally delegated to the GitHub runner.

## Remaining maintenance work

Tailwind CSS 4 and Next.js 16 should be treated as separate migration projects rather than automatic dependency updates. They require source-level compatibility work, visual regression testing, and a controlled rollout. GitHub’s default branch continues to report a dependency vulnerability backlog; that backlog should be triaged separately from the completed branch consolidation.
