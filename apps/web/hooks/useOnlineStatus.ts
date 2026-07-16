"use client";

import { useState, useEffect, useCallback } from "react";
import { isOnline, onOnline, onOffline } from "../lib/offline";
import { getPendingMutationCount, processQueue } from "../lib/sync";

export function useOnlineStatus() {
  const [online, setOnline] = useState(true);
  const [pendingCount, setPendingCount] = useState(0);
  const [syncing, setSyncing] = useState(false);

  useEffect(() => {
    setOnline(isOnline());

    const unsubOnline = onOnline(() => {
      setOnline(true);
      syncNow();
    });
    const unsubOffline = onOffline(() => setOnline(false));

    const interval = setInterval(async () => {
      setPendingCount(await getPendingMutationCount());
    }, 5000);

    getPendingMutationCount().then(setPendingCount);

    return () => {
      unsubOnline();
      unsubOffline();
      clearInterval(interval);
    };
  }, []);

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

  return { online, pendingCount, syncing, syncNow };
}
