# ZarishLog — Standardized Terminology Glossary

## Abbreviations

### Organizational

| Abbreviation | Expansion |
|---|---|
| CPI | Center for Peace and Integrity |
| CPI-BD | CPI Bangladesh Country Office |
| CPI-CXB | CPI Cox's Bazar Project Office |
| CPI-DHK | CPI Dhaka Project Office |
| CPI-MM | CPI Myanmar Country Office |
| CPI-YGN | CPI Yangon Project Office |

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
| Stock In | Goods Receipt, Inbound | `stock_in_form_mother/child` |
| Stock Out | Issue, Dispatch, SRF | `stock_out_form_mother/child` |
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
