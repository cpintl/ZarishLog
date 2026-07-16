export function isOnline(): boolean {
  return navigator.onLine;
}

export function onOnline(cb: () => void): () => void {
  window.addEventListener("online", cb);
  return () => window.removeEventListener("online", cb);
}

export function onOffline(cb: () => void): () => void {
  window.addEventListener("offline", cb);
  return () => window.removeEventListener("offline", cb);
}
