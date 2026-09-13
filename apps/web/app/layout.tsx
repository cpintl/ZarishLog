import "./globals.css";
import type { Metadata, Viewport } from "next";
import OfflineIndicator from "../components/OfflineIndicator";

const SITE_URL = "https://cpintl.github.io/ZarishLog";

export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: {
    default: "ZarishLog",
    template: "%s · ZarishLog",
  },
  description:
    "Humanitarian Logistics, Supply Chain & Asset Management Platform — offline-first, open source.",
  applicationName: "ZarishLog",
  keywords: ["humanitarian", "logistics", "supply chain", "warehouse", "PWA", "open source"],
  manifest: "/manifest.json",
  icons: {
    icon: "/logo.svg",
    shortcut: "/logo.svg",
    apple: "/icons/apple-touch-icon.png",
  },
  appleWebApp: {
    capable: true,
    statusBarStyle: "default",
    title: "ZarishLog",
  },
  openGraph: {
    type: "website",
    siteName: "ZarishLog",
    title: "ZarishLog — Humanitarian Logistics, made calm",
    description:
      "Offline-first, multi-tenant humanitarian logistics platform. Go + PostgreSQL + Next.js PWA, with a no-code Config Studio.",
    images: [
      {
        url: "/branding/og-banner.png",
        width: 1200,
        height: 630,
        alt: "ZarishLog — Humanitarian Logistics, Supply Chain & Asset Management",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "ZarishLog — Humanitarian Logistics, made calm",
    description:
      "Offline-first, multi-tenant humanitarian logistics platform with a no-code Config Studio.",
    images: ["/branding/og-banner.png"],
  },
  other: {
    "mobile-web-app-capable": "yes",
  },
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  themeColor: "#0f766e",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <link rel="apple-touch-icon" href="/icons/apple-touch-icon.png" />
      </head>
      <body className="min-h-screen bg-slate-50 text-slate-900 antialiased">
        {children}
        <OfflineIndicator />
      </body>
    </html>
  );
}