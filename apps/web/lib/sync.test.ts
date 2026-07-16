import "fake-indexeddb/auto";
import { describe, it, expect, beforeEach, vi } from "vitest";
import db from "./db";
import { queueMutation, getPendingMutationCount } from "./sync";

describe("sync manager", () => {
  beforeEach(async () => {
    await db.mutations.clear();
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2025-01-15T10:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("queues a mutation", async () => {
    const id = await queueMutation("products", "create", "p1", { name: "Test" });
    expect(id).toBeGreaterThan(0);

    const mutations = await db.mutations.toArray();
    expect(mutations).toHaveLength(1);
    expect(mutations[0].table).toBe("products");
    expect(mutations[0].action).toBe("create");
    expect(mutations[0].recordId).toBe("p1");
    expect(mutations[0].retryCount).toBe(0);
  });

  it("reports pending mutation count", async () => {
    expect(await getPendingMutationCount()).toBe(0);

    await queueMutation("products", "update", "p1", { name: "Updated" });
    await queueMutation("products", "delete", "p2", {});

    expect(await getPendingMutationCount()).toBe(2);
  });

  it("queues multiple mutations", async () => {
    await queueMutation("stockLevels", "create", "s1", { qty: 10 });
    await queueMutation("stockLevels", "update", "s1", { qty: 5 });
    await queueMutation("stockLevels", "delete", "s1", {});

    expect(await db.mutations.count()).toBe(3);
  });
});
