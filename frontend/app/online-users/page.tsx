"use client";
import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useWebSocket } from "../contexts/WebSocketContext";
import {
  Box,
  Typography,
  Card,
  CardContent,
  Avatar,
  Stack,
  Chip,
  CircularProgress,
} from "@mui/material";
import FiberManualRecordIcon from "@mui/icons-material/FiberManualRecord";
import PersonIcon from "@mui/icons-material/Person";

interface OnlineUser {
  user_id: string;
  username: string;
  email?: string;
  status?: string;
}

export default function OnlineUsersPage() {
  const router = useRouter();
  const [users, setUsers] = useState<OnlineUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const { lastMessage, isConnected } = useWebSocket();

  useEffect(() => {
    const token =
      localStorage.getItem("token") || sessionStorage.getItem("token");
    if (!token) {
      router.replace("/login");
      return;
    }

    fetchOnlineUsers(token);
  }, [router]);

  // Handle real-time updates via WebSocket
  useEffect(() => {
    if (!lastMessage) return;

    console.log("[Online Users] WebSocket message received:", lastMessage);
    console.log("[Online Users] Message type:", lastMessage.type);
    console.log("[Online Users] Message data:", lastMessage.data);

    // Backend now sends: { success: true, type: "user_joined"/"user_left", data: {...} }

    // Handle user joined event - add to list
    if (lastMessage.type === "user_joined") {
      const userData = lastMessage.data;
      console.log("[Online Users] User joined:", userData);

      setUsers((prev) => {
        // Check if user already exists
        const exists = prev.find((u) => u.user_id === userData.user_id);
        if (exists) {
          console.log("[Online Users] User already in list, skipping");
          return prev;
        }

        // Add new user
        const newUser: OnlineUser = {
          user_id: userData.user_id,
          username: userData.username,
          email: userData.email || "",
          status: "online",
        };
        console.log("[Online Users] Adding new user to list:", newUser);
        return [...prev, newUser];
      });
    }

    // Handle user left event - remove from list
    if (lastMessage.type === "user_left") {
      const { user_id } = lastMessage.data || {};
      console.log("[Online Users] User left:", user_id);

      setUsers((prev) => {
        const filtered = prev.filter((u) => u.user_id !== user_id);
        console.log("[Online Users] Users after removal:", filtered.length);
        return filtered;
      });
    }
  }, [lastMessage]);

  const fetchOnlineUsers = async (token: string) => {
    try {
      setLoading(true);
      setError("");

      const response = await fetch("/api/chats/users", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        if (response.status === 401) {
          router.replace("/login");
          return;
        }
        throw new Error(`Failed to fetch online users: ${response.status}`);
      }

      const data = await response.json();
      console.log("[Online Users] Fetched data:", data);

      if (data.success && data.data) {
        // Backend returns { success: true, data: { online_users: [{user_id, username, email}] } }
        const onlineUsers = data.data.online_users || [];
        console.log("[Online Users] Online users:", onlineUsers);

        // Map to our OnlineUser interface
        const usersList: OnlineUser[] = onlineUsers.map((user: any) => ({
          user_id: user.user_id,
          username: user.username,
          email: user.email,
          status: "online",
        }));

        console.log("[Online Users] Final usersList:", usersList);
        setUsers(usersList);
      } else {
        console.log("[Online Users] No valid data, setting empty array");
        setUsers([]);
      }
    } catch (err: any) {
      console.error("[Online Users] Error fetching users:", err);
      setError(err.message || "Failed to load online users");
    } finally {
      setLoading(false);
    }
  };

  const getInitials = (username: string) => {
    return username
      .split(" ")
      .map((word) => word[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  if (loading) {
    return (
      <Box
        sx={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: "100vh",
          bgcolor: "#f5f5f5",
        }}
      >
        <CircularProgress sx={{ color: "#8aa79b" }} />
      </Box>
    );
  }

  return (
    <Box
      sx={{
        minHeight: "100vh",
        bgcolor: "#f5f5f5",
        p: 3,
      }}
    >
      <Box
        sx={{
          maxWidth: 1200,
          mx: "auto",
        }}
      >
        {/* Header */}
        <Box
          sx={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            mb: 3,
          }}
        >
          <Box>
            <Typography
              variant="h4"
              sx={{
                fontWeight: 600,
                color: "#2d3748",
                mb: 0.5,
              }}
            >
              Online Users
            </Typography>
            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
              <Chip
                icon={<FiberManualRecordIcon sx={{ fontSize: 12 }} />}
                label={`${users.length} online`}
                size="small"
                sx={{
                  bgcolor: isConnected ? "#e8f5e9" : "#ffebee",
                  color: isConnected ? "#2e7d32" : "#c62828",
                  "& .MuiChip-icon": {
                    color: isConnected ? "#4caf50" : "#f44336",
                  },
                }}
              />
              <Typography variant="caption" color="text.secondary">
                {isConnected ? "Real-time updates active" : "Reconnecting..."}
              </Typography>
            </Box>
          </Box>
        </Box>

        {/* Error Message */}
        {error && (
          <Card sx={{ mb: 2, bgcolor: "#ffebee" }}>
            <CardContent>
              <Typography color="error">{error}</Typography>
            </CardContent>
          </Card>
        )}

        {/* Users Grid */}
        {users.length === 0 ? (
          <Card>
            <CardContent
              sx={{
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                py: 6,
              }}
            >
              <PersonIcon sx={{ fontSize: 64, color: "#bdbdbd", mb: 2 }} />
              <Typography variant="h6" color="text.secondary">
                No users online
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Check back later or refresh to see who's online
              </Typography>
            </CardContent>
          </Card>
        ) : (
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: {
                xs: "1fr",
                sm: "repeat(2, 1fr)",
                md: "repeat(3, 1fr)",
                lg: "repeat(4, 1fr)",
              },
              gap: 2,
            }}
          >
            {users.map((user) => (
              <Card
                key={user.user_id}
                sx={{
                  cursor: "pointer",
                  transition: "transform 0.2s, box-shadow 0.2s",
                  "&:hover": {
                    transform: "translateY(-4px)",
                    boxShadow: 3,
                  },
                }}
                onClick={() => {
                  // Navigate to user profile
                  router.push(`/profile/${user.user_id}`);
                }}
              >
                <CardContent>
                  <Stack spacing={2} alignItems="center">
                    <Box sx={{ position: "relative" }}>
                      <Avatar
                        sx={{
                          width: 64,
                          height: 64,
                          bgcolor: "#8aa79b",
                          fontSize: "1.5rem",
                        }}
                      >
                        {getInitials(user.username)}
                      </Avatar>
                      <Box
                        sx={{
                          position: "absolute",
                          bottom: 2,
                          right: 2,
                          width: 14,
                          height: 14,
                          bgcolor: "#4caf50",
                          borderRadius: "50%",
                          border: "2px solid white",
                        }}
                      />
                    </Box>
                    <Box sx={{ textAlign: "center", width: "100%" }}>
                      <Typography
                        variant="subtitle1"
                        sx={{
                          fontWeight: 600,
                          color: "#2d3748",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {user.username}
                      </Typography>
                      {user.email && (
                        <Typography
                          variant="caption"
                          color="text.secondary"
                          sx={{
                            display: "block",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap",
                          }}
                        >
                          {user.email}
                        </Typography>
                      )}
                      <Chip
                        icon={<FiberManualRecordIcon sx={{ fontSize: 10 }} />}
                        label="Online"
                        size="small"
                        sx={{
                          mt: 1,
                          height: 20,
                          fontSize: "0.7rem",
                          bgcolor: "#e8f5e9",
                          color: "#2e7d32",
                          "& .MuiChip-icon": {
                            color: "#4caf50",
                          },
                        }}
                      />
                    </Box>
                  </Stack>
                </CardContent>
              </Card>
            ))}
          </Box>
        )}
      </Box>
    </Box>
  );
}
