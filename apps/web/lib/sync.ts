import db, { type OfflineMutation } from "./db";

export async function queueMutation(
  table: string,
  action: OfflineMutation["action"],
  recordId: string,
  payload: Record<string, unknown>
): Promise<void> {
  await db.mutations.add({
    table,
    action,
    recordId,
    payload: JSON.stringify(payload),
    createdAt: new Date().toISOString(),
    retryCount: 0,
  });

  if ("serviceWorker" in navigator && "SyncManager" in window) {
    const registration = await navigator.serviceWorker.ready;
    await registration.sync.register("zarishlog-sync");
  }
}

export async function processQueue(
  apiBase: string,
  onProgress?: (processed: number, total: number) => void
): Promise<{ success: number; failed: number }> {
  const mutations = await db.mutations.toArray();
  if (mutations.length === 0) return { success: 0, failed: 0 };

  let success = 0;
  let failed = 0;

  for (let i = 0; i < mutations.length; i++) {
    const m = mutations[i];
    onProgress?.(i + 1, mutations.length);

    try {
      const method = m.action === "delete" ? "DELETE" : m.action === "create" ? "POST" : "PUT";
      const endpoint = m.action === "create" ? `/${m.table}` : `/${m.table}/${m.recordId}`;

      const res = await fetch(`${apiBase}${endpoint}`, {
        method,
        headers: { "Content-Type": "application/json" },
        body: method !== "DELETE" ? m.payload : undefined,
      });

      if (!res.ok) throw new Error(`HTTP ${res.status}`);

      await db.mutations.delete(m.id!);
      success++;
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : String(err);
      await db.mutations.update(m.id!, {
        retryCount: m.retryCount + 1,
        lastError: errMsg,
      });
      failed++;
    }
  }

  return { success, failed };
}

export async function getPendingMutationCount(): Promise<number> {
  return db.mutations.count();
}
