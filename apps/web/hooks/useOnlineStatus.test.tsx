import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useOnlineStatus } from "./useOnlineStatus";

const { processQueue, getPendingMutationCount } = vi.hoisted(() => ({
  processQueue: vi.fn(),
  getPendingMutationCount: vi.fn(),
}));

vi.mock("../lib/sync", () => ({
  getPendingMutationCount,
  processQueue,
}));

vi.mock("../lib/offline", () => ({
  isOnline: () => true,
  onOnline: (callback: () => void) => {
    window.addEventListener("online", callback);
    return () => window.removeEventListener("online", callback);
  },
  onOffline: (callback: () => void) => {
    window.addEventListener("offline", callback);
    return () => window.removeEventListener("offline", callback);
  },
}));

describe("useOnlineStatus", () => {
  beforeEach(() => {
    processQueue.mockResolvedValue({ success: 0, failed: 0 });
    getPendingMutationCount.mockResolvedValue(0);
    processQueue.mockClear();
    getPendingMutationCount.mockClear();
  });

  it("syncs queued mutations when the browser comes back online", async () => {
    const { result } = renderHook(() => useOnlineStatus());

    await act(async () => {
      window.dispatchEvent(new Event("online"));
    });

    expect(processQueue).toHaveBeenCalledWith("/api/v1");
    expect(result.current.online).toBe(true);
  });
});
