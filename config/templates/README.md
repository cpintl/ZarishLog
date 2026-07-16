# ZarishLog Template Files — User Guide

This folder contains CSV templates that you can fill in to import data into the ZarishLog platform. These templates are designed for non-technical users — no coding required.

---

## Quick Start

1. **Download** the template you need from this folder.
2. **Open** it in Excel, Google Sheets, or any spreadsheet program.
3. **Fill in** the rows with your data. Each column header tells you what to put in.
4. **Save** as a `.csv` file (Comma Separated Values). In Excel, choose *Save As → CSV UTF-8 (Comma delimited)*.
5. **Upload** the file to ZarishLog — your logistics officer or system admin can import it.

> **Important:** Do not change the column header row. The system needs those exact column names to read your data. Do not delete the instruction lines starting with `#` — these are notes for you, the system ignores them.

---

## Template Reference

### 1. Physical Stock Count (`stock_count_template.xlsx.csv`)

**Purpose:** Record how many items you actually found on the shelf during a physical inventory count.

**Who fills it out:** Field staff, warehouse keepers, inventory clerks.

**How to use:**
- Fill in the warehouse, date, and your name at the top.
- For each product listed, write how many you actually counted in the `counted_qty` column.
- The `variance` column will be calculated automatically by the system (expected minus counted).
- Write notes if any products are damaged, expired, or missing.

**Validation rules:**
- `warehouse_code` and `product_sku` must already exist in the system.
- `counted_qty` must be a number (can be 0).

---

### 2. Goods Receipt Note / GRN (`goods_receipt_template.xlsx.csv`)

**Purpose:** Record products received from a supplier against a purchase order (PO).

**Who fills it out:** Warehouse team receiving shipments.

**How to use:**
- Enter the GRN number, date, receiver name, PO number, and supplier at the top.
- One row per product received.
- Fill in the batch number and expiry date for tracked items.

**Validation rules:**
- `sku` must exist in the product master.
- `expiry_date` must be a valid date in YYYY-MM-DD format if provided.
- `condition` should be: Good, Damaged, Expired, or Partial.

---

### 3. Stock Issue Note / SRF (`stock_issue_template.xlsx.csv`)

**Purpose:** Record products issued from a warehouse to a program, department, or beneficiary.

**Who fills it out:** Storekeepers, logistics officers.

**How to use:**
- Fill in the issue number, date, requester, approver, program, and department at the top.
- Write how many were requested and how many were actually issued.
- Leave `quantity_issued` blank if the full request is denied — the warehouse team will fill it in.

**Validation rules:**
- `sku` must exist in the product master.
- `quantity_issued` cannot exceed `quantity_requested`.
- Each issue reduces stock in the issuing warehouse.

---

### 4. Product Import (`product_import_template.xlsx.csv`)

**Purpose:** Add new products to the ZarishLog master product catalog.

**Who fills it out:** Program managers, logistics coordinators, procurement officers.

**How to use:**
- One row per product. The `sku` must be unique.
- For `category_name`, use one of the existing categories (e.g., Medicines & Drugs, Nutrition, Supplies & Sundries, IT Equipment, Wash & Dignity).
- For `item_type`, choose from: `drug`, `medical_supply`, `equipment`, `instrument`, `material`, `vaccine`, `nutrition`, `lab_reagent`, `asset`, `consumable`.
- Boolean fields (`is_batch_tracked`, `is_expiry_tracked`, etc.): write TRUE or FALSE.

**Validation rules:**
- `sku`, `name`, `category_name`, `uom_abbreviation`, and `item_type` are required.
- `category_name` must match an existing category in the system.
- `uom_abbreviation` must match an existing unit of measure (e.g., EA, BX, SA, KG, L).
- `sku` must be unique — no duplicates allowed.

---

### 5. Organization Import (`organization_import_template.xlsx.csv`)

**Purpose:** Define your organization's hierarchy — headquarters, country offices, project offices, and program sites.

**Who fills it out:** Country directors, program managers, HQ operations teams.

**How to use:**
- Level 1 (Global HQ): Leave `parent_code` blank.
- Level 2 (Country Office): `parent_code` = the HQ code.
- Level 3 (Project Office): `parent_code` = the country office code.
- Level 4 (Program Site): `parent_code` = the project office code.

**Validation rules:**
- `code` must be unique.
- `level` must be 1, 2, 3, or 4.
- Parent codes must exist for levels 2, 3, and 4.

---

### 6. Distribution / Delivery (`distribution_template.xlsx.csv`)

**Purpose:** Record products distributed to beneficiaries at a distribution point.

**Who fills it out:** Distribution teams, field monitors, program staff.

**How to use:**
- Fill in the distribution number, date, location, program, and beneficiary count at the top.
- One row per product being distributed.
- Record how much was actually handed out versus how much was planned.

