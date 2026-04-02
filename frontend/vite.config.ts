import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const frontendHost = process.env.FRONTEND_HOST ?? "0.0.0.0";
const frontendPort = Number(process.env.FRONTEND_PORT ?? "3000");
const backendTarget = process.env.VITE_PROXY_TARGET ?? "http://127.0.0.1:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    host: frontendHost,
    port: frontendPort,
    strictPort: true,
    proxy: {
      "/api": {
        target: backendTarget,
        changeOrigin: true,
      },
      "/health": {
        target: backendTarget,
        changeOrigin: true,
      },
    },
  },
  preview: {
    host: frontendHost,
    port: frontendPort,
    strictPort: true,
  },
});
