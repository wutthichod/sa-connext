'use client'

import * as React from 'react'
import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { Gloock, Gantari } from 'next/font/google'
import {
  Box,
  Stack,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Button,
  CircularProgress,
  Alert,
} from '@mui/material'
import PersonOutlineIcon from '@mui/icons-material/PersonOutline'
import ChatBubbleOutlineIcon from '@mui/icons-material/ChatBubbleOutline'

// Fonts
const gloock = Gloock({ subsets: ['latin'], weight: '400' })
const gantari400 = Gantari({ subsets: ['latin'], weight: '400' })

// Small status chip
function StatusChip({ status }: { status: string }) {
  const active = status.toLowerCase() === 'available' || status.toLowerCase() === 'active'
  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        px: 1.2,
        py: 0.5,
        borderRadius: 2,
        background: active ? '#87A98D' : 'rgba(0,0,0,0.03)',
        color: active ? 'white' : '#666',
        fontSize: 15,
        width: '7vw',
        height: '4vh',
        textAlign: 'center',
      }}
    >
      {status}
    </Box>
  )
}

type Participant = {
  user_id: string
  username: string
  role?: string
}

type Event = {
  event_id: number
  name: string
  detail?: string
  location: string
  date: string
  joining_code: string
  organizer_id: string
}

