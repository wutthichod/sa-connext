// components/Sidebar.tsx
"use client";
import React from "react";
import Image from "next/image";
import { Box, IconButton, Stack, Tooltip, Typography } from "@mui/material";
import KeyboardDoubleArrowLeftIcon from "@mui/icons-material/KeyboardDoubleArrowLeft";
import MeetingRoomOutlinedIcon from "@mui/icons-material/MeetingRoomOutlined";
import HomeOutlinedIcon from "@mui/icons-material/HomeOutlined";
import ChatBubbleOutlineIcon from "@mui/icons-material/ChatBubbleOutline";
import GroupsIcon from "@mui/icons-material/Groups";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import FilterHdrIcon from "@mui/icons-material/FilterHdr";
import LogoutIcon from "@mui/icons-material/Logout";
import EventIcon from "@mui/icons-material/Event";
import PeopleIcon from "@mui/icons-material/People";
import { usePathname, useRouter } from "next/navigation";

const ICON_COLOR = "#8aa79b";
const ICON_HOVER_BG = "rgba(138,167,155,0.12)";
const ICON_SELECTED_BG = "rgba(138,167,155,0.22)";

export default function Sidebar() {
  const [open, setOpen] = React.useState(false);
  const pathname = usePathname() ?? "/";
  const router = useRouter();

  const items = [
    {
      key: "Room",
      label: "Rooms",
      icon: <MeetingRoomOutlinedIcon sx={{ fontSize: 26 }} />,
      path: "/join",
    },
    {
      key: "Create",
      label: "Home",
      icon: <HomeOutlinedIcon sx={{ fontSize: 26 }} />,
      path: "/create",
    },
    {
      key: "MyEvents",
      label: "My Events",
      icon: <EventIcon sx={{ fontSize: 26 }} />,
      path: "/my-events",
    },
    {
      key: "Chat",
      label: "Chat",
      icon: <ChatBubbleOutlineIcon sx={{ fontSize: 26 }} />,
      path: "/chat",
    },
    {
      key: "GroupChat",
      label: "Group Chat",
      icon: <GroupsIcon sx={{ fontSize: 26 }} />,
      path: "/group-chat",
    },
    {
      key: "OnlineUsers",
      label: "Online Users",
      icon: <PeopleIcon sx={{ fontSize: 26 }} />,
      path: "/online-users",
    },
    {
      key: "Profile",
      label: "Profile",
      icon: <PersonOutlineIcon sx={{ fontSize: 26 }} />,
      path: "/profile",
    },
  ];

  // derive selected index from pathname so selection persists across navigation/refresh
  const selectedFromPath = React.useMemo(() => {
    const idx = items.findIndex((it) => {
      // match either exact or subtree (e.g. /chat/123)
      return pathname === it.path || pathname.startsWith(it.path + "/");
    });
    return idx >= 0 ? idx : null;
  }, [pathname]);

  const [selected, setSelected] = React.useState<number | null>(
    selectedFromPath
  );

  React.useEffect(() => {
    setSelected(selectedFromPath);
  }, [selectedFromPath]);

  const handleClick = (path: string, idx: number) => {
    // update UI selected state immediately for snappy UX
    setSelected(idx);
    // navigate client-side without full page reload
    router.push(path);
  };

  return (
    <Box
      component="aside"
      sx={{
        height: "100vh",
        width: open ? 270 : 80,
        transition: "width 250ms ease",
        bgcolor: "white",
        borderRight: "1px solid rgba(0,0,0,0.06)",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        overflow: "hidden",
        position: "relative",
        py: 2,
        boxSizing: "border-box",
      }}
    >
      {/* Expand/Collapse Button (top, always visible) */}
      <IconButton
        aria-label={open ? "collapse sidebar" : "expand sidebar"}
        onClick={() => setOpen((prev) => !prev)}
        size="medium"
        sx={{
          color: ICON_COLOR,
          bgcolor: "transparent",
          "&:hover": { bgcolor: ICON_HOVER_BG },
          mb: 2,
        }}
      >
        {open ? (
          <KeyboardDoubleArrowLeftIcon />
        ) : (
          // show your logo image when collapsed (same as you used earlier)
          <Image
            src="/images/primary_logo_alt.png"
            alt="Expand"
            width={36}
            height={36}
            style={{ objectFit: "contain" }}
          />
        )}
      </IconButton>

      {/* Spacer in the place of the portrait */}
      <Box
        sx={{
          width: "100%",
          display: "flex",
          justifyContent: "center",
          mb: 3,
          minHeight: 48,
        }}
      >
        {open ? <Box sx={{ width: "85%" }} /> : <Box sx={{ width: 40 }} />}
      </Box>

      {/* Menu Items */}
      <Stack
        spacing={1.5}
        sx={{
          width: "100%",
          alignItems: "center",
        }}
      >
        {items.map((it, idx) => {
          const isSelected = selected === idx;
          return (
            <Box
              key={it.key}
              onClick={() => handleClick(it.path, idx)}
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: open ? "flex-start" : "center",
                gap: 1,
                width: open ? "85%" : "auto",
                px: open ? 1 : 0,
                py: 0.5,
                borderRadius: 1,
                cursor: "pointer",
                bgcolor: isSelected ? ICON_SELECTED_BG : "transparent",
                "&:hover": { bgcolor: ICON_HOVER_BG },
              }}
            >
              <Tooltip title={open ? "" : it.label} placement="right" arrow>
                <IconButton
                  size="small"
                  aria-label={it.label}
                  sx={{
                    color: ICON_COLOR,
                    "&:hover": { bgcolor: "transparent" },
                  }}
                >
                  {it.icon}
                </IconButton>
              </Tooltip>

              {open && (
                <Typography
                  variant="body2"
                  sx={{
                    ml: 0.5,
                    fontFamily: "Gloock, serif",
                    color: "rgba(40, 82, 64, 0.8)",
                    letterSpacing: "0.2px",
                  }}
                >
                  {it.label}
                </Typography>
              )}
            </Box>
          );
        })}
      </Stack>

      <Box sx={{ flex: 1 }} />

      {/* Logout Button */}
      <Box
        onClick={() => {
          // Clear all auth data
          localStorage.removeItem("token");
          sessionStorage.removeItem("token");
          localStorage.removeItem("isLoggedIn");
          localStorage.removeItem("lastRoomId");

          // Force reload to ensure WebSocket is properly closed
          window.location.href = "/login";
        }}
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: open ? "flex-start" : "center",
          gap: 1,
          width: open ? "85%" : "auto",
          px: open ? 1 : 0,
          py: 0.5,
          borderRadius: 1,
          cursor: "pointer",
          bgcolor: "transparent",
          "&:hover": { bgcolor: "rgba(211, 47, 47, 0.12)" },
          mb: 2,
        }}
      >
        <Tooltip title={open ? "" : "Logout"} placement="right" arrow>
          <IconButton
            size="small"
            aria-label="Logout"
            sx={{
              color: "#d32f2f",
              "&:hover": { bgcolor: "transparent" },
            }}
          >
            <LogoutIcon sx={{ fontSize: 26 }} />
          </IconButton>
        </Tooltip>

        {open && (
          <Typography
            variant="body2"
            sx={{ ml: 0.5, color: "#d32f2f", fontFamily: "Gloock, serif" }}
          >
            Logout
          </Typography>
        )}
      </Box>
    </Box>
  );
}
