export * from "@prisma/client";
export { PrismaClient } from "@prisma/client";

// Shared Zod validation schemas used by both apps/api and apps/web/mobile
// forms, so validation rules are written once. Extend as modules are built.
import { z } from "zod";

export const ProductInputSchema = z.object({
  sku: z.string().min(1),
  name: z.string().min(1),
  categoryId: z.string().min(1),
  uomId: z.string().min(1),
  itemType: z.enum([
    "DRUG",
    "MEDICAL_SUPPLY",
    "EQUIPMENT",
    "INSTRUMENT",
    "MATERIAL",
    "VACCINE",
    "NUTRITION",
    "LAB_REAGENT",
    "ASSET",
    "CONSUMABLE",
  ]),
  batchTracked: z.boolean().default(false),
  serialTracked: z.boolean().default(false),
  expiryTracked: z.boolean().default(false),
  isHazardous: z.boolean().default(false),
  coldChain: z.boolean().default(false),
});
export type ProductInput = z.infer<typeof ProductInputSchema>;

export const StockMovementInputSchema = z.object({
  productId: z.string(),
  warehouseId: z.string(),
  locationId: z.string().optional(),
  batchId: z.string().optional(),
  type: z.enum([
    "RECEIPT",
    "ISSUE",
    "TRANSFER_OUT",
    "TRANSFER_IN",
    "ADJUSTMENT_INCREASE",
    "ADJUSTMENT_DECREASE",
    "RETURN",
    "DISPOSAL",
  ]),
  quantity: z.number().positive(),
  reasonCode: z.string().optional(),
  clientEventId: z.string(), // required — enables idempotent offline sync
  occurredAt: z.string().datetime(),
});
export type StockMovementInput = z.infer<typeof StockMovementInputSchema>;
