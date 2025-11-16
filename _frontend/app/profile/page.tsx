'use client'

import React, { useState, useEffect } from 'react';
import { Box, Button, TextField, Typography, Paper, CircularProgress, Alert, Stack } from '@mui/material';
import { Gloock, Gantari } from 'next/font/google';
import { useRouter } from 'next/navigation';

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

export default function ProfilePage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const router = useRouter();

  // Form state
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [university, setUniversity] = useState('');
  const [major, setMajor] = useState('');
  const [jobTitle, setJobTitle] = useState('');
  const [interests, setInterests] = useState('');

  useEffect(() => {
    const token = localStorage.getItem('token') || sessionStorage.getItem('token');
    if (!token) {
      router.replace('/login');
      return;
    }

    fetchProfile(token);
  }, [router]);

  const fetchProfile = async (token: string) => {
    try {
      const res = await fetch('/api/users/me', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!res.ok) {
        throw new Error('Failed to fetch profile');
      }

      const data = await res.json();
      if (data.success && data.data) {
        const userData = data.data;
        setProfile(userData);
        setUsername(userData.username || '');
        setEmail(userData.contact?.email || '');
        setPhone(userData.contact?.phone || '');
        setUniversity(userData.education?.university || '');
        setMajor(userData.education?.major || '');
        setJobTitle(userData.job_title || '');
        setInterests(userData.interests?.join(', ') || '');
      }
    } catch (err: any) {
      setError(err.message || 'Failed to load profile');
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    setSuccess(false);

    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      if (!token) {
        throw new Error('Not authenticated');
      }

      const interestsArray = interests
        .split(',')
        .map(i => i.trim())
        .filter(i => i.length > 0);

      const updateData = {
        username,
        contact: {
          email: email || undefined,
          phone: phone || undefined,
        },
        education: {
          university: university || undefined,
          major: major || undefined,
        },
        job_title: jobTitle || undefined,
        interests: interestsArray.length > 0 ? interestsArray : undefined,
      };

      const res = await fetch('/api/users/me', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updateData),
      });

      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(errorData.error || 'Failed to update profile');
      }

      setSuccess(true);
      setEditMode(false);
      await fetchProfile(token);
      
      setTimeout(() => setSuccess(false), 3000);
    } catch (err: any) {
      setError(err.message || 'Failed to update profile');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="100vh" bgcolor="#56796F">
        <CircularProgress sx={{ color: '#fff' }} />
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
                Profile
              </Typography>
            </Box>
            {!editMode && (
              <Button
                variant="outlined"
                onClick={() => setEditMode(true)}
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
                Edit
              </Button>
            )}
          </Box>
          {editMode ? (
            <Box sx={{ px: 3, pb: 4, bgcolor: '#fff', borderRadius: 5, p: 5 }}>
              <Box display="flex" flexDirection="column" gap={3}>
                <TextField
                  label="Username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  fullWidth
                  required
                />
                <TextField
                  label="Email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  fullWidth
                />
                <TextField
                  label="Phone"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  fullWidth
                />
                <TextField
                  label="University"
                  value={university}
                  onChange={(e) => setUniversity(e.target.value)}
                  fullWidth
                />
                <TextField
                  label="Major"
                  value={major}
                  onChange={(e) => setMajor(e.target.value)}
                  fullWidth
                />
                <TextField
                  label="Job Title"
                  value={jobTitle}
                  onChange={(e) => setJobTitle(e.target.value)}
                  fullWidth
                />
                <TextField
                  label="Interests (comma-separated)"
                  value={interests}
                  onChange={(e) => setInterests(e.target.value)}
                  fullWidth
                  placeholder="coding, music, travel"
                />
                <Box display="flex" gap={2}>
                  <Button
                    variant="contained"
                    onClick={handleSave}
                    disabled={saving}
                    sx={{
                      bgcolor: '#56796F',
                      color: '#fff',
                      '&:hover': { bgcolor: '#7d9b87' },
                      textTransform: 'none',
                    }}
                  >
                    {saving ? 'Saving...' : 'Save'}
                  </Button>
                  <Button
                    variant="outlined"
                    onClick={() => {
                      setEditMode(false);
                      setError(null);
                      // Reset form to original values
                      if (profile) {
                        setUsername(profile.username || '');
                        setEmail(profile.contact?.email || '');
                        setPhone(profile.contact?.phone || '');
                        setUniversity(profile.education?.university || '');
                        setMajor(profile.education?.major || '');
                        setJobTitle(profile.job_title || '');
                        setInterests(profile.interests?.join(', ') || '');
                      }
                    }}
                    sx={{
                      textTransform: 'none',
                    }}
                  >
                    Cancel
                  </Button>
                </Box>
              </Box>
            </Box>
          ) : (
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
          )}
        </Box>
      </Box>
    </Stack>
  );
}
