# ZarishLog: Modern Humanitarian Logistics & Supply Chain Platform

[![CI Pipeline](https://github.com/cpintl/zarishlog/actions/workflows/ci-test-pipeline.yml/badge.svg)](https://github.com/cpintl/zarishlog/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Stack: Offline-First PWA](https://img.shields.io/badge/Stack-Next.js%20%7C%20Supabase%20%7C%20PouchDB-blue)](https://github.com/cpintl)

## 1.0 Executive & Strategic Blueprint

**ZarishLog** is a purpose-built enterprise-grade technology system designed to address supply chain deficits in unpredictable, volatile, and resource-constrained humanitarian field environments. Operating at the intersection of **Disaster Management**, **Humanitarian Logistics**, and **Sustainable Relief Logistics (SRHL)**, ZarishLog provides real-time transaction transparency and high-accuracy asset accountability.

The system features an engine designed for **offline-first local resilience**, enabling seamless workflow continuation across low-bandwidth environments (e.g., Cox's Bazar Refugee Camps, Bangladesh) via edge data caches, moving downstream transactions back to cloud synchronization layers upon connection re-establishment.

---

## 2.0 Core Tech Stack Architecture

ZarishLog is architected as a high-performance, containerized monorepo structured around fully open-source tech stacks and zero-cost scaling models:

* **Frontend Ecosystem:** PWA-native Next.js core providing adaptive layout responsiveness across desktop, tablet, and field mobile devices.
* **Mobile Framework:** React Native / Expo engine utilizing a unified application runtime for Android and iOS systems.
* **Database Architecture:** PostgreSQL (Supabase Engine) optimized with strict Row-Level Security (RLS) policies for secure multi-tenant isolation, combined with PouchDB/SQLite client interfaces for offline transaction persistence.
* **Infrastructure Management:** Docker Compose automated environment provisioning suitable for self-hosted instances, local machines, or cloud deployment layers.

---

## 3.0 Operational & Functional Hierarchies

### 3.1 Structural Site Mapping Matrix
All dynamic logistics interactions track systematically down through a strict 4-level structural hierarchy:

```text
Level 01: Global Headquarters (CPI HQ - Berkeley, CA)
   └── Level 02: Country Offices (CPI Bangladesh Mission - Gulshan, Dhaka)
          └── Level 03: Project Offices (Cox's Bazar Operations Infrastructure Hub)
                 └── Level 04: Program Site Stores (Anchor Health Posts, WASH Centers)
```

### 3.2 Canonical Terminology Standards

To maintain complete data integrity across fragmented legacy resources, the enterprise data pipeline maps structural concepts to verified system entities:

* `SKU` $\rightarrow$ **`Item Code`** / **`Item Unique Code`**
* `Bin / Shelf / Zone` $\rightarrow$ **`Location ID`** (Canonical spatial bin coordinate)
* `Theme / Group / Focus` $\rightarrow$ **`Intervention Thematic Domain`**
* `Equipment / Fixed Resource` $\rightarrow$ **`Asset`** ($>1$ Year Lifecycle; Serial Tracked)
* `Consumables / Medical Supplies` $\rightarrow$ **`Inventory`** (Short-term; Batch/Lot Tracked)

---

## 4.0 Local Quick-Start Execution Guide

### 4.1 Prerequisites

Ensure the target deployment machine has the following packages installed:

* Docker Desktop / Docker Engine ver 24.0+
* Node.js v24.x or higher LTS release
* Git SCM engine

### 4.2 Local Environment Provisioning

To instantiate the comprehensive framework stack locally (Frontend PWA, App Backend, Database Engine), clone the repository and execute the composition build:

```bash
# Clone the repository framework
git clone [https://github.com/cpintl/ZarishLog.git](https://github.com/cpintl/ZarishLog.git)
cd zarishlog

# Execute configuration workspace installation
npm install

# Initialize local Docker environment infrastructure
docker-compose -f infra/docker/docker-compose.yml up --build -d

```

Once running, the standard local access points will route as follows:

* **ZarishLog PWA Web App Interface:** `http://localhost:3000`
* **Core Central API Platform Gateway:** `http://localhost:8080`
* **Local Supabase Administration Dashboard:** `http://localhost:54321`

---

## 5.0 Database Seeding & Structural Schema

The localized structural persistence tables are declared within `infra/supabase/migrations/001_core_schema.sql`. The engine tracks core transactions through a **Mother-Child (Header-Line)** relational table array:

```sql
-- Conceptual Overview of Transaction Tracking Mechanics
SELECT 
    m.grn_number, 
    m.stock_in_date, 
    c.item_code, 
    c.quantity_in, 
    c.batch_number, 
    c.expiry_date
FROM stock_in_mother m
JOIN stock_in_child c ON m.stock_in_id = c.stock_in_id
WHERE m.warehouse_id = 'L3-CXB-CW';

```

---

## 6.0 Policy & Governance Compliance Standards

The system systematically hardcodes operational procedures defined by the technical advisory team:

1. **FEFO Routing System (First Expiry, First Out):** Enforced at database level for all pharmaceutical categories (`Pharmaceutical Stock`). Items with the nearest explicit expiry date are surfaced to the pick list first.
2. **FIFO Routing System (First In, First Out):** Applied across all `Consumable Stock` and visibility materials to prevent long-term degradation.
3. **Environmental Integrity Constraints:** Prompts automatic alerts if the ambient pharmacy zones exceed **30°C** or if Cold Chain configurations breach the critical **2°C to 8°C** operating boundary.
4. **Partner Risk Governance:** Automated ranking tiering (Silver, Gold, Platinum Capacity Tiers) derived from Pre-Award Due Diligence Assessments to systematically dictate fund distribution mechanics.

---

## 7.0 Open Source Licensing & Authority

Distributed under the terms of the MIT License. See `LICENSE` for structural context.

**System Authority & Technical Maintenance Governance:** Designed, developed, and maintained by the **ZarishLog Team** in partnership with the Community Partners International Bangladesh Mission Health Programs Division. For integration specifications, contact `bgd.cpms@cpintl.org`.

---

## Repository Tree

```text
ZarishLog/
├── .github/
│   ├── ISSUE_TEMPLATE/
│   │   ├── 01_bug_report.md
│   │   ├── 02_feature_request.md
│   │   └── 03_field_incident.md
│   └── workflows/
│       ├── cd-production-deploy.yml
│       ├── ci-test-pipeline.yml
│       └── profile-readme-sync.yml
├── apps/
│   ├── backend/
│   │   ├── src/
│   │   │   ├── controllers/
│   │   │   ├── middleware/
│   │   │   ├── routes/
│   │   │   └── index.js
│   │   ├── Dockerfile
│   │   └── package.json
│   ├── frontend-pwa/
│   │   ├── public/
│   │   │   ├── assets/
│   │   │   └── manifest.json
│   │   ├── src/
│   │   │   ├── components/
│   │   │   ├── hooks/
│   │   │   ├── pages/
│   │   │   └── styles/
│   │   ├── next.config.js
│   │   ├── Dockerfile
│   │   └── package.json
│   └── mobile-app/
│       ├── assets/
│       ├── src/
│       │   ├── database/ (Local SQLite / PouchDB offline syncing)
│       │   └── screens/
│       ├── app.json
│       └── package.json
├── docs/
│   ├── 001-master-catalogue.md
│   ├── 003-field-directory.json
│   ├── 004-warehouse-handbook.md
│   └── 005-due-diligence.md
├── infra/
│   ├── docker/
│   │   └── docker-compose.yml
│   └── supabase/
│       ├── migrations/
│       │   ├── 001_core_schema.sql
│       │   └── 002_rls_policies.sql
│       └── seed.sql
├── packages/
│   └── shared/
│       ├── constants/
│       ├── utils/
│       └── package.json
├── .gitignore
├── LICENSE
├── README.md
└── turbo.json
```
