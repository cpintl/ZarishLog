# ZarishLog — No-Code Configuration Guide

> **Purpose:** This document lets non-developers configure ZarishLog's entire system behavior through CSV files, Markdown tables, and VS Code GUI interactions — no coding required.

---

## Table of Contents

1. [Configuration Philosophy](#1-configuration-philosophy)
2. [Quick Start — 5 Minute Setup](#2-quick-start--5-minute-setup)
3. [VS Code GUI Configuration Center](#3-vs-code-gui-configuration-center)
4. [Master Data Configuration (CSV)](#4-master-data-configuration-csv)
5. [Organizational Hierarchy (Markdown)](#5-organizational-hierarchy-markdown)
6. [Product Catalogue (CSV)](#6-product-catalogue-csv)
7. [User Roles & Permissions (Markdown)](#7-user-roles--permissions-markdown)
8. [Warehouse & Location Setup (JSON)](#8-warehouse--location-setup-json)
9. [System Settings (Environment Variables)](#9-system-settings-environment-variables)
10. [Advanced Configuration](#10-advanced-configuration)
11. [Configuration Workflow Diagram](#11-configuration-workflow-diagram)
12. [Validation & Testing](#12-validation--testing)

---

## 1. Configuration Philosophy

ZarishLog uses **Configuration-as-Data** — all system behavior is driven by data files (CSV, JSON, Markdown, YAML) that you edit without touching code.

### Configuration Layers

```
┌─────────────────────────────────────────────┐
│  Layer 1: Master Data (CSV / Markdown)       │ ← You edit these
│  - Product catalogue, org hierarchy, roles   │
├─────────────────────────────────────────────┤
│  Layer 2: Infrastructure (docker-compose.yml)│ ← Pre-configured
│  - Postgres, Redis, Keycloak, Meilisearch    │
├─────────────────────────────────────────────┤
│  Layer 3: Environment (.env)                 │ ← You customize
│  - Ports, URLs, secrets, localization        │
├─────────────────────────────────────────────┤
│  Layer 4: Business Rules (Go code)           │ ← Pre-written
│  - FEFO, AMC, reorder point calculations     │
└─────────────────────────────────────────────┘
```

---

## 2. Quick Start — 5 Minute Setup

### Step 1: Open Configuration Center

```bash
# Option A: VS Code (Recommended)
code .vscode/zarishlog.code-snippets

# Option B: Browser-based
# Open http://localhost:3000/admin/setup after starting the app
```

### Step 2: Fill Out Your Organization Details

Edit `config/organization.csv`:

```csv
org_name,org_code,level_1_name,level_2_name,level_3_name,level_4_name
Center for Peace and Integrity,CPI,Global HQ,Country Office,Project Office,Program Site
```

### Step 3: Load Your Product Catalogue

Edit `config/metadata/master_product_list.csv` and run:

```bash
make db-seed
```

### Step 4: Configure Your Warehouse

Edit `config/location/warehouse.json`:

```json
{
  "warehouses": [
    {
      "name": "My Central Warehouse",
      "code": "MY-CWH",
      "type": "central",
      "city": "My City",
      "country": "My Country"
    }
  ]
}
```

### Step 5: Start the System

```bash
make dev
```

---

## 3. VS Code GUI Configuration Center

ZarishLog provides a **visual configuration interface** through VS Code tasks and snippets.

### 3.1 Configuration Tasks

Press `Ctrl+Shift+P` → **Tasks: Run Task** and choose:

| Task Name | What It Does |
|-----------|-------------|
| **Configure: Organization** | Opens the org hierarchy CSV editor |
| **Configure: Product Catalogue** | Opens the master product list CSV |
| **Configure: Warehouse** | Opens warehouse JSON editor |
| **Configure: User Roles** | Opens role permissions markdown |
| **Configure: System Settings** | Opens .env file for editing |
| **Validate Configuration** | Runs all config validation checks |
| **Apply Configuration** | Seeds/updates database with your config |

### 3.2 Configuration Snippets

Type these prefixes in any file and press `Tab`:

| Prefix | Inserts |
|--------|---------|
| `z-org` | Organization hierarchy template |
| `z-warehouse` | Warehouse + location template |
| `z-product` | Product/item template |
| `z-role` | Role/permission template |
| `z-uom` | Unit of measure template |

### 3.3 Visual Configuration Dashboard (Web UI)

Once the app is running, visit:

- **http://localhost:3000/admin/settings** — System settings
- **http://localhost:3000/admin/catalogue** — Catalogue management
- **http://localhost:3000/admin/users** — User management

---

## 4. Master Data Configuration (CSV)

### 4.1 Organization Hierarchy

File: `config/metadata/organization.csv`

```csv
name,code,level,parent_code
Center for Peace and Integrity,CPI,1,
Country Office - Bangladesh,CPI-BD,2,CPI
Project Office - Cox Bazar,CPI-CXB,3,CPI-BD
Camp 5 Health Post,CPI-CXB-C5,4,CPI-CXB
```

**Rules:**
- `level` must be 1-4 (L1 = Global, L2 = Country, L3 = Project, L4 = Site)
- `parent_code` must match an existing entry's `code`
- All codes must be unique

### 4.2 Programs / Thematic Areas

File: `config/metadata/programs.csv`

```csv
code,name,description
H&N,Health and Nutrition,Medical and nutrition programs
WASH,Water Sanitation and Hygiene,Clean water and sanitation
LVL,Livelihood,Economic empowerment programs
EDU,Education,Formal and non-formal education
EPRR,Emergency Preparedness,Disaster response readiness
R&L,Research and Learning,M&E and research activities
```

### 4.3 Units of Measure

File: `config/metadata/uom.csv`

```csv
name,abbreviation,category
Each,EA,count
Box,BX,count
Carton,CTN,count
Kilogram,KG,weight
Liter,L,volume
```

**Categories:** `count`, `weight`, `volume`, `length`, `time`, `area`

---

## 5. Organizational Hierarchy (Markdown)

For visual configuration, edit `docs/org-chart.md`:

```markdown
# Organization: Center for Peace and Integrity (CPI)

## L1 - Global
- **Global HQ** (CPI-GHQ)
  - Location: [Your City]

### L2 - Country Offices
- **Bangladesh** (CPI-BD)
  - Country Director: [Name]
  - Programs: H&N, WASH, EDU

#### L3 - Project Offices
- **Cox's Bazar Project** (CPI-CXB)
  - Project Manager: [Name]
  - Central Warehouse: CXB-CWH

##### L4 - Program Sites
- **Camp 5 Health Post** (CPI-CXB-C5)
  - Sub-Warehouse: CXB-C5-SWH
  - Staff: [Count]
```

> **Note:** The markdown file is for human readability. The actual data is loaded from the CSV files.

---

## 6. Product Catalogue (CSV)

The master product catalogue is the heart of ZarishLog. Edit `config/metadata/master_product_list.csv`:

### Required Fields

```csv
sku,name,category_name,uom_abbreviation,item_type
MED-AMOX-500,Amoxicillin 500mg Capsules,Medicines & Drugs,EA,drug
SUP-LTX-MED,Latex Examination Gloves,Medical Supplies,BX,medical_supply
```

### All Available Fields

| Field | Required | Type | Options | Description |
|-------|----------|------|---------|-------------|
| `sku` | Yes | Text | Unique | Stock Keeping Unit identifier |
| `name` | Yes | Text | — | Product display name |
| `category_name` | Yes | Text | Must match categories CSV | Product category |
| `uom_abbreviation` | Yes | Text | Must match UoM CSV | Unit of measure |
| `item_type` | Yes | Enum | `drug`, `medical_supply`, `equipment`, `instrument`, `material`, `vaccine`, `nutrition`, `lab_reagent`, `asset`, `consumable` | Item classification |
| `description` | No | Text | — | Product description |
| `gtin` | No | Text | — | Global Trade Item Number (barcode) |
| `brand` | No | Text | — | Manufacturer brand |
| `manufacturer` | No | Text | — | Manufacturer name |
| `is_batch_tracked` | No | Boolean | `TRUE` / `FALSE` | Track by batch/lot |
| `is_serial_tracked` | No | Boolean | `TRUE` / `FALSE` | Track individual serial numbers |
| `is_expiry_tracked` | No | Boolean | `TRUE` / `FALSE` | Track expiration dates |
| `is_hazardous` | No | Boolean | `TRUE` / `FALSE` | Hazardous material |
| `is_cold_chain` | No | Boolean | `TRUE` / `FALSE` | Requires cold storage |
| `min_stock` | No | Number | >= 0 | Minimum stock threshold |
| `max_stock` | No | Number | >= 0 | Maximum stock level |
| `reorder_point` | No | Number | >= 0 | Auto-reorder trigger level |
| `lead_time_days` | No | Integer | >= 0 | Days to replenish |
| `unit_cost` | No | Number | >= 0 | Cost per unit |

### Sample Catalogue

A sample catalogue with 9 items is pre-seeded. To load your own:

1. Replace `config/metadata/master_product_list.csv` with your data
2. Run `make db-seed`
3. The system validates: duplicates, required fields, category/UoM existence

---

## 7. User Roles & Permissions (Markdown)

Edit `config/metadata/roles.md`:

```markdown
# Role Configuration

## Available Roles (R01-R09)

| Role ID | Name | Level | Can View | Can Edit | Can Approve | Can Admin |
|---------|------|-------|----------|----------|-------------|-----------|
| R01 | Global Admin | L1 | ✓ All | ✓ All | ✓ All | ✓ All |
| R02 | Country Rep | L2 | ✓ Country | ✗ | ✗ | ✗ |
| R03 | Theme Manager | L2 | ✓ Theme | ✗ | ✗ | ✗ |
| R04 | Warehouse Officer | L3 | ✓ Warehouse | ✓ Transactions | ✓ Adjustments | ✗ |
| R05 | Warehouse Storekeeper | L3 | ✓ Warehouse | ✓ Receive/Issue | ✗ | ✗ |
| R06 | Admin Logistics Officer | L3 | ✓ Office | ✓ Assets | ✗ | ✗ |
| R07 | Dept Manager | L3 | ✓ Department | ✓ Approvals | ✓ Budget | ✓ Department |
| R08 | Dept Coordinator | L4 | ✓ Site | ✓ Validate | ✗ | ✗ |
| R09 | Dept Officer | L4 | ✓ Site | ✓ Operations | ✗ | ✗ |

## Custom Roles

To add a custom role, edit `packages/data-models/sql/seed.sql` and add:

```sql
INSERT INTO roles (code, name, description, level) VALUES
  ('R10', 'CUSTOM_ROLE', 'Description of custom role', 3);
```

## Permission Matrix

| Module | Actions |
|--------|---------|
| Products | create, read, update, delete |
| Categories | create, read, update, delete |
| Warehouses | create, read, update, delete |
| Stock | receive, issue, transfer, adjust, read |
| QA | inspect, read |
| Assets | create, read, update, transfer |
| Users | create, read, update |
| Reports | read |
```

---

## 8. Warehouse & Location Setup (JSON)

File: `config/location/warehouse.json`

```json
{
  "warehouse": {
    "name": "Cox Bazar Central Warehouse",
    "code": "CXB-CWH",
    "type": "central",
    "address": "Main Logistics Hub, Cox Bazar",
    "city": "Cox Bazar",
    "country": "Bangladesh",
    "is_cold_chain": false
  },
  "locations": [
    { "code": "RECV", "name": "Receiving Area", "type": "area" },
    { "code": "GEN-A", "name": "General Storage A", "type": "zone", "parent": "RECV" },
    { "code": "GEN-B", "name": "General Storage B", "type": "zone" },
    { "code": "COLD", "name": "Cold Chain Storage", "type": "zone", "is_cold_chain": true },
    { "code": "QA", "name": "QA/Quarantine", "type": "area", "is_secure": true },
    { "code": "DISP", "name": "Dispatch Area", "type": "area" }
  ]
}
```

### Location Types

| Type | Purpose | Naming Convention |
|------|---------|-------------------|
| `area` | Functional area (Receiving, Dispatch) | 4-letter uppercase |
| `zone` | Storage zone (A, B, Cold Chain) | PREFIX-ZONE |
| `rack` | Rack within zone | ZONE-R-## |
| `bin` | Bin within rack | RACK-BIN-## |
| `shelf` | Shelf within bin | BIN-S-# |

---

## 9. System Settings (Environment Variables)

Edit `.env` to configure system behavior:

### Essential Settings

```bash
# Default timezone
DEFAULT_TIMEZONE="Asia/Dhaka"          # Change to your timezone

# Date/time format (GMT+6 default)
DEFAULT_LOCALE="en-BD"                 # Change to your locale

# API and Web ports (change if ports conflict)
API_PORT=8080
WEB_PORT=3000
```

### Timezone Reference

```bash
# Common timezones:
# South Asia:    Asia/Dhaka, Asia/Kolkata, Asia/Karachi
# Africa:        Africa/Nairobi, Africa/Addis_Ababa, Africa/Johannesburg
# Middle East:   Asia/Dubai, Asia/Riyadh, Asia/Amman
# Europe:        Europe/London, Europe/Berlin, Europe/Paris
# Americas:      America/New_York, America/Chicago, America/Los_Angeles
```

---

## 10. Advanced Configuration

### 10.1 Business Rules

Business rules are pre-written Go code in `packages/business-logic/`. To customize:

| Rule | File | Default Behavior |
|------|------|-----------------|
| **FEFO** | `fefo.go` | Earliest expiry first |
| **AMC** | `amc.go` | 3-month rolling average |
| **Reorder Point** | `amc.go` | AMC × lead time / 30 + safety stock |

### 10.2 Notification Rules (Coming Soon)

Configure email/SMS alerts for:
- Low stock (triggered by `reorder_point`)
- Expiring items (X days before expiry)
- QA failures
- Unauthorized adjustments

### 10.3 Approval Workflows (Coming Soon)

Configure multi-level approval chains for:
- Stock adjustments above threshold
- Asset transfers
- New user registration

---

## 11. Configuration Workflow Diagram

```
┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐
│  Edit CSV/MD     │ ──→ │  Validate Config │ ──→ │  Apply to DB     │
│  (config/*)      │     │  (make validate) │     │  (make db-seed)  │
└──────────────────┘     └──────────────────┘     └──────────────────┘
         │                        │                        │
         ▼                        ▼                        ▼
  VS Code snippets          Error reporting          System restarts
  + GUI dashboard          + fix suggestions         with new config
```

### Validation Workflow

```bash
# Step 1: Edit config files
code config/metadata/master_product_list.csv

# Step 2: Validate
bash scripts/validate-config.sh

# Step 3: Fix any errors shown

# Step 4: Apply
make db-seed

# Step 5: Verify
curl http://localhost:8080/api/v1/products
```

---

## 12. Validation & Testing

### 12.1 Configuration Validation Script

```bash
bash scripts/validate-config.sh
```

The validator checks:
- ✓ All CSV files have correct headers
- ✓ No duplicate SKUs in product catalogue
- ✓ All category references exist
- ✓ All UoM references exist
- ✓ Org hierarchy is complete (no orphan nodes)
- ✓ Location codes are unique per warehouse

### 12.2 Visual Validation Dashboard

Once the app is running:
- **http://localhost:3000/admin/validate** — Configuration validation dashboard
- Shows green checkmarks or red errors for each config file
- Provides one-click fix suggestions

### 12.3 Run Tests

```bash
# Validate all configs
bash scripts/validate-config.sh

# Run Go tests (includes FEFO, AMC validation)
cd apps/api && go test ./... -v

# Run frontend tests
cd apps/web && pnpm test
```

---

## Appendix: Configuration Files Reference

| File Path | Format | Purpose | Edit Method |
|-----------|--------|---------|-------------|
| `config/metadata/organization.csv` | CSV | Organization hierarchy | VS Code / Spreadsheet |
| `config/metadata/programs.csv` | CSV | Program/thematic areas | VS Code / Spreadsheet |
| `config/metadata/uom.csv` | CSV | Units of measure | VS Code / Spreadsheet |
| `config/metadata/master_product_list.csv` | CSV | Product catalogue | VS Code / Spreadsheet |
| `config/metadata/roles.md` | Markdown | Role definitions | VS Code |
| `config/location/warehouse.json` | JSON | Warehouse + locations | VS Code |
| `config/templates/goods_receipt_form.json` | JSON | GRN form fields | VS Code |
| `config/templates/stock_issue_form.json` | JSON | SRF form fields | VS Code |
| `.env` | INI | System settings | VS Code |
| `docker-compose.yml` | YAML | Infrastructure | VS Code |
| `packages/data-models/sql/seed.sql` | SQL | Seed data overrides | VS Code |

---

> **Need help?** Open an issue at https://github.com/cpintl/zarishlog/issues
> **For no-code support:** Use the VS Code Configuration Center (Ctrl+Shift+P → Tasks: Run Task → Configure:*)
