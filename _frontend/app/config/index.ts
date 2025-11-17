// Configuration for backend API and WebSocket connections
// This allows the app to work with different network configurations

// Get the backend URL from environment variable or use default
// For production: set NEXT_PUBLIC_BACKEND_URL to your actual backend URL
// For local development with same device: http://localhost:8080
// For local development with other devices on same WiFi: http://<YOUR_IP>:8080
export const BACKEND_URL =
  process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";

// WebSocket URL is derived from BACKEND_URL
// Replaces http/https with ws/wss protocol
export const WS_URL = BACKEND_URL.replace(/^http/, "ws");

// Export configuration object
export const config = {
  backendUrl: BACKEND_URL,
  wsUrl: WS_URL,
} as const;

export default config;
