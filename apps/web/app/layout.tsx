import "./globals.css";
import type { Metadata, Viewport } from "next";
import OfflineIndicator from "../components/OfflineIndicator";

export const metadata: Metadata = {
  title: "ZarishLog",
  description: "Humanitarian Logistics, Supply Chain & Asset Management Platform",
  manifest: "/manifest.json",
  appleWebApp: {
    capable: true,
    statusBarStyle: "default",
    title: "ZarishLog",
  },
  other: {
    "mobile-web-app-capable": "yes",
  },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  themeColor: "#0f172a",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <link rel="apple-touch-icon" href="/icons/icon-192x192.png" />
      </head>
      <body className="min-h-screen bg-slate-50 text-slate-900 antialiased">
        {children}
        <OfflineIndicator />
      </body>
    </html>
  );
}
