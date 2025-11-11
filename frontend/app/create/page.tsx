'use client'

import React, { useState, useEffect } from 'react'
import { Box, Button, Paper, TextField, Typography, Alert } from '@mui/material'
import { useRouter } from 'next/navigation'

export default function CreatePage() {
  const [name, setName] = useState('')
  const [detail, setDetail] = useState('')
  const [location, setLocation] = useState('')
  const [date, setDate] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token')
    if (!token) {
      router.replace('/login')
    }
  }, [router])

  const handleCreate = async (e?: React.FormEvent) => {
    e?.preventDefault()
    if (!name.trim() || !location.trim() || !date.trim()) {
      setError('Name, location, and date are required')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token')
      if (!token) {
        router.replace('/login')
        return
      }

      // Convert datetime-local format to RFC3339 (ISO 8601 with timezone)
      // datetime-local format: "2024-01-15T14:30" (no timezone, interpreted as local time)
      // RFC3339 format: "2024-03-15T10:00:00Z" (UTC with Z suffix, no milliseconds)
      const dateObj = new Date(date.trim())
      if (isNaN(dateObj.getTime())) {
        throw new Error('Invalid date format')
      }
      // Convert to ISO string and remove milliseconds to match backend format
      // Format: YYYY-MM-DDTHH:mm:ssZ (e.g., "2024-03-15T10:00:00Z")
      const dateRFC3339 = dateObj.toISOString().replace(/\.\d{3}Z$/, 'Z')

      const response = await fetch('/api/events', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: name.trim(),
          detail: detail.trim() || undefined,
          location: location.trim(),
          date: dateRFC3339,
        }),
      })

      if (!response.ok) {
        const errorData = await response.json()
        throw new Error(errorData.error || 'Failed to create event')
      }

      const data = await response.json()
      
      if (data.success && data.data) {
        const eventId = data.data.event_id
        const joiningCode = data.data.joining_code
        
        // Redirect to the created event page
        router.push(`/create/${joiningCode}`)
      } else {
        throw new Error('Invalid response from server')
      }
    } catch (err: any) {
      setError(err.message || 'Failed to create event')
    } finally {
      setLoading(false)
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
        <Typography
          variant="h3"
          sx={{
            fontFamily: 'Gloock',
            color: '#fff',
            mb: 4,
            fontWeight: 600,
            letterSpacing: 0.5,
          }}
        >
          Connext
        </Typography>

        <Paper
          elevation={3}
          sx={{
            display: 'inline-block',
            p: 2,
            px: 3,
            borderRadius: 1,
            backgroundColor: 'rgba(255,255,255,0.97)',
            minWidth: 320,
          }}
        >
          <form onSubmit={handleCreate}>
            {error && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {error}
              </Alert>
            )}

            <TextField
              variant="outlined"
              placeholder="Event Name *"
              value={name}
              onChange={(e) => setName(e.target.value)}
              fullWidth
              required
              sx={{
                mb: 1.5,
                '& .MuiOutlinedInput-root': { borderRadius: '6px' },
              }}
              InputProps={{ style: { background: '#fff' } }}
            />

            <TextField
              variant="outlined"
              placeholder="Location *"
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              fullWidth
              required
              sx={{
                mb: 1.5,
                '& .MuiOutlinedInput-root': { borderRadius: '6px' },
              }}
              InputProps={{ style: { background: '#fff' } }}
            />

            <TextField
              variant="outlined"
              placeholder="Date *"
              type="datetime-local"
              value={date}
              onChange={(e) => setDate(e.target.value)}
              fullWidth
              required
              InputLabelProps={{ shrink: true }}
              sx={{
                mb: 1.5,
                '& .MuiOutlinedInput-root': { borderRadius: '6px' },
              }}
              InputProps={{ style: { background: '#fff' } }}
            />

            <TextField
              variant="outlined"
              placeholder="Details (optional)"
              value={detail}
              onChange={(e) => setDetail(e.target.value)}
              fullWidth
              multiline
              rows={3}
              sx={{
                mb: 1.5,
                '& .MuiOutlinedInput-root': { borderRadius: '6px' },
              }}
              InputProps={{ style: { background: '#fff' } }}
            />

            <Button
              type="submit"
              fullWidth
              disabled={loading}
              sx={{
                mt: 0.5,
                backgroundColor: '#8aa79b',
                color: '#fff',
                textTransform: 'none',
                '&:hover': { backgroundColor: '#7d9b87' },
              }}
            >
              {loading ? 'Creating...' : 'Create Event'}
            </Button>
          </form>
        </Paper>
      </Box>
    </Box>
  )
}
