/**
 * Core stock business rules shared by apps/api, apps/web, and apps/mobile.
 * Keeping these as pure functions (no DB access) means the same logic runs
 * identically online (server) and offline (client), which is required for
 * offline-first correctness — see ARCHITECTURE.md §6.
 */

export interface BatchStockPosition {
  batchId: string;
  batchNo: string;
  expiryDate: Date | null;
  quantityAvailable: number;
}

/**
 * FEFO (First Expiry, First Out) picking allocation.
 * Given available batches and a quantity to pick, returns which batches
 * (and how much from each) should be picked, earliest expiry first.
 * Batches with no expiry date are treated as expiring last (FIFO fallback
 * by insertion order is the caller's responsibility if needed).
 */
export function allocateFefoPicking(
  batches: BatchStockPosition[],
  quantityRequested: number
): { batchId: string; batchNo: string; quantity: number }[] {
  const sorted = [...batches].sort((a, b) => {
    if (a.expiryDate === null && b.expiryDate === null) return 0;
    if (a.expiryDate === null) return 1;
    if (b.expiryDate === null) return -1;
    return a.expiryDate.getTime() - b.expiryDate.getTime();
  });

  const allocation: { batchId: string; batchNo: string; quantity: number }[] = [];
  let remaining = quantityRequested;

  for (const batch of sorted) {
    if (remaining <= 0) break;
    if (batch.quantityAvailable <= 0) continue;
    const take = Math.min(batch.quantityAvailable, remaining);
    allocation.push({ batchId: batch.batchId, batchNo: batch.batchNo, quantity: take });
    remaining -= take;
  }

  if (remaining > 0) {
    throw new Error(
      `Insufficient stock: requested ${quantityRequested}, only ${quantityRequested - remaining} available across batches.`
    );
  }

  return allocation;
}

/**
 * Average Monthly Consumption — PRD §6.6 FR-6.1.
 * `monthlyConsumption` should be the most recent N months of issued
 * quantity, oldest first or any order (order doesn't matter for the mean).
 */
export function calculateAmc(monthlyConsumption: number[]): number {
  if (monthlyConsumption.length === 0) return 0;
  const total = monthlyConsumption.reduce((sum, qty) => sum + qty, 0);
  return total / monthlyConsumption.length;
}

/**
 * Buffer/security stock — half of consumption during the delivery interval,
 * per the definition captured from source policy docs (Master Catalogue
 * Definition, §2.2 "Buffer Stock").
 */
export function calculateBufferStock(amc: number, deliveryIntervalMonths: number): number {
  return amc * deliveryIntervalMonths * 0.5;
}

/**
 * Reorder point = (AMC * lead time in months) + buffer stock.
 */
export function calculateReorderPoint(
  amc: number,
  leadTimeMonths: number,
  bufferStock: number
): number {
  return amc * leadTimeMonths + bufferStock;
}

export type StockAlertLevel = "OK" | "LOW_STOCK" | "OVERSTOCK" | "SLEEPING_STOCK";

/**
 * Classifies current stock position for dashboard/alerting purposes.
 * `monthsSinceLastMovement` >= 6 with zero recent consumption => sleeping stock,
 * per the "Sleeping Stock" definition used throughout the source planning docs.
 */
export function classifyStockLevel(params: {
  currentQty: number;
  reorderPoint: number;
  maxStockQty: number | null;
  monthsSinceLastMovement: number;
}): StockAlertLevel {
  const { currentQty, reorderPoint, maxStockQty, monthsSinceLastMovement } = params;
  if (monthsSinceLastMovement >= 6) return "SLEEPING_STOCK";
  if (currentQty <= reorderPoint) return "LOW_STOCK";
  if (maxStockQty !== null && currentQty > maxStockQty) return "OVERSTOCK";
  return "OK";
}
