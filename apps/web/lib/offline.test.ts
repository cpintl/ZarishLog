import { describe, expect, it, vi } from "vitest";
import { isOnline, onOffline, onOnline } from "./offline";

describe("offline connectivity helpers", () => {
  it("reads the browser online state", () => {
    Object.defineProperty(navigator, "onLine", {
      configurable: true,
      value: false,
    });

    expect(isOnline()).toBe(false);
  });

  it("registers and removes online listeners", () => {
    const callback = vi.fn();
    const remove = onOnline(callback);

    window.dispatchEvent(new Event("online"));
    expect(callback).toHaveBeenCalledTimes(1);

    remove();
    window.dispatchEvent(new Event("online"));
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it("registers and removes offline listeners", () => {
    const callback = vi.fn();
    const remove = onOffline(callback);

    window.dispatchEvent(new Event("offline"));
    expect(callback).toHaveBeenCalledTimes(1);

    remove();
    window.dispatchEvent(new Event("offline"));
    expect(callback).toHaveBeenCalledTimes(1);
  });
});
