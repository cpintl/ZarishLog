import "fake-indexeddb/auto";
import { describe, it, expect, beforeEach } from "vitest";
import db from "./db";

describe("Dexie DB", () => {
  beforeEach(async () => {
    await db.products.clear();
    await db.stockLevels.clear();
    await db.warehouses.clear();
    await db.mutations.clear();
  });

  it("has correct tables count", () => {
    expect(db.tables.length).toBe(4);
  });

  it("stores and retrieves products", async () => {
    await db.products.add({
      id: "p1",
      orgId: "org1",
      name: "Paracetamol",
      sku: "MED-PAR-500",
      categoryId: "cat1",
      uomId: "uom1",
      itemType: "drug",
      isBatchTracked: true,
      isSerialTracked: false,
      isActive: true,
      updatedAt: new Date().toISOString(),
    });

    const count = await db.products.count();
    expect(count).toBe(1);

    const product = await db.products.get("p1");
    expect(product?.name).toBe("Paracetamol");
    expect(product?.sku).toBe("MED-PAR-500");
  });

  it("stores and retrieves mutations", async () => {
    await db.mutations.add({
      table: "products",
      action: "create",
      recordId: "p1",
      payload: JSON.stringify({ name: "Test" }),
      createdAt: new Date().toISOString(),
      retryCount: 0,
    });

    const count = await db.mutations.count();
    expect(count).toBe(1);

    const mutations = await db.mutations.toArray();
    expect(mutations[0].table).toBe("products");
    expect(mutations[0].action).toBe("create");
  });

  it("deletes mutations after processing", async () => {
    await db.mutations.add({
      table: "products",
      action: "delete",
      recordId: "p1",
      payload: "{}",
      createdAt: new Date().toISOString(),
      retryCount: 0,
    });

    expect(await db.mutations.count()).toBe(1);
    await db.mutations.clear();
    expect(await db.mutations.count()).toBe(0);
  });

  it("queries products by sku via indexes", async () => {
    await db.products.add({
      id: "p1", orgId: "org1", name: "A", sku: "SKU-001",
      categoryId: "c1", uomId: "u1", itemType: "drug",
      isBatchTracked: false, isSerialTracked: false, isActive: true,
      updatedAt: new Date().toISOString(),
    });
    await db.products.add({
      id: "p2", orgId: "org1", name: "B", sku: "SKU-002",
      categoryId: "c2", uomId: "u2", itemType: "supply",
      isBatchTracked: true, isSerialTracked: false, isActive: true,
      updatedAt: new Date().toISOString(),
    });

    const bySku = await db.products.where("sku").equals("SKU-001").toArray();
    expect(bySku).toHaveLength(1);

    const byOrg = await db.products.where("orgId").equals("org1").toArray();
    expect(byOrg).toHaveLength(2);
  });
});
