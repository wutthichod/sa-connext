'use client'

import React, { useState, useEffect } from 'react'
import { Box, Button, Paper, TextField, Typography } from '@mui/material'
import { useRouter, usePathname } from 'next/navigation'

export default function JoinPage() {
  const [roomId, setRoomId] = useState('')
  const router = useRouter()
  const pathname = usePathname() ?? '/'

  useEffect(() => {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token')
    if (!token) {
      router.replace('/login')
      return
    }

    // Redirect to last room if exactly on /join, but only if we have a valid room ID
    // If the event was deleted, the lastRoomId will be cleared by the [roomId] page
    if (pathname === '/join' || pathname === '/join/') {
      const lastRoomId = localStorage.getItem('lastRoomId')
      if (lastRoomId && lastRoomId.trim() !== '') {
        // Small delay to allow any cleanup from previous navigation
        const timer = setTimeout(() => {
          router.replace(`/join/${lastRoomId}`)
        }, 100)
        return () => clearTimeout(timer)
      }
    }
  }, [pathname, router])

  const handleJoin = async (e?: React.FormEvent) => {
    e?.preventDefault()
    const trimmedRoomId = roomId.trim()
    if (!trimmedRoomId) {
      alert('Please enter a joining code')
      return
    }

    const token = localStorage.getItem('token') || sessionStorage.getItem('token')
    if (!token) {
      router.replace('/login')
      return
    }

    try {
      const res = await fetch('/api/join', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          joining_code: trimmedRoomId
        })
      })

      if (!res.ok) {
        const data = await res.json().catch(() => null)
        throw new Error(data?.error || data?.message || `Join failed with status ${res.status}`)
      }

      const data = await res.json()
      
      // Success - Save last room and redirect
      localStorage.setItem('lastRoomId', trimmedRoomId)
      router.push(`/join/${trimmedRoomId}`)
    } catch (err: any) {
      alert(err.message || 'Unknown error')
    }
  }

  return (
    <Box
      component="main"
      sx={{
        height: '100%',
        width: '100%',
        display: 'flex',
        bgcolor: '#5b756c',
        alignItems: 'center',
        justifyContent: 'center',
        overflow: 'hidden'
      }}
    >
      <Box textAlign="center" sx={{ transform: 'translateY(-6px)' }}>
        <Typography variant="h3" sx={{ color: '#fff', mb: 4, fontFamily: 'Gloock' }}>Connext</Typography>

        <Paper
          elevation={3}
          sx={{
            display: 'inline-block',
            p: 2,
            px: 3,
            borderRadius: 1,
            backgroundColor: 'rgba(255,255,255,0.97)',
            minWidth: 320
          }}
        >
          <form onSubmit={handleJoin}>
            <TextField
              variant="outlined"
              placeholder="Room ID"
              value={roomId}
              onChange={(e) => setRoomId(e.target.value)}
              fullWidth
              sx={{ mb: 1.5 }}
            />

            <Button
              type="submit"
              fullWidth
              sx={{ backgroundColor: '#8aa79b', color: '#fff', '&:hover': { backgroundColor: '#7d9b87' } }}

            >
              Join
            </Button>
          </form>
        </Paper>
      </Box>
    </Box>
  )
}