'use client'

import React, { useState, useEffect } from 'react';
import { Box, Button, TextField, Typography, Paper, CircularProgress, Alert, Stack } from '@mui/material';
import { Gloock, Gantari } from 'next/font/google';
import { useRouter, useParams } from 'next/navigation';

const gloock = Gloock({ subsets: ['latin'], weight: '400' });
const gantari400 = Gantari({ subsets: ['latin'], weight: '400' });

interface UserProfile {
  user_id: string;
  username: string;
  contact?: {
    email?: string;
    phone?: string;
  };
  education?: {
    university?: string;
    major?: string;
  };
  job_title?: string;
  interests?: string[];
}

export default function ViewUserProfilePage() {
  const params = useParams();
  const router = useRouter();
  const userId = (params as { userId?: string } | null)?.userId ?? '';
  
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) {
      router.replace('/login');
      return;
    }

    if (!userId) {
      setError('User ID is required');
      setLoading(false);
      return;
    }

    fetchProfile(token);
  }, [userId, router]);

  const fetchProfile = async (token: string) => {
    try {
      setLoading(true);
      setError(null);

      const response = await fetch(`/api/users/${userId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to fetch profile');
      }

      const data = await response.json();
      if (data.success && data.data) {
        setProfile(data.data);
      } else {
        throw new Error('Invalid response from server');
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load profile');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="100vh" bgcolor="#56796F">
        <CircularProgress sx={{ color: '#fff' }} />
      </Box>
    );
  }

  if (error || !profile) {
    return (
      <Box sx={{ minHeight: '100vh', bgcolor: '#56796F', p: 4 }}>
        <Box sx={{ bgcolor: '#fff', p: 6, borderRadius: 1 }}>
          <Alert severity="error" sx={{ mb: 2 }}>
            {error || 'Profile not found'}
          </Alert>
          <Button onClick={() => router.back()} variant="contained">
            Go Back
          </Button>
        </Box>
      </Box>
    );
  }

  return (
    <Stack direction="row" sx={{ minHeight: '100vh', bgcolor: '#56796F' }}>
      <Box component="main" sx={{ flex: 1, p: 4, px: 10 }}>
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
                {profile.username}'s Profile
              </Typography>
            </Box>
            <Button
              variant="outlined"
              onClick={() => router.back()}
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
              Go Back
            </Button>
          </Box>
          <Box sx={{ px: 3, pb: 4, bgcolor: '#fff', borderRadius: 5, p: 5 }}>
            <Box display="flex" flexDirection="column" gap={3}>
              <TextField
                label="Username"
                value={profile?.username || ''}
                fullWidth
                disabled
                variant="standard"
                sx={{
                  '& .MuiInputLabel-root': {
                    fontSize: '1.5rem',
                    color: 'black',
                    fontWeight: 'bold',
                    paddingBottom: '8px',
                  },
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: 'gray', // For Safari
                    color: 'gray', // For other browsers
                  },
                  '& .MuiInput-underline:before': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:after': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:hover:not(.Mui-disabled):before': {
                    borderBottom: 'none !important',
                  },
                }}
              />
              <TextField
                label="Email"
                value={profile?.contact?.email || ''}
                fullWidth
                disabled
                variant="standard"
                sx={{
                  '& .MuiInputLabel-root': {
                    fontSize: '1.5rem',
                    color: 'black',
                    fontWeight: 'bold',
                    paddingBottom: '8px',
                  },
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: 'gray', // For Safari
                    color: 'gray', // For other browsers
                  },
                  '& .MuiInput-underline:before': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:after': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:hover:not(.Mui-disabled):before': {
                    borderBottom: 'none !important',
                  },
                }}
              />
              <TextField
                label="Phone"
                value={profile?.contact?.phone || ''}
                fullWidth
                disabled
                variant="standard"
                sx={{
                  '& .MuiInputLabel-root': {
                    fontSize: '1.5rem',
                    color: 'black',
                    fontWeight: 'bold',
                    paddingBottom: '8px',
                  },
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: 'gray', // For Safari
                    color: 'gray', // For other browsers
                  },
                  '& .MuiInput-underline:before': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:after': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:hover:not(.Mui-disabled):before': {
                    borderBottom: 'none !important',
                  },
                }}
              />
              <TextField
                label="University"
                value={profile?.education?.university || ''}
                fullWidth
                disabled
                variant="standard"
                sx={{
                  '& .MuiInputLabel-root': {
                    fontSize: '1.5rem',
                    color: 'black',
                    fontWeight: 'bold',
                    paddingBottom: '8px',
                  },
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: 'gray', // For Safari
                    color: 'gray', // For other browsers
                  },
                  '& .MuiInput-underline:before': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:after': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:hover:not(.Mui-disabled):before': {
                    borderBottom: 'none !important',
                  },
                }}
              />
              <TextField
                label="Major"
                value={profile?.education?.major || ''}
                fullWidth
                disabled
                variant="standard"
                sx={{
                  '& .MuiInputLabel-root': {
                    fontSize: '1.5rem',
                    color: 'black',
                    fontWeight: 'bold',
                    paddingBottom: '8px',
                  },
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: 'gray', // For Safari
                    color: 'gray', // For other browsers
                  },
                  '& .MuiInput-underline:before': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:after': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:hover:not(.Mui-disabled):before': {
                    borderBottom: 'none !important',
                  },
                }}
              />
              <TextField
                label="Job Title"
                value={profile?.job_title || ''}
                fullWidth
                disabled
                variant="standard"
                sx={{
                  '& .MuiInputLabel-root': {
                    fontSize: '1.5rem',
                    color: 'black',
                    fontWeight: 'bold',
                    paddingBottom: '8px',
                  },
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: 'gray', // For Safari
                    color: 'gray', // For other browsers
                  },
                  '& .MuiInput-underline:before': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:after': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:hover:not(.Mui-disabled):before': {
                    borderBottom: 'none !important',
                  },
                }}
              />
              <TextField
                label="Interests"
                value={
                  profile?.interests?.filter(i => i.trim() !== '').join(', ') || ''
                }
                fullWidth
                disabled
                variant="standard"
                sx={{
                  '& .MuiInputLabel-root': {
                    fontSize: '1.5rem',
                    color: 'black',
                    fontWeight: 'bold',
                    paddingBottom: '8px',
                  },
                  '& .MuiInputBase-input.Mui-disabled': {
                    WebkitTextFillColor: 'gray', // Safari
                    color: 'gray', // Other browsers
                  },
                  '& .MuiInput-underline:before': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:after': {
                    borderBottom: 'none !important',
                  },
                  '& .MuiInput-underline:hover:not(.Mui-disabled):before': {
                    borderBottom: 'none !important',
                  },
                }}
              />

            </Box>
          </Box>
        </Box>
      </Box>
    </Stack>
  );
}