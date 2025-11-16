'use client'

import * as React from 'react'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Gloock } from 'next/font/google'
import {
  Box,
  Stack,
  Typography,
  Paper,
  CircularProgress,
  Alert,
  Button,
  Card,
  CardContent,
  CardActions,
  Chip,
} from '@mui/material'
import EventIcon from '@mui/icons-material/Event'
import LocationOnIcon from '@mui/icons-material/LocationOn'
import CalendarTodayIcon from '@mui/icons-material/CalendarToday'
import PeopleIcon from '@mui/icons-material/People'

const gloock = Gloock({ subsets: ['latin'], weight: '400' })

type Event = {
  event_id: number
  name: string
  detail?: string
  location: string
  date: string
  joining_code: string
  organizer_id: string
}

export default function MyEventsPage() {
  const router = useRouter()
  const [events, setEvents] = useState<Event[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token')
    if (!token) {
      router.replace('/login')
      return
    }

    fetchMyEvents(token)
  }, [router])

  const fetchMyEvents = async (token: string) => {
    try {
      setLoading(true)
      setError(null)

      const res = await fetch('/api/events/my', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })

      if (!res.ok) {
        const errorData = await res.json()
        throw new Error(errorData.error || 'Failed to fetch events')
      }

      const data = await res.json()
      if (data.success && data.data) {
        const eventsList = Array.isArray(data.data) ? data.data : [data.data]
        setEvents(eventsList)
      } else {
        setEvents([])
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load events')
    } finally {
      setLoading(false)
    }
  }

  const handleViewEvent = (joiningCode: string) => {
    router.push(`/create/${joiningCode}`)
  }

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="100vh" bgcolor="#56796F">
        <CircularProgress sx={{ color: '#fff' }} />
      </Box>
    )
  }

  return (
    <Box
      component="main"
      sx={{
        minHeight: '100vh',
        bgcolor: '#56796F',
        p: 4,
      }}
    >
      <Box sx={{ maxWidth: 1200, mx: 'auto' }}>
        <Typography
          variant="h3"
          className={gloock.className}
          sx={{
            mb: 4,
            color: '#fff',
            fontSize: '2.5rem',
            fontFamily: gloock.style.fontFamily,
          }}
        >
          My Events
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}

        {events.length === 0 ? (
          <Paper
            sx={{
              p: 6,
              textAlign: 'center',
              bgcolor: 'rgba(255, 255, 255, 0.97)',
              borderRadius: 2,
            }}
          >
            <EventIcon sx={{ fontSize: 64, color: '#ccc', mb: 2 }} />
            <Typography variant="h6" sx={{ mb: 2, color: '#666' }}>
              No events found
            </Typography>
            <Typography variant="body2" sx={{ mb: 3, color: '#999' }}>
              You haven't created any events yet. Create your first event to get started!
            </Typography>
            <Button
              variant="contained"
              onClick={() => router.push('/create')}
              sx={{
                bgcolor: '#56796F',
                '&:hover': { bgcolor: '#4a6b5f' },
              }}
            >
              Create Event
            </Button>
          </Paper>
        ) : (
          <Stack spacing={3}>
            {events.map((event) => (
              <Card
                key={event.event_id}
                sx={{
                  bgcolor: 'rgba(255, 255, 255, 0.97)',
                  borderRadius: 2,
                  boxShadow: 2,
                  transition: 'transform 0.2s, box-shadow 0.2s',
                  '&:hover': {
                    transform: 'translateY(-4px)',
                    boxShadow: 4,
                  },
                }}
              >
                <CardContent>
                  <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                    <Typography
                      variant="h5"
                      className={gloock.className}
                      sx={{
                        color: '#384F52',
                        fontFamily: gloock.style.fontFamily,
                        fontSize: '1.8rem',
                        fontWeight: 'bold',
                      }}
                    >
                      {event.name}
                    </Typography>
                    <Chip
                      label={`Code: ${event.joining_code}`}
                      size="small"
                      sx={{
                        bgcolor: '#F1F0CC',
                        color: '#384F52',
                        fontWeight: 'bold',
                      }}
                    />
                  </Box>

                  {event.detail && (
                    <Typography
                      variant="body1"
                      sx={{ mb: 2, color: '#666', lineHeight: 1.6 }}
                    >
                      {event.detail}
                    </Typography>
                  )}

                  <Stack spacing={1.5} mt={2}>
                    <Box display="flex" alignItems="center" gap={1}>
                      <LocationOnIcon sx={{ color: '#56796F', fontSize: 20 }} />
                      <Typography variant="body2" sx={{ color: '#666' }}>
                        {event.location}
                      </Typography>
                    </Box>

                    <Box display="flex" alignItems="center" gap={1}>
                      <CalendarTodayIcon sx={{ color: '#56796F', fontSize: 20 }} />
                      <Typography variant="body2" sx={{ color: '#666' }}>
                        {new Date(event.date).toLocaleString()}
                      </Typography>
                    </Box>
                  </Stack>
                </CardContent>

                <CardActions sx={{ px: 2, pb: 2 }}>
                  <Button
                    variant="contained"
                    startIcon={<PeopleIcon />}
                    onClick={() => handleViewEvent(event.joining_code)}
                    sx={{
                      bgcolor: '#56796F',
                      '&:hover': { bgcolor: '#4a6b5f' },
                      textTransform: 'none',
                    }}
                  >
                    View Event
                  </Button>
                </CardActions>
              </Card>
            ))}
          </Stack>
        )}
      </Box>
    </Box>
  )
}

