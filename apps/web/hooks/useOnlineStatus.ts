"use client";

import { useState, useEffect, useCallback, useSyncExternalStore } from "react";
import { isOnline, onOnline, onOffline } from "../lib/offline";
import { getPendingMutationCount, processQueue } from "../lib/sync";

export function useOnlineStatus() {
  const [pendingCount, setPendingCount] = useState(0);
  const [syncing, setSyncing] = useState(false);

  const syncNow = useCallback(async () => {
    if (syncing || !isOnline()) return;
    setSyncing(true);
    try {
      const result = await processQueue("/api/v1");
      setPendingCount(await getPendingMutationCount());
      return result;
    } finally {
      setSyncing(false);
    }
  }, [syncing]);

  const online = useSyncExternalStore(
    (callback) => {
      const unsubOnline = onOnline(() => {
        callback();
        syncNow();
      });
      const unsubOffline = onOffline(callback);
      return () => {
        unsubOnline();
        unsubOffline();
      };
    },
    () => isOnline(),
    () => true,
  );

  useEffect(() => {
    const interval = setInterval(async () => {
      setPendingCount(await getPendingMutationCount());
    }, 5000);

    getPendingMutationCount().then(setPendingCount);

    return () => clearInterval(interval);
  }, []);

  return { online, pendingCount, syncing, syncNow };
}
