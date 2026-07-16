/**
 * ZarishLog seed script
 * -----------------------------------------------------------------------
 * Loads:
 *   1. Reference data: roles/permissions, units of measure, product categories
 *   2. Master org hierarchy: Organization -> OrgLevel (L1-L4) -> Programs -> Departments
 *      (values taken directly from the source planning docs: warehouse list,
 *      role table, and program/department hierarchy)
 *   3. Sample warehouses/locations for Cox's Bazar project office
 *   4. Master product catalogue, imported from config/metadata/master_product_list.csv
 *
 * Run with: pnpm db:seed
 *
 * IMPORTANT: the shipped master_product_list.csv is a small SAMPLE (the
 * source file provided had ~18 rows, not the full 22,000+ item catalogue).
 * The column mapping below (CSV_COLUMNS) is a best-effort inference from
 * that sample and MUST be verified/adjusted against the real, full export
 * before importing production data. Re-run `pnpm db:seed` after fixing
 * the mapping — it is idempotent (upserts by SKU).
 */
import { PrismaClient, OrgType, WarehouseType, LocationType, ItemType } from "@prisma/client";
import { parse } from "csv-parse/sync";
import { readFileSync, existsSync } from "fs";
import { join } from "path";

const prisma = new PrismaClient();

// Best-effort column mapping observed in the sample CSV. Adjust indices
// once the real full catalogue export is available.
const CSV_COLUMNS = {
  sku: 0,
  name: 1,
  itemTypeCode: 5, // e.g. "D" = Drugs
  itemTypeName: 6, // e.g. "Drugs"
  categoryName: 11, // e.g. "Pharmaceutical Stock"
  batchTracked: 13, // "Yes"/"No"
  serialTracked: 14, // "Yes"/"No"
  status: 19, // "Active"/"Inactive"
};

const ITEM_TYPE_MAP: Record<string, ItemType> = {
  D: "DRUG",
  DRUGS: "DRUG",
  S: "MEDICAL_SUPPLY",
  SUPPLIES: "MEDICAL_SUPPLY",
  E: "EQUIPMENT",
  EQUIPMENT: "EQUIPMENT",
  A: "ASSET",
  ASSETS: "ASSET",
};

async function seedRolesAndPermissions() {
  const roles: { id: any; name: string; description: string }[] = [
    { id: "R01_GLOBAL_ADMIN", name: "Global Admin", description: "Full access, all tenants" },
    { id: "R02_COUNTRY_REP", name: "Country Representative", description: "View-all, reporting only (L2)" },
    { id: "R03_THEME_MANAGER", name: "Theme Manager", description: "View-all within one program theme (L2)" },
    { id: "R04_WAREHOUSE_OFFICER", name: "Warehouse Officer", description: "Full central warehouse ops (L3)" },
    { id: "R05_WAREHOUSE_STOREKEEPER", name: "Warehouse Storekeeper", description: "Stock ops only (L3)" },
    { id: "R06_ADMIN_LOG_OFFICER", name: "Admin/Logistics Officer", description: "Office asset/logistics stock (L3)" },
    { id: "R07_DEPT_MANAGER", name: "Department Manager", description: "Budget/flow approval authority" },
    { id: "R08_DEPT_COORDINATOR", name: "Department Coordinator", description: "Validates stock flow (L4)" },
    { id: "R09_DEPT_OFFICER", name: "Department Officer", description: "Day-to-day sub-warehouse ops (L4)" },
  ];

  for (const role of roles) {
    await prisma.role.upsert({ where: { id: role.id }, update: {}, create: role });
  }

  const modules = ["products", "warehouses", "stock", "grn", "qa", "assets", "reports", "users"];
  const actions = ["create", "read", "update", "delete", "approve"];
  for (const module of modules) {
    for (const action of actions) {
      await prisma.permission.upsert({
        where: { module_action: { module, action } },
        update: {},
        create: { module, action },
      });
    }
  }

  // Grant GLOBAL_ADMIN every permission
  const allPermissions = await prisma.permission.findMany();
  for (const p of allPermissions) {
    await prisma.rolePermission.upsert({
      where: { roleId_permissionId: { roleId: "R01_GLOBAL_ADMIN", permissionId: p.id } },
      update: {},
      create: { roleId: "R01_GLOBAL_ADMIN", permissionId: p.id },
    });
  }

  console.log(`Seeded ${roles.length} roles and ${modules.length * actions.length} permissions.`);
}

