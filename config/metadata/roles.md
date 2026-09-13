# Roles

ZarishLog ships with twelve built-in roles. Each role maps to a `level` — the
depth in the organization hierarchy at which it is expected to operate:

| Level | Scope                                                        |
| ----- | ------------------------------------------------------------ |
| 1     | Global — the full platform, all tenants (organizations)      |
| 2     | Country / thematic — one organization or program across sites |
| 3     | Warehouse / department HQ — central store and office ops     |
| 4     | Department / field — sub-warehouse and distribution points    |

| Code | Name                  | Level | Purpose                                                              |
| ---- | --------------------- | ----- | -------------------------------------------------------------------- |
| R01  | GLOBAL_ADMIN          | 1     | Full system administration across all tenants                         |
| R02  | COUNTRY_REP           | 2     | Country representative — read-only, reporting, no operations          |
| R03  | THEME_MANAGER         | 2     | Thematic program manager — read + reports within program scope        |
| R04  | WAREHOUSE_OFFICER     | 3     | Full central warehouse operations — procure, QA, count                 |
| R05  | WAREHOUSE_STOREKEEPER | 3     | Stock operations — receive, issue, count, read                        |
| R06  | ADMIN_LOG_OFFICER     | 3     | Office asset and logistics stock management                           |
| R07  | DEPT_MANAGER          | 3     | Department head — stock read, reports, approval authority             |
| R08  | DEPT_COORDINATOR      | 4     | Sub-warehouse stock flow validation                                   |
| R09  | DEPT_OFFICER          | 4     | Day-to-day sub-warehouse operations                                   |
| R10  | FIELD_WORKER          | 4     | Field distribution staff — distribution create/read                   |
| R11  | QUALITY_OFFICER       | 3     | QA inspection staff — full QA, stock read                             |
| R12  | AUDITOR               | 2     | Read-only audit access — read + audit + reports                       |

Roles are seeded in `packages/data-models/sql/seed.sql` and composed with
permissions (`module` × `action`) to build access policies. The mapping of
roles to Keycloak realm roles is defined in the Keycloak realm export under
`infrastructure/keycloak/`.

This file is imported by the configuration validator; keep the table above in
sync with `seed.sql` when adding or renaming roles. Reference: `R01`–`R12`.