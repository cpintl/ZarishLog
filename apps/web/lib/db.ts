import Dexie, { type EntityTable } from "dexie";

export interface OfflineProduct {
  id: string;
  orgId: string;
  name: string;
  sku: string;
  categoryId: string;
  uomId: string;
  itemType: string;
  isBatchTracked: boolean;
  isSerialTracked: boolean;
  isActive: boolean;
  updatedAt: string;
  syncedAt?: string;
}

export interface OfflineStockLevel {
  id: string;
  orgId: string;
  productId: string;
  warehouseId: string;
  batchId: string | null;
  quantity: number;
  updatedAt: string;
  syncedAt?: string;
}

export interface OfflineWarehouse {
  id: string;
  orgId: string;
  name: string;
  code: string;
  type: string;
  isActive: boolean;
  updatedAt: string;
}

export interface OfflineMutation {
  id?: number;
  table: string;
  action: "create" | "update" | "delete";
  recordId: string;
  payload: string;
  createdAt: string;
  retryCount: number;
  lastError?: string;
}

const db = new Dexie("zarishlog") as Dexie & {
  products: EntityTable<OfflineProduct, "id">;
  stockLevels: EntityTable<OfflineStockLevel, "id">;
  warehouses: EntityTable<OfflineWarehouse, "id">;
  mutations: EntityTable<OfflineMutation, "id">;
};

db.version(1).stores({
  products: "id, orgId, sku, name, isActive, updatedAt",
  stockLevels: "id, orgId, productId, warehouseId, updatedAt",
  warehouses: "id, orgId, code, isActive, updatedAt",
  mutations: "++id, table, action, recordId, createdAt",
});

export default db;
