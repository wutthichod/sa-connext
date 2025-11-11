'use client'

import React, { type ReactNode } from "react";
import { usePathname } from 'next/navigation';
import { useState, useEffect } from 'react';
import dynamic from 'next/dynamic';
import { Box } from "@mui/material";
import './globals.css';

// Dynamically import Sidebar with no SSR to avoid hydration issues
const Sidebar = dynamic(() => import("./components/Sidebar"), {
  ssr: false,
});

interface RootLayoutProps {
  children: ReactNode;
}

export default function RootLayout({ children }: RootLayoutProps) {
  const pathname = usePathname();
  const [mounted, setMounted] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  // Hide sidebar on these routes - check after mount to avoid hydration mismatch
  const excludedRoutes = ['/login', '/register'];
  const hideLayout = mounted && pathname && excludedRoutes.includes(pathname);

  const noScrollRoutes = ['/join', '/create'];
  const shouldHideScroll = mounted && pathname ? noScrollRoutes.includes(pathname) : false;

  const bodyStyle: React.CSSProperties = {
    margin: 0,
    padding: 0,
    height: '100vh',
    width: '100vw',
    overflow: shouldHideScroll ? 'hidden' : 'auto',
  };

  // Always render the exact same structure on server and client
  // Server: hideLayout = false (because mounted = false)
  // Client initial: hideLayout = false (because mounted = false initially)
  // Client after mount: hideLayout updates based on pathname
  return (
    <html lang="en">
      <body style={bodyStyle}>
        <Box
          display="flex"
          flexDirection="row"
          minHeight="100vh"
          width="100vw"
          suppressHydrationWarning
        >
          <Box
            height="100vh"
            bgcolor="background.paper"
            sx={{ display: hideLayout ? 'none' : 'block' }}
            suppressHydrationWarning
          >
            {mounted && !hideLayout && <Sidebar open={sidebarOpen} setOpen={setSidebarOpen} />}
          </Box>
          <Box flex="1" p={0} sx={{ paddingLeft: sidebarOpen ? '270px' : '80px', transition: 'padding-left 250ms ease' }}>
            {children}
          </Box>
        </Box>
      </body>
    </html>
  );
}
