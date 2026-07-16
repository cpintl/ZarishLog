import withPWA from "next-pwa";

const pwaConfig = {
  dest: "public",
  register: true,
  skipWaiting: true,
  swSrc: "sw.ts",
  swDest: "public/sw.js",
  disable: process.env.NODE_ENV === "development",
};

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  output: "standalone",
};

export default withPWA(pwaConfig)(nextConfig);
