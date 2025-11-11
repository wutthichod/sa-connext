// components/Sidebar.tsx
'use client'
import React from 'react'
import {
  Box,
  IconButton,
  Stack,
  Tooltip,
  Typography,
} from '@mui/material'
import KeyboardDoubleArrowLeftIcon from '@mui/icons-material/KeyboardDoubleArrowLeft'
import MeetingRoomOutlinedIcon from '@mui/icons-material/MeetingRoomOutlined'
import HomeOutlinedIcon from '@mui/icons-material/HomeOutlined'
import ChatBubbleOutlineIcon from '@mui/icons-material/ChatBubbleOutline'
import PersonOutlineIcon from '@mui/icons-material/PersonOutline'
import LogoutIcon from '@mui/icons-material/Logout'
import EventIcon from '@mui/icons-material/Event'
import { usePathname, useRouter } from 'next/navigation'
import Image from 'next/image'

const ICON_COLOR = '#8aa79b'
const ICON_HOVER_BG = 'rgba(138,167,155,0.12)'
const ICON_SELECTED_BG = 'rgba(138,167,155,0.22)'

interface SidebarProps {
  open: boolean
  setOpen: React.Dispatch<React.SetStateAction<boolean>>
}

export default function Sidebar({ open, setOpen }: SidebarProps) {
  const pathname = usePathname() ?? '/'
  const router = useRouter()

  const items = [
    {
      key: 'Room',
      label: 'Rooms',
      icon: <MeetingRoomOutlinedIcon sx={{ fontSize: 26 }} />,
      path: '/join',
    },
    {
      key: 'Create',
      label: 'Home',
      icon: <HomeOutlinedIcon sx={{ fontSize: 26 }} />,
      path: '/create',
    },
    {
      key: 'MyEvents',
      label: 'My Events',
      icon: <EventIcon sx={{ fontSize: 26 }} />,
      path: '/my-events',
    },
    {
      key: 'Chat',
      label: 'Chat',
      icon: <ChatBubbleOutlineIcon sx={{ fontSize: 26 }} />,
      path: '/chat',
    },
    {
      key: 'Profile',
      label: 'Profile',
      icon: <PersonOutlineIcon sx={{ fontSize: 26 }} />,
      path: '/profile',
    },
  ]

  const selectedFromPath = React.useMemo(() => {
    const idx = items.findIndex((it) => {
      return pathname === it.path || pathname.startsWith(it.path + '/')
    })
    return idx >= 0 ? idx : null
  }, [pathname])

  const [selected, setSelected] = React.useState<number | null>(selectedFromPath)

  React.useEffect(() => {
    setSelected(selectedFromPath)
  }, [selectedFromPath])

  const handleClick = (path: string, idx: number) => {
    setSelected(idx)
    router.push(path)
  }

  return (
    <Box
      component="aside"
      sx={{
        height: '100vh',
        width: open ? 270 : 80,
        transition: 'width 250ms ease',
        bgcolor: 'white',
        borderRight: '1px solid rgba(0,0,0,0.06)',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        overflow: 'hidden',
        position: 'fixed',
        top: 0,
        left: 0,
        py: 2,
        boxSizing: 'border-box',
      }}
    >
      {/* Expand/Collapse Button */}
      <IconButton
        aria-label={open ? 'collapse sidebar' : 'expand sidebar'}
        onClick={() => setOpen(prev => !prev)}
        size="medium"
        sx={{
          color: ICON_COLOR,
          bgcolor: 'transparent',
          '&:hover': { bgcolor: ICON_HOVER_BG },
          mb: 2,
        }}
      >
        {open ? (
          <KeyboardDoubleArrowLeftIcon />
        ) : (
          <Image
            src="/images/primary_logo_alt.png"
            alt="Expand"
            width={36}
            height={36}
            style={{ objectFit: 'contain' }}
          />
        )}
      </IconButton>

      {/* Spacer / Portrait Area */}
      <Box
        sx={{
          width: '100%',
          display: 'flex',
          justifyContent: 'center',
          mb: 3,
          minHeight: 48,
        }}
      >
        {open ? <Box sx={{ width: '85%' }} /> : <Box sx={{ width: 40 }} />}
      </Box>

      {/* Menu Items */}
      <Stack
        spacing={1.5}
        sx={{
          width: '100%',
          alignItems: 'center',
        }}
      >
        {items.map((it, idx) => {
          const isSelected = selected === idx
          return (
            <Box
              key={it.key}
              onClick={() => handleClick(it.path, idx)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: open ? 'flex-start' : 'center',
                gap: 1,
                width: open ? '85%' : 'auto',
                px: open ? 1 : 0,
                py: 0.5,
                borderRadius: 1,
                cursor: 'pointer',
                bgcolor: isSelected ? ICON_SELECTED_BG : 'transparent',
                '&:hover': { bgcolor: ICON_HOVER_BG },
              }}
            >
              <Tooltip title={open ? '' : it.label} placement="right" arrow>
                <IconButton
                  size="small"
                  aria-label={it.label}
                  sx={{
                    color: ICON_COLOR,
                    '&:hover': { bgcolor: 'transparent' },
                  }}
                >
                  {it.icon}
                </IconButton>
              </Tooltip>

              {open && (
                <Typography variant="body2" sx={{ ml: 0.5, fontFamily: 'Gloock', color: 'rgba(40, 82, 64, 0.8)' }}>
                  {it.label}
                </Typography>
              )}
            </Box>
          )
        })}
      </Stack>

      <Box sx={{ flex: 1 }} />

      {/* Logout Button */}
      <Box
        onClick={() => {
          localStorage.removeItem('token')
          sessionStorage.removeItem('token')
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('lastRoomId')
          router.push('/login')
        }}
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: open ? 'flex-start' : 'center',
          gap: 1,
          width: open ? '85%' : 'auto',
          px: open ? 1 : 0,
          py: 0.5,
          borderRadius: 1,
          cursor: 'pointer',
          bgcolor: 'transparent',
          '&:hover': { bgcolor: 'rgba(211, 47, 47, 0.12)' },
          mb: 2,
        }}
      >
        <Tooltip title={open ? '' : 'Logout'} placement="right" arrow>
          <IconButton
            size="small"
            aria-label="Logout"
            sx={{
              color: '#d32f2f',
              '&:hover': { bgcolor: 'transparent' },
            }}
          >
            <LogoutIcon sx={{ fontSize: 26 }} />
          </IconButton>
        </Tooltip>

        {open && (
          <Typography variant="body2" sx={{ ml: 0.5, color: '#d32f2f' }}>
            Logout
          </Typography>
        )}
      </Box>
    </Box>
  )
}