**Validation rules:**
- `sku` must exist in the product master.
- `quantity_distributed` cannot exceed `quantity_planned`.
- Distribution reduces stock in the delivering warehouse.

---

### 7. QA Inspection Checklist (`inspection_checklist_template.xlsx.csv`)

**Purpose:** Record quality assurance inspection results for incoming goods.

**Who fills it out:** Quality assurance officers, pharmacy inspectors, logistics officers.

**How to use:**
- Fill in the inspection date, your name, the product SKU, batch number, and GRN number at the top.
- For each question, write Yes or No in the `answer` column.
- If you answer No to a critical question (marked Yes in the `is_critical` column), the shipment may be rejected.
- Write notes explaining any failures.

**Validation rules:**
- `answer` must be Yes or No (case-insensitive).
- Critical failures require notes explaining the issue.

---

### 8. Asset Registration (`asset_registration_template.xlsx.csv`)

**Purpose:** Register fixed assets like laptops, radios, generators, furniture, and vehicles.

**Who fills it out:** Asset managers, admin officers, field office coordinators.

**How to use:**
- One row per physical asset. Each asset needs a unique `asset_tag`.
- Use the naming convention: `ORG-TYPE-NUMBER` (e.g., CPI-LAP-001).
- Assign a custodian (the person responsible for the asset) using their email.
- Choose a `status`: `in_use`, `in_storage`, `under_maintenance`, `disposed`, or `lost`.

**Validation rules:**
- `asset_tag` must be unique.
- `custodian_email` must belong to a registered user.
- `location_code` must exist in the system.
- `status` must be one of the five options listed above.

---

### 9. Stock Transfer (`stock_transfer_template.xlsx.csv`)

**Purpose:** Record stock moved from one warehouse to another.

**Who fills it out:** Warehouse managers, logistics officers.

**How to use:**
- Fill in the transfer number, date, origin warehouse, destination warehouse, and authorizing person at the top.
- One row per product being transferred.
- The system will reduce stock at the origin and add it to the destination.

**Validation rules:**
- `sku` must exist in the product master.
- `quantity` must be a positive number.
- Stock at the origin warehouse must be sufficient.
- Origin and destination warehouses must be different.

---

### 10. Stock Adjustment (`stock_adjustment_template.xlsx.csv`)

**Purpose:** Correct inventory records after a physical count, or record losses due to damage, theft, expiry, or other reasons.

**Who fills it out:** Warehouse managers, inventory controllers.

**How to use:**
- Fill in the adjustment number, date, and warehouse at the top.
- Enter the expected quantity on record and the actual quantity found.
- Choose a reason code:
  - `DAMAGE` — Items damaged in storage or handling
  - `EXPIRY` — Items passed their expiry date
  - `THEFT` — Items stolen
  - `COUNT_ERROR` — System inventory was wrong
  - `BREAKAGE` — Items broken
  - `DONATION` — Items donated out
  - `SAMPLE` — Items taken as samples
  - `RETURN_TO_SUPPLIER` — Items sent back

**Validation rules:**
- `reason_code` must be one of the codes listed above.
- A negative difference means stock is being removed; positive means stock is being added.
- Provide notes explaining the adjustment, especially for theft, damage, and expiry.

---

### 11. Disposal Form (`disposal_form_template.xlsx.csv`)

**Purpose:** Record the disposal of expired, damaged, or unusable stock.

**Who fills it out:** Warehouse managers, pharmacy officers, disposal committee members.

**How to use:**
- Fill in the disposal number, date, warehouse, disposal method, authorized person, and witness at the top.
- One row per product being disposed.
- Choose a disposal method:
  - `INCINERATION` — Burning (for pharmaceuticals, medical waste)
  - `BURIAL` — Pit burial (for certain medical waste)
  - `CHEMICAL` — Chemical treatment
  - `RECYCLING` — Recycling (for packaging, plastics)
  - `LANDFILL` — Sending to landfill
  - `RETURN` — Returning to supplier/distributor

**Validation rules:**
- `sku` must exist in the product master.
- `disposal_method` must be one of the six options listed above.
- Disposal reduces stock permanently.
- Both the authorized person and witness must sign off.

---

## General Tips

- **Dates** should be in YYYY-MM-DD format (e.g., 2026-07-15 for July 15, 2026).
- **Numbers** should not have commas or currency symbols (write 1200.00 not $1,200.00).
- **Leave cells blank** if the information is not available or not applicable — do not write N/A.
- **Do not rename columns** or insert new columns — the system expects exactly the headers shown.
- **Check for duplicates** before importing products, organizations, or assets.
- **Save a copy** of your filled-in file before uploading, in case you need to re-import or correct errors.

## Need Help?

Contact your ZarishLog system administrator or logistics lead. They can help validate your data before import and troubleshoot any errors.
