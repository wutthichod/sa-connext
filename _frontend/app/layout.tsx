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

  useEffect(() => {
    setMounted(true);
  }, []);

  // Hide sidebar on these routes - check after mount to avoid hydration mismatch
  const excludedRoutes = ['/login', '/register'];
  const hideLayout = mounted && pathname && excludedRoutes.includes(pathname);

  // Always render the exact same structure on server and client
  // Server: hideLayout = false (because mounted = false)
  // Client initial: hideLayout = false (because mounted = false initially)
  // Client after mount: hideLayout updates based on pathname
  return (
    <html lang="en">
      <body style={{ margin: 0, padding: 0, height: '100vh', width: '100vw' }}>
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
            {mounted && !hideLayout && <Sidebar />}
          </Box>
          <Box flex="1" p={0}>
            {children}
          </Box>
        </Box>
      </body>
    </html>
  );
}
