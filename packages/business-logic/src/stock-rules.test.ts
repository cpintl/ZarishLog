import { describe, it, expect } from "vitest";
import {
  allocateFefoPicking,
  calculateAmc,
  calculateBufferStock,
  calculateReorderPoint,
  classifyStockLevel,
} from "./stock-rules";

describe("allocateFefoPicking", () => {
  it("picks from earliest-expiry batch first", () => {
    const batches = [
      { batchId: "b2", batchNo: "B002", expiryDate: new Date("2027-06-01"), quantityAvailable: 50 },
      { batchId: "b1", batchNo: "B001", expiryDate: new Date("2026-12-01"), quantityAvailable: 30 },
    ];
    const result = allocateFefoPicking(batches, 40);
    expect(result[0].batchId).toBe("b1");
    expect(result[0].quantity).toBe(30);
    expect(result[1].batchId).toBe("b2");
    expect(result[1].quantity).toBe(10);
  });

  it("throws when insufficient stock across all batches", () => {
    const batches = [{ batchId: "b1", batchNo: "B001", expiryDate: null, quantityAvailable: 5 }];
    expect(() => allocateFefoPicking(batches, 10)).toThrow(/Insufficient stock/);
  });
});

describe("calculateAmc", () => {
  it("averages monthly consumption", () => {
    expect(calculateAmc([100, 120, 110])).toBeCloseTo(110);
  });
  it("returns 0 for no data", () => {
    expect(calculateAmc([])).toBe(0);
  });
});

describe("calculateBufferStock / calculateReorderPoint", () => {
  it("computes buffer stock as half of interval consumption", () => {
    expect(calculateBufferStock(100, 3)).toBe(150);
  });
  it("computes reorder point from AMC, lead time, and buffer", () => {
    expect(calculateReorderPoint(100, 1, 150)).toBe(250);
  });
});

describe("classifyStockLevel", () => {
  it("flags sleeping stock after 6+ months of no movement", () => {
    expect(
      classifyStockLevel({ currentQty: 500, reorderPoint: 100, maxStockQty: 1000, monthsSinceLastMovement: 7 })
    ).toBe("SLEEPING_STOCK");
  });
  it("flags low stock below reorder point", () => {
    expect(
      classifyStockLevel({ currentQty: 50, reorderPoint: 100, maxStockQty: 1000, monthsSinceLastMovement: 1 })
    ).toBe("LOW_STOCK");
  });
  it("flags overstock above max", () => {
    expect(
      classifyStockLevel({ currentQty: 1200, reorderPoint: 100, maxStockQty: 1000, monthsSinceLastMovement: 1 })
    ).toBe("OVERSTOCK");
  });
});
