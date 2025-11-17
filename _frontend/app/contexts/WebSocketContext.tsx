"use client";

import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useRef,
  useCallback,
} from "react";
import { WS_URL } from "../config";

interface WebSocketContextType {
  ws: WebSocket | null;
  isConnected: boolean;
  sendMessage: (message: any) => void;
  lastMessage: any;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(
  undefined
);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return context;
};

interface WebSocketProviderProps {
  children: React.ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({
  children,
}) => {
  const [ws, setWs] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<any>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttempts = useRef(0);
  const maxReconnectAttempts = 5;
  const reconnectDelay = 3000; // 3 seconds
  const wsRef = useRef<WebSocket | null>(null);

  const connect = useCallback(() => {
    const token =
      localStorage.getItem("token") || sessionStorage.getItem("token");
    if (!token) {
      console.log("[WebSocket] No token found, skipping connection");
      return;
    }

    // Don't reconnect if already connected
    const currentWs = wsRef.current;
    if (currentWs && currentWs.readyState === WebSocket.OPEN) {
      console.log("[WebSocket] Already connected, skipping");
      return;
    }

    // Close existing connection if it exists
    if (currentWs) {
      currentWs.close();
      wsRef.current = null;
    }

    console.log("[WebSocket] Connecting to WebSocket...");
    const wsUrl = `${WS_URL}/chats/ws/?token=${encodeURIComponent(token)}`;
    console.log("[WebSocket] Using URL:", wsUrl);
    const websocket = new WebSocket(wsUrl);
    wsRef.current = websocket;

    websocket.onopen = () => {
      console.log("[WebSocket] Connected successfully");
      setIsConnected(true);
      reconnectAttempts.current = 0; // Reset reconnect attempts on successful connection
    };

    websocket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log("[WebSocket] Message received:", data);
        setLastMessage(data);
      } catch (err) {
        console.error("[WebSocket] Error parsing message:", err);
      }
    };

    websocket.onerror = (error) => {
      console.error("[WebSocket] Error:", error);
      setIsConnected(false);
    };

    websocket.onclose = (event) => {
      console.log("[WebSocket] Disconnected", event.code, event.reason);
      setIsConnected(false);
      wsRef.current = null;
      setWs(null);

      // Attempt to reconnect if it wasn't a manual close
      const token =
        localStorage.getItem("token") || sessionStorage.getItem("token");
      if (token && reconnectAttempts.current < maxReconnectAttempts) {
        reconnectAttempts.current += 1;
        console.log(
          `[WebSocket] Attempting to reconnect (${reconnectAttempts.current}/${maxReconnectAttempts})...`
        );

        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, reconnectDelay);
      } else if (reconnectAttempts.current >= maxReconnectAttempts) {
        console.error("[WebSocket] Max reconnect attempts reached");
      }
    };

    setWs(websocket);
  }, []);

  const sendMessage = useCallback((message: any) => {
    const currentWs = wsRef.current;
    if (currentWs && currentWs.readyState === WebSocket.OPEN) {
      currentWs.send(JSON.stringify(message));
    } else {
      console.warn(
        "[WebSocket] Cannot send message, WebSocket is not connected"
      );
    }
  }, []);

  // Connect when token is available
  useEffect(() => {
    const checkAndConnect = () => {
      const token =
        localStorage.getItem("token") || sessionStorage.getItem("token");
      const currentWs = wsRef.current;
      if (
        token &&
        (!currentWs ||
          currentWs.readyState === WebSocket.CLOSED ||
          currentWs.readyState === WebSocket.CLOSING)
      ) {
        connect();
      }
    };

    // Check immediately
    checkAndConnect();

    // Listen for storage changes (e.g., when user logs in from another tab)
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === "token") {
        if (e.newValue) {
          // Token was added, connect
          connect();
        } else {
          // Token was removed, disconnect
          const currentWs = wsRef.current;
          if (currentWs) {
            currentWs.close();
            wsRef.current = null;
            setWs(null);
            setIsConnected(false);
          }
        }
      }
    };

    // Listen for custom token-set event (for same-window login)
    const handleTokenSet = () => {
      console.log("[WebSocket] Token set event received, connecting...");
      connect();
    };

    window.addEventListener("storage", handleStorageChange);
    window.addEventListener("token-set", handleTokenSet);

    // Also check periodically (in case storage event doesn't fire in same window)
    const interval = setInterval(checkAndConnect, 1000);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
      window.removeEventListener("token-set", handleTokenSet);
      clearInterval(interval);
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      const currentWs = wsRef.current;
      if (currentWs && currentWs.readyState === WebSocket.OPEN) {
        currentWs.close();
      }
    };
  }, [connect]);

  return (
    <WebSocketContext.Provider
      value={{ ws, isConnected, sendMessage, lastMessage }}
    >
      {children}
    </WebSocketContext.Provider>
  );
};