async function seedReferenceData() {
  const uoms = [
    { code: "PCS", name: "Piece", category: "count" },
    { code: "BOX", name: "Box", category: "count" },
    { code: "VIAL", name: "Vial", category: "count" },
    { code: "BOTTLE", name: "Bottle", category: "count" },
    { code: "PACK", name: "Pack", category: "count" },
    { code: "CASE", name: "Case", category: "count" },
    { code: "PALLET", name: "Pallet", category: "count" },
    { code: "KG", name: "Kilogram", category: "weight" },
    { code: "L", name: "Litre", category: "volume" },
  ];
  for (const uom of uoms) {
    await prisma.unitOfMeasure.upsert({ where: { code: uom.code }, update: {}, create: uom });
  }

  const categories = [
    { code: "PHARMA", name: "Pharmaceutical Stock", isPharma: true },
    { code: "MED_SUPPLY", name: "Medical Supplies", isPharma: false },
    { code: "NON_MED_SUPPLY", name: "Non-Medical Supplies", isPharma: false },
    { code: "EQUIPMENT", name: "Equipment", isPharma: false },
    { code: "IT_ASSET", name: "IT Equipment / Asset", isPharma: false },
    { code: "COLD_CHAIN", name: "Cold Chain Items", isPharma: true },
  ];
  for (const c of categories) {
    await prisma.productCategory.upsert({ where: { code: c.code }, update: {}, create: c });
  }

  console.log(`Seeded ${uoms.length} UoMs and ${categories.length} product categories.`);
}

async function seedOrganizationHierarchy() {
  const org = await prisma.organization.upsert({
    where: { code: "ORG_001" },
    update: {},
    create: {
      // Fixed, human-readable id (instead of a random cuid) so local dev
      // tooling (e.g. apps/web's demo header) can reference it directly
      // before the auth module supplies a real tenant context.
      id: "org_001_seed",
      code: "ORG_001",
      name: "Community Partners International",
      acronym: "CPI",
      type: OrgType.NGO,
    },
  });

  const l1 = await prisma.orgLevel.upsert({
    where: { organizationId_code: { organizationId: org.id, code: "L1-ZS-HQ" } },
    update: {},
    create: {
      organizationId: org.id,
      code: "L1-ZS-HQ",
      name: "ZarishLog Global",
      level: "L1_GLOBAL",
      description: "The global unit of all warehouses, sub-stores, and stock",
    },
  });

  const l2 = await prisma.orgLevel.upsert({
    where: { organizationId_code: { organizationId: org.id, code: "L2-BD-CO" } },
    update: {},
    create: {
      organizationId: org.id,
      code: "L2-BD-CO",
      name: "Bangladesh Country Office",
      level: "L2_COUNTRY",
      parentId: l1.id,
      country: "Bangladesh",
    },
  });

  const l3 = await prisma.orgLevel.upsert({
    where: { organizationId_code: { organizationId: org.id, code: "L3-CXB-PO" } },
    update: {},
    create: {
      organizationId: org.id,
      code: "L3-CXB-PO",
      name: "Cox's Bazar Project Office",
      level: "L3_PROJECT_OFFICE",
      parentId: l2.id,
      region: "Cox's Bazar",
    },
  });

  // Programs (from ZarishLogMasterCatalogue.md hierarchy)
  const programDefs = [
    { code: "PRG-HLT", name: "Health" },
    { code: "PRG-WASH", name: "Water, Sanitation, and Hygiene" },
    { code: "PRG-NUT", name: "Nutrition" },
    { code: "PRG-PRO", name: "Protection" },
    { code: "PRG-SHL", name: "Shelter and Settlements" },
    { code: "PRG-LOG", name: "Logistics and Supply Chain" },
    { code: "PRG-SUP", name: "Program Support" },
  ];
  for (const p of programDefs) {
    await prisma.program.upsert({
      where: { organizationId_code: { organizationId: org.id, code: p.code } },
      update: {},
      create: { organizationId: org.id, code: p.code, name: p.name },
    });
  }

  // Warehouses (from the warehouse master list in project source docs)
  const cwh = await prisma.warehouse.upsert({
    where: { organizationId_code: { organizationId: org.id, code: "L3-CXB-CW" } },
    update: {},
    create: {
      organizationId: org.id,
      orgLevelId: l3.id,
      code: "L3-CXB-CW",
      name: "Central Warehouse — Cox's Bazar",
      type: WarehouseType.CENTRAL,
    },
  });

  const locationDefs: { code: string; name: string; type: LocationType }[] = [
    { code: "RECV", name: "Receiving Area", type: LocationType.RECEIVING },
    { code: "GEN-A", name: "General Storage A", type: LocationType.GENERAL_STORAGE },
    { code: "PROG-HLT", name: "Health Program Storage", type: LocationType.PROGRAM_STORAGE },
    { code: "QA", name: "QA / Quarantine Area", type: LocationType.QUARANTINE },
    { code: "DISPATCH", name: "Dispatch Area", type: LocationType.DISPATCH },
  ];
  for (const loc of locationDefs) {
    await prisma.location.upsert({
      where: { warehouseId_code: { warehouseId: cwh.id, code: loc.code } },
      update: {},
      create: { warehouseId: cwh.id, ...loc },
    });
  }

  console.log("Seeded organization hierarchy: 1 org, 3 org levels, 7 programs, 1 warehouse, 5 locations.");
  return org;
}

