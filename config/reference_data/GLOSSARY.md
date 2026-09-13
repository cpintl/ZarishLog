# ZarishLog — Standardized Terminology Glossary

## Abbreviations

### Organizational

| Abbreviation | Expansion |
|---|---|
| CPI | Community Partners International |
| CPI-BD | Community Partners International Bangladesh Mission |
| CPIL | Community Partners International Limited |
| YPSA | Young Power in Social Action |
| RTMI | Research Training and Management International |
| Friendship | Friendship NGO |
| FIVDB | Friends in Village Development Bangladesh |
| MoH&FW | Ministry of Health and Family Welfare |
| CXB-DSH | Cox's Bazar District Sadar Hospital |
| CXB-CS | Cox's Bazar Civil Surgeon |
| RHU | Refugee Health Unit |
| RRRC | Refugee Relief Repatriation Commissioner |
| WHO | World Health Organization |
| CXB-HS | Cox's Bazar Health Sector |
| UNHCR |  |
| UNFPA |  |
| UNICEF |  |
| UNDP |  |
| IOM |  |
| SRH-WG | Sexual and Reproductive Health Working Group |
| CHW-WG | Community Health Working Group |
| IPC-TWG | Infection Prevention and Control Working Group |







### Technical & Database

| Abbreviation | Expansion |
|---|---|
| AMC | Average Monthly Consumption |
| CRM | Customer/Constituent Relationship Management |
| FEFO | First Expiry First Out |
| GRN | Goods Receipt Note |
| GTIN | Global Trade Item Number |
| MCB | Miniature Circuit Breaker |
| MPPT | Maximum Power Point Tracking (solar) |
| PO | Purchase Order |
| QA | Quality Assurance |
| RCCB | Residual Current Circuit Breaker |
| RLS | Row-Level Security |
| SKU | Stock Keeping Unit |
| UNSPSC | United Nations Standard Products and Services Code |
| UoM | Unit of Measure |
| UUID | Universally Unique Identifier |
| CAPA | Corrective and Preventive Action |
| DGDA | Directorate General of Drug Administration (Bangladesh drug regulator) |
| DNC | Directorate of Narcotics Control (Bangladesh controlled-substance regulator) |
| NBR | National Board of Revenue (Bangladesh customs/tax) |
| WHO PQ | WHO Prequalification of Medicines Programme |
| OOS | Out of Specification |
| OOT | Out of Trend |

### Regulatory & Compliance

| Abbreviation | Expansion |
|---|---|
| CRITICAL | Complaint severity: suspected harm, falsification, contamination, wrong product, controlled-product diversion, or widespread quality risk (immediate escalation required) |
| MAJOR | Complaint severity: material quality/labeling/traceability failure without immediate harm |
| MINOR | Complaint severity: localized or cosmetic issue with no material product-risk |
| MOCK | Annual mock-recall exercise testing traceability of a batch end-to-end |
| ACTUAL | Real recall of distributed product from the market or field |

### Logistics & Supply Chain

| Abbreviation | Expansion |
|---|---|
| EA | Each (unit count) |
| BX | Box |
| CTN | Carton |
| PL | Pallet |
| KG | Kilogram |
| G | Gram |
| L | Liter |
| ML | Milliliter |
| M | Meter |
| M2 | Square Meter |
| DZ | Dozen |
| PR | Pair |
| ST | Set |
| PK | Packet |
| DR | Drum |
| SA | Sachet |

### Health & Nutrition

| Abbreviation | Expansion |
|---|---|
| CNS | Central Nervous System |
| ENT | Ear, Nose, and Throat |
| ITN | Insecticide-Treated Net |
| IV | Intravenous |
| LLIN | Long-Lasting Insecticidal Net |
| LNS | Lipid-based Nutrient Supplement |
| MUAC | Mid-Upper Arm Circumference |
| NaDCC | Sodium Dichloroisocyanurate (water purification) |
| NFI | Non-Food Items |
| ORS | Oral Rehydration Salts |
| PPE | Personal Protective Equipment |
| RUTF | Ready-to-Use Therapeutic Food |
| TPN | Total Parenteral Nutrition |
| WASH | Water, Sanitation, and Hygiene |

---

## Canonical Term Mappings

