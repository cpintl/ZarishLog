import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "ZarishLog",
  description: "Humanitarian Logistics, Supply Chain & Asset Management Platform",
  manifest: "/manifest.json",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-slate-50 text-slate-900 antialiased">{children}</body>
    </html>
  );
}