async function seedProductsFromCsv(organizationId: string) {
  const csvPath = join(__dirname, "..", "..", "..", "config", "metadata", "master_product_list.csv");
  if (!existsSync(csvPath)) {
    console.warn(`No product CSV found at ${csvPath} — skipping product import.`);
    return;
  }

  const raw = readFileSync(csvPath, "utf-8");
  const rows: string[][] = parse(raw, { skip_empty_lines: true, relax_column_count: true });

  const defaultUom = await prisma.unitOfMeasure.findUniqueOrThrow({ where: { code: "PCS" } });
  const fallbackCategory = await prisma.productCategory.findUniqueOrThrow({ where: { code: "MED_SUPPLY" } });

  let imported = 0;
  let skipped = 0;

  for (const row of rows) {
    const sku = row[CSV_COLUMNS.sku]?.trim();
    const name = row[CSV_COLUMNS.name]?.trim();
    if (!sku || !name) {
      skipped++;
      continue;
    }

    const typeCode = row[CSV_COLUMNS.itemTypeCode]?.trim().toUpperCase();
    const typeName = row[CSV_COLUMNS.itemTypeName]?.trim().toUpperCase();
    const itemType = ITEM_TYPE_MAP[typeCode] ?? ITEM_TYPE_MAP[typeName] ?? "CONSUMABLE";

    const categoryName = row[CSV_COLUMNS.categoryName]?.trim();
    const category = categoryName
      ? await prisma.productCategory.upsert({
          where: { code: categoryName.toUpperCase().replace(/\s+/g, "_").slice(0, 30) },
          update: {},
          create: {
            code: categoryName.toUpperCase().replace(/\s+/g, "_").slice(0, 30),
            name: categoryName,
            isPharma: itemType === "DRUG",
          },
        })
      : fallbackCategory;

    const batchTracked = row[CSV_COLUMNS.batchTracked]?.trim().toLowerCase() === "yes";
    const serialTracked = row[CSV_COLUMNS.serialTracked]?.trim().toLowerCase() === "yes";
    const statusRaw = row[CSV_COLUMNS.status]?.trim().toLowerCase();
    const status = statusRaw === "inactive" ? "INACTIVE" : "ACTIVE";

    await prisma.product.upsert({
      where: { organizationId_sku: { organizationId, sku } },
      update: { name },
      create: {
        organizationId,
        sku,
        name,
        itemType,
        categoryId: category.id,
        uomId: defaultUom.id,
        batchTracked,
        serialTracked,
        expiryTracked: batchTracked, // conservative default: batch-tracked items assumed expiry-relevant
        status: status as "ACTIVE" | "INACTIVE",
        source: "master_product_list.csv",
      },
    });
    imported++;
  }

  console.log(`Imported ${imported} products, skipped ${skipped} malformed rows.`);
  if (imported > 0) {
    console.log(
      "NOTE: verify CSV_COLUMNS mapping in seed.ts against the full 22,000+ item catalogue before relying on this import in production."
    );
  }
}

async function main() {
  console.log("Seeding ZarishLog database...\n");
  await seedRolesAndPermissions();
  await seedReferenceData();
  const org = await seedOrganizationHierarchy();
  await seedProductsFromCsv(org.id);
  console.log("\nSeed complete.");
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
