import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: [],
    include: ["lib/**/*.test.ts", "lib/**/*.test.tsx", "hooks/**/*.test.ts", "hooks/**/*.test.tsx"],
    clearMocks: true,
    fakeTimers: {
      toFake: [],
      doNotFake: ["setTimeout", "setInterval", "requestAnimationFrame"],
    },
    testTimeout: 5000,
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "."),
    },
  },
});