| Master Catalogue Term | Also Known As / Synonym | Standardized To |
|---|---|---|
| Program | Theme, Pillar, Sector | `PRG-*` code pattern |
| Department | Division, Unit, Cluster | `HLT-*`, `LOG-*`, `FIN-*`, `HR-*` code pattern |
| Function | Role, Process, Workflow Node | `WHS-REC`, `PHA-STK`, `TRN-DSP` code pattern |
| Entity | Object, Resource, Instance | Typed (warehouse, vehicle, user) |
| Product Catalogue | Item Master, Commodity Catalog | `products` table |
| Stock Card | Inventory Ledger, Bin Card | `stock_card_form` |
| Stock In | Goods Receipt, Inbound | `goods_receipt_form` |
| Stock Out | Issue, Dispatch, SRF | `stock_issue_form` |
| Batch | Lot, Serial | `batches` table |
| Warehouse | Store, Depot, Central Storage | `warehouses` table |
| Location | Storage Location, Bin, Slot | Zone → Aisle → Rack → Bin hierarchy |
| AMC | Average Monthly Consumption | 3/6/12-month rolling calculation |
| FEFO | First Expiry, First Out | Picking strategy |
| Reorder Point | Minimum Stock Level, Trigger Point | `reorder_point` field on products |
| Safety Stock | Buffer Stock, Reserve | `safety_stock` field on products |
| Lead Time | Delivery Time, Procurement Lead Time | `lead_time_days` field on products |
| GRN | Goods Received Note, Delivery Note | `goods_receipts` table |
| SRF | Stock Request Form, Issue Note | `stock_issues` table |
| WR | Withdrawal Requisition | Internal program stock request |
| PO | Purchase Order | `purchase_orders` table |
| PR | Purchase Request, Procurement Request | Internal pre-PO request |
| Justification Code | MSF Order Reason | P (Recurring), M (Campaign), E (Emergency), F (Forecast), A (Asset), S (Special) |
| Custodian | Asset Holder, Responsible Person | `users` table FK on assets |
| UoM | Unit of Measure | `units_of_measure` table |
| Regulatory Approval | Licence, Permit, Registration, Certificate of analysis | `regulatory_approvals` table |
| Deviation | Incident, Non-conformance, Quality Event | `deviations` table |
| CAPA | Corrective/Preventive Action, Improvement Action | `capa_actions` table |
| Temperature Excursion | Cold-chain breach, freeze event, OOS temperature | `temperature_excursions` table |
| Controlled Stock | Narcotic/Controlled-Substance Register, CD Register | `controlled_stock_register` table |
| Donation | In-kind Gift, Humanitarian Consignment | `donations` + `donation_line_items` |
| Complaint | Customer Complaint, Quality Complaint, Product Complaint | `complaints` table |
| Recall | Market Withdrawal, Product Retrieval, Mock Recall Exercise | `recalls` + `recall_line_items` |
| Waybill | Dispatch Note, Delivery Note, Bill of Lading | `dispatch_waybills` table |
| Delivery Confirmation | Proof of Delivery, POD, Receiving Confirmation | `delivery_confirmations` table |
| Emergency Plan | Business Continuity Plan, Disaster Response Plan | `emergency_plans` table |
| Change Control | Change Management, Modification Request | `change_controls` table |
| Short-Expiry Review | Near-Expiry Review, Expiry Risk Review | `short_expiry_reviews` table |

## Organization Hierarchy (Standard)

```
Organization (ORG)
  └── Program (PRG-*)           e.g., PRG-HLT, PRG-WASH
       └── Department (HLT-*)   e.g., HLT-PHA, LOG-WHS
            └── Function        e.g., WHS-REC, PHA-STK
                 └── Entity     e.g., Warehouse WH-CXB-CWH
```

## Inventory Statuses (stock_status ENUM)

| Status | Meaning |
|---|---|
| `on_hand` | Physically in stock, available for use |
| `reserved` | Allocated to a specific order/request but not yet picked |
| `committed` | Pledged to a program or project (awaiting dispatch) |
| `in_transit` | En route between warehouses or from supplier |
| `backordered` | Ordered from supplier but not yet received |
| `on_hold` | Temporarily blocked (QA hold, investigation) |
| `damaged` | Identified as damaged, pending disposal decision |
| `expired` | Past expiry date, segregated for disposal |
| `quarantined` | Awaiting QA inspection results |
| `disposed` | Permanently removed from inventory |

## Location Types (location_type ENUM — hierarchy order)

```
Zone (largest area)
  └── Aisle
       └── Rack
            └── Bin (smallest pick face)
```

Also: `shelf` (within room), `area` (functional area like Receiving, Dispatch)

## Supply Chain Workflow

```
Donation/Supplier
     │
     ▼
Stock In (Mother/Child) ──→ QA Inspection ──→ Stock Card ──→ Stock Level
     │                                                    │
     └── Donation Form                                   │
                                                         ▼
                                              Stock Out (Mother/Child)
                                              Withdrawal Requisition
                                              Dispense Form
                                              Emergency Supply
                                                    │
                                                    ▼
                                         Distribution / Beneficiary
                                         Return Form (if applicable)
                                                    │
                                                    ▼
                                         Waste/Disposal (if expired/damaged)
```
