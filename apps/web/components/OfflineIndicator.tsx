"use client";

import { useOnlineStatus } from "../hooks/useOnlineStatus";

export default function OfflineIndicator() {
  const { online, pendingCount, syncing, syncNow } = useOnlineStatus();

  if (online && pendingCount === 0) return null;

  return (
    <div
      className={`fixed bottom-0 left-0 right-0 z-50 px-4 py-2 text-sm text-white flex items-center justify-between gap-2 ${
        online ? "bg-amber-600" : "bg-red-600"
      }`}
    >
      <span>
        {online
          ? `${pendingCount} pending change${pendingCount !== 1 ? "s" : ""} — will sync automatically`
          : "You are offline — changes will sync when connected"}
      </span>
      {online && pendingCount > 0 && (
        <button
          onClick={syncNow}
          disabled={syncing}
          className="rounded bg-white/20 px-3 py-1 text-xs font-medium hover:bg-white/30 disabled:opacity-50"
        >
          {syncing ? "Syncing..." : "Sync now"}
        </button>
      )}
    </div>
  );
}
