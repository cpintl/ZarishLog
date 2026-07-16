import "fake-indexeddb/auto";
import { describe, it, expect, vi } from "vitest";
import { isOnline, onOnline, onOffline } from "./offline";

describe("offline utilities", () => {
  it("isOnline returns current navigator.onLine", () => {
    expect(typeof isOnline()).toBe("boolean");
  });

  it("onOnline registers a callback for online event", () => {
    const cb = vi.fn();
    const cleanup = onOnline(cb);

    window.dispatchEvent(new Event("online"));
    expect(cb).toHaveBeenCalledOnce();

    cleanup();
    window.dispatchEvent(new Event("online"));
    expect(cb).toHaveBeenCalledOnce();
  });

  it("onOffline registers a callback for offline event", () => {
    const cb = vi.fn();
    const cleanup = onOffline(cb);

    window.dispatchEvent(new Event("offline"));
    expect(cb).toHaveBeenCalledOnce();

    cleanup();
    window.dispatchEvent(new Event("offline"));
    expect(cb).toHaveBeenCalledOnce();
  });

  it("cleanup removes event listeners", () => {
    const onlineCb = vi.fn();
    const offlineCb = vi.fn();

    const cleanupOnline = onOnline(onlineCb);
    const cleanupOffline = onOffline(offlineCb);

    cleanupOnline();
    cleanupOffline();

    window.dispatchEvent(new Event("online"));
    window.dispatchEvent(new Event("offline"));

    expect(onlineCb).not.toHaveBeenCalled();
    expect(offlineCb).not.toHaveBeenCalled();
  });
});