export default function RoomParticipantsPageClient() {
  const params = useParams()
  const router = useRouter()
  const roomId = (params as { roomId?: string } | null)?.roomId ?? ''
  
  const [event, setEvent] = useState<Event | null>(null)
  const [participants, setParticipants] = useState<Participant[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [currentUserId, setCurrentUserId] = useState<string | null>(null)

  useEffect(() => {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token')
    if (!token) {
      router.replace('/login')
      return
    }

    // First, find the event by joining code by getting all events and filtering
    // (Note: This is a workaround since we don't have a direct endpoint to get event by joining code)
    fetchEventAndParticipants(token)
  }, [roomId, router])

  // Redirect to /join if event is not found (deleted)
  useEffect(() => {
    if (!loading && (error || !event)) {
      // Clear the lastRoomId since the event is deleted/not found
      localStorage.removeItem('lastRoomId')
      // Redirect to code entering page
      router.replace('/join')
    }
  }, [loading, error, event, router])

  const fetchEventAndParticipants = async (token: string) => {
    try {
      setLoading(true)
      setError(null)

      // Get all events to find the one with matching joining code
      const eventsRes = await fetch('/api/events', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })

      if (!eventsRes.ok) {
        throw new Error('Failed to fetch events')
      }

      const eventsData = await eventsRes.json()
      
      if (!eventsData.success || !eventsData.data) {
        throw new Error('No events found')
      }

      // Find event with matching joining code
      const allEvents = Array.isArray(eventsData.data) ? eventsData.data : [eventsData.data]
      const foundEvent = allEvents.find((e: any) => e.joining_code === roomId)

      if (!foundEvent) {
        throw new Error('Event not found')
      }

      setEvent(foundEvent)

      // Get current user to identify "You"
      const userRes = await fetch('/api/users/me', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })

      if (userRes.ok) {
        const userData = await userRes.json()
        if (userData.success && userData.data) {
          setCurrentUserId(userData.data.user_id?.toString())
        }
      }

      // Get participants
      const participantsRes = await fetch(`/api/events/${foundEvent.event_id}/participants`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      })

      if (participantsRes.ok) {
        const participantsData = await participantsRes.json()
        console.log('[Join Event] Participants data:', JSON.stringify(participantsData, null, 2))
        if (participantsData.success && participantsData.data) {
          const users = Array.isArray(participantsData.data) ? participantsData.data : [participantsData.data]
          console.log('[Join Event] Users array:', JSON.stringify(users, null, 2))
          const mappedParticipants = users.map((u: any) => {
            // Ensure user_id is properly extracted and converted to string
            let userId = '';
            if (u.user_id !== undefined && u.user_id !== null) {
              userId = String(u.user_id);
            } else if (u.userId !== undefined && u.userId !== null) {
              userId = String(u.userId);
            } else if (u.id !== undefined && u.id !== null) {
              userId = String(u.id);
            }
            console.log('[Join Event] Mapping user:', { raw: u, userId, user_id_field: u.user_id, userId_field: u.userId, id_field: u.id })
            return {
              user_id: userId,
              username: u.username || 'Unknown',
              role: userId === foundEvent.organizer_id?.toString() ? 'Organizer' : 'Participant',
            }
          })
          console.log('[Join Event] Mapped participants:', mappedParticipants)
          setParticipants(mappedParticipants)
        }
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load event data')
    } finally {
      setLoading(false)
    }
  }

  const handleLeaveRoom = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      if (!token) {
        router.replace('/login');
        return;
      }

      const res = await fetch('/api/users/leave-event', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (res.ok) {
        localStorage.removeItem('lastRoomId');
        router.push('/join');
      } else {
        const errorData = await res.json().catch(() => ({ error: 'Unknown error' }));
        alert(errorData.error || errorData.message || 'Failed to leave event');
      }
    } catch (err: any) {
      console.error('Error leaving event:', err);
      alert(err.message || 'Failed to leave event');
    }
  }

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="100vh" bgcolor="#56796F">
        <CircularProgress sx={{ color: '#fff' }} />
      </Box>
    )
  }

  // If event is not found or error, redirect will happen via useEffect
  // Show loading while redirecting
  if (error || !event) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="100vh" bgcolor="#56796F">
        <CircularProgress sx={{ color: '#fff' }} />
      </Box>
    )
  }

  return (
    <Stack direction="row" sx={{ minHeight: '100vh', bgcolor: '#56796F' }}>
      <Box component="main" sx={{ flex: 1, p: 4 }}>
        <Box sx={{ p: 0 }}>
            <Box display="flex" flexDirection="row">
              <Box display="flex" flexDirection="column" mr="2%">
                <Typography
                  variant="h4"
                  className={gloock.className}
                  sx={{
                    mb: 1,
                    fontSize: '3rem',
                    mt: 2,
                    color: '#fff',
                    fontFamily: gloock.style.fontFamily,
                  }}
                >
                  {event.name}
                </Typography>
                <Typography
                  sx={{
                    mb: 2,
                    fontSize: '1.5rem',
                    color: '#fff',
                    fontFamily: gantari400.style.fontFamily,
                  }}
                >
                  Joining Code: {roomId}
                </Typography>
                <Typography
                  sx={{
                    mb: 2,
                    fontSize: '1rem',
                    color: '#fff',
                    fontFamily: gantari400.style.fontFamily,
                  }}
                >
                  Location: {event.location} | Date: {new Date(event.date).toLocaleString()}
                </Typography>
              </Box>
              <Button
                variant="outlined"
                onClick={handleLeaveRoom}
                sx={{ 
                  alignSelf: 'center',
                  color: '#fff',
                  borderColor: '#fff',
                  borderWidth: 2,
                  px: 3,
                  py: 1.5,
                  fontSize: '1rem',
                  textTransform: 'none',
                  '&:hover': {
                    borderColor: '#fff',
                    borderWidth: 2,
                    bgcolor: 'rgba(255, 255, 255, 0.1)',
                  },
                  ml: 'auto',
                  width: '150px',
                  height: '50px',
                }}
              >
                Leave Event
              </Button>
            </Box>

            <Box sx={{ px: 3, pb: 4, bgcolor: '#fff', borderRadius: 5, p: 2 }}>
              <TableContainer
                component={Paper}
                sx={{
                  width: '100%',
                  boxShadow: 0,
                  borderRadius: 1,
                  overflow: 'auto',
                  maxHeight: 'calc(80vh - 200px)',
                }}
              >
                <Table
                  sx={{ width: '100%', fontFamily: gantari400.style.fontFamily }}
                  aria-label="participants table"
                  stickyHeader
                >
                  <TableHead>
                    <TableRow>
                      <TableCell sx={{ fontSize: '1.5rem' }}>Name</TableCell>
                      <TableCell sx={{ fontSize: '1.5rem' }}>Role</TableCell>
                      <TableCell sx={{ fontSize: '1.5rem' }}>Status</TableCell>
                      <TableCell align="center" sx={{ fontSize: '1.5rem' }}>
                        Actions
                      </TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {participants.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={4} align="center" sx={{ fontSize: '1rem' }}>
                          No participants yet
                        </TableCell>
                      </TableRow>
                    ) : (
                      participants.map((p) => (
                        <TableRow key={p.user_id} hover>
                          <TableCell component="th" scope="row" sx={{ fontSize: '1rem' }}>
                            {p.username}{currentUserId === p.user_id ? ' (You)' : ''}
                          </TableCell>
                          <TableCell sx={{ fontSize: '1rem' }}>{p.role || 'Participant'}</TableCell>
                          <TableCell sx={{ fontSize: '1rem' }}>
                            <StatusChip status="Available" />
                          </TableCell>
                          <TableCell align="center" sx={{ fontSize: '1rem' }}>
                            <IconButton 
                              aria-label="profile"
                              onClick={() => router.push(`/profile/${p.user_id}`)}
                              sx={{ '&:hover': { bgcolor: 'rgba(0,0,0,0.04)' } }}
                            >
                              <PersonOutlineIcon />
                            </IconButton>
                            <IconButton 
                              aria-label="chat"
                              onClick={async () => {
                                try {
                                  const token = localStorage.getItem('token') || sessionStorage.getItem('token');
                                  if (!token) {
                                    router.replace('/login');
                                    return;
                                  }

                                  // Ensure user_id is a string
                                  const recipientId = String(p.user_id || '');
                                  console.log('[Join Event] Creating chat with user:', recipientId, '(original:', p.user_id, 'type:', typeof p.user_id, ')');

                                  if (!recipientId || recipientId === 'undefined' || recipientId === 'null') {
                                    alert('Invalid user ID');
                                    return;
                                  }

                                  // Create chat with this user
                                  const res = await fetch('/api/chats', {
                                    method: 'POST',
                                    headers: {
                                      'Authorization': `Bearer ${token}`,
                                      'Content-Type': 'application/json',
                                    },
                                    body: JSON.stringify({
                                      recipient_id: recipientId,
                                    }),
                                  });

                                  console.log('[Join Event] Chat creation response status:', res.status);

                                  if (res.ok) {
                                    const data = await res.json();
                                    console.log('[Join Event] Chat created successfully:', data);
                                    // Navigate to chat page after creating chat
                                    router.push('/chat');
                                  } else {
                                    const errorData = await res.json().catch(() => ({ error: 'Unknown error' }));
                                    console.error('[Join Event] Chat creation failed:', errorData);
                                    alert(errorData.error || errorData.message || 'Failed to create chat');
                                  }
                                } catch (err: any) {
                                  console.error('[Join Event] Error creating chat:', err);
                                  alert(err.message || 'Failed to create chat');
                                }
                              }}
                              sx={{ '&:hover': { bgcolor: 'rgba(0,0,0,0.04)' } }}
                            >
                              <ChatBubbleOutlineIcon />
                            </IconButton>
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            </Box>
          </Box>
      </Box>
    </Stack>
  )
}
