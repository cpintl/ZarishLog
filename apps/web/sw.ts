import { precacheAndRoute } from "workbox-precaching";
import { registerRoute } from "workbox-routing";
import { NetworkFirst, CacheFirst, StaleWhileRevalidate } from "workbox-strategies";
import { BackgroundSyncPlugin } from "workbox-background-sync";

precacheAndRoute((self as any).__WB_MANIFEST);

const bgSyncPlugin = new BackgroundSyncPlugin("zarishlog-sync", {
  maxRetentionTime: 24 * 60,
});

registerRoute(
  ({ url }) => url.pathname.startsWith("/api/v1/"),
  new NetworkFirst({
    cacheName: "api-cache",
    plugins: [bgSyncPlugin],
  })
);

registerRoute(
  ({ request }) => request.destination === "document",
  new NetworkFirst({ cacheName: "pages" })
);

registerRoute(
  ({ request }) =>
    request.destination === "style" ||
    request.destination === "script" ||
    request.destination === "worker",
  new StaleWhileRevalidate({ cacheName: "static-resources" })
);

registerRoute(
  ({ request }) => request.destination === "image" || request.destination === "font",
  new CacheFirst({ cacheName: "assets" })
);

(self as any).onmessage = (event: MessageEvent) => {
  if (event.data === "SKIP_WAITING") {
    (self as any).skipWaiting();
  }
};
