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
  Button,
} from "@mui/material";
import FiberManualRecordIcon from "@mui/icons-material/FiberManualRecord";
import PersonIcon from "@mui/icons-material/Person";
import ChatIcon from "@mui/icons-material/Chat";

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
  const [chattingWith, setChattingWith] = useState<string | null>(null);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const { lastMessage, isConnected } = useWebSocket();

  useEffect(() => {
    const token =
      localStorage.getItem("token") || sessionStorage.getItem("token");
    if (!token) {
      router.replace("/login");
      return;
    }

    fetchCurrentUserAndOnlineUsers(token);
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

  const fetchCurrentUserAndOnlineUsers = async (token: string) => {
    try {
      setLoading(true);
      setError("");

      // Fetch current user info
      const userResponse = await fetch("/api/users/me", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (userResponse.ok) {
        const userData = await userResponse.json();
        if (userData.success && userData.data) {
          setCurrentUserId(userData.data.user_id?.toString() || "");
          console.log("[Online Users] Current user ID:", userData.data.user_id);
        }
      }

      // Fetch online users
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

  const handleChatWithUser = async (userId: string, username: string) => {
    const token =
      localStorage.getItem("token") || sessionStorage.getItem("token");
    if (!token) {
      router.replace("/login");
      return;
    }

    try {
      setChattingWith(userId);
      setError(""); // Clear any previous errors

      console.log(
        `[Online Users] Starting chat with user: ${username} (${userId})`
      );

      // First, check if a chat already exists with this user
      const chatsResponse = await fetch("/api/chats", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!chatsResponse.ok) {
        throw new Error("Failed to fetch chats");
      }

      const chatsData = await chatsResponse.json();
      console.log(`[Online Users] Fetched chats:`, chatsData);

      if (chatsData.success && chatsData.data) {
        const chats = Array.isArray(chatsData.data)
          ? chatsData.data
          : [chatsData.data];

        // Look for an existing direct chat with this user
        const existingChat = chats.find(
          (chat: any) =>
            !chat.is_group && chat.other_participants_id?.includes(userId)
        );

        if (existingChat) {
          // Chat exists, navigate directly to chat page with chat_id
          console.log(`[Online Users] Found existing chat:`, existingChat);
          router.push(`/chat?chatId=${existingChat.chat_id}`);
          return;
        }
      }

      // No existing chat found, create a new one
      console.log(
        `[Online Users] No existing chat found. Creating new chat with ${username}...`
      );
      const createResponse = await fetch("/api/chats", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          recipient_id: userId,
        }),
      });

      console.log(
        `[Online Users] Create chat response status:`,
        createResponse.status
      );

      const createData = await createResponse.json();
      console.log(
        `[Online Users] Create chat response data:`,
        JSON.stringify(createData, null, 2)
      );

      if (!createResponse.ok) {
        console.error(`[Online Users] Create chat failed:`, {
          status: createResponse.status,
          statusText: createResponse.statusText,
          error: createData.error,
          fullResponse: createData,
        });
        throw new Error(
          createData.error ||
            `Failed to create chat (status: ${createResponse.status})`
        );
      }

      if (createData.success && createData.data) {
        console.log(
          `[Online Users] Chat created successfully. Navigating to chat page...`
        );

        // Get the chat_id from the response
        const chatId = createData.data.chat_id;

        // Navigate to chat page with chat_id
        router.push(`/chat?chatId=${chatId}`);
      } else {
        throw new Error(createData.error || "Failed to create chat");
      }
    } catch (err: any) {
      console.error("[Online Users] Error creating/finding chat:", err);
      setError(err.message || "Failed to start chat");
    } finally {
      setChattingWith(null);
    }
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
                  transition: "transform 0.2s, box-shadow 0.2s",
                  "&:hover": {
                    transform: "translateY(-4px)",
                    boxShadow: 3,
                  },
                }}
              >
                <CardContent>
                  <Stack spacing={2} alignItems="center">
                    <Box
                      sx={{
                        position: "relative",
                        cursor: "pointer",
                      }}
                      onClick={() => {
                        router.push(`/profile/${user.user_id}`);
                      }}
                    >
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
                          cursor: "pointer",
                        }}
                        onClick={() => {
                          router.push(`/profile/${user.user_id}`);
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
                      {/* Only show Chat button if it's not the current user */}
                      {currentUserId && user.user_id !== currentUserId && (
                        <Button
                          variant="contained"
                          size="small"
                          startIcon={
                            chattingWith === user.user_id ? (
                              <CircularProgress
                                size={14}
                                sx={{ color: "white" }}
                              />
                            ) : (
                              <ChatIcon sx={{ fontSize: 16 }} />
                            )
                          }
                          onClick={(e) => {
                            e.stopPropagation();
                            handleChatWithUser(user.user_id, user.username);
                          }}
                          disabled={chattingWith === user.user_id}
                          sx={{
                            mt: 1.5,
                            bgcolor: "#8aa79b",
                            color: "white",
                            textTransform: "none",
                            fontSize: "0.75rem",
                            px: 2,
                            py: 0.5,
                            borderRadius: 2,
                            "&:hover": {
                              bgcolor: "#7a9688",
                            },
                            "&:disabled": {
                              bgcolor: "#8aa79b",
                              opacity: 0.7,
                            },
                          }}
                        >
                          {chattingWith === user.user_id
                            ? "Opening..."
                            : "Chat"}
                        </Button>
                      )}
                      {/* Show "You" label for current user */}
                      {currentUserId && user.user_id === currentUserId && (
                        <Chip
                          label="You"
                          size="small"
                          sx={{
                            mt: 1.5,
                            bgcolor: "#e3f2fd",
                            color: "#1976d2",
                            fontSize: "0.75rem",
                            fontWeight: 600,
                          }}
                        />
                      )}
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
