'use client'

import React, { useState, useEffect } from 'react';
import { Box, Typography, Paper, CircularProgress, Alert, Button } from '@mui/material';
import { Gloock } from 'next/font/google';
import { useRouter, useParams } from 'next/navigation';

const gloock = Gloock({ subsets: ['latin'], weight: '400' });

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
    <Box sx={{ minHeight: '100vh', bgcolor: '#56796F', p: 4 }}>
      <Box sx={{ maxWidth: 800, mx: 'auto' }}>
        <Paper sx={{ p: 4, borderRadius: 2 }}>
          <Typography
            variant="h4"
            className={gloock.className}
            sx={{ mb: 3, color: '#384F52' }}
          >
            {profile.username}'s Profile
          </Typography>

          <Box sx={{ mb: 3 }}>
            <Typography variant="h6" sx={{ mb: 1, color: '#384F52', fontWeight: 'bold' }}>
              Contact Information
            </Typography>
            {profile.contact?.email && (
              <Typography sx={{ mb: 1, color: '#666' }}>
                Email: {profile.contact.email}
              </Typography>
            )}
            {profile.contact?.phone && (
              <Typography sx={{ mb: 1, color: '#666' }}>
                Phone: {profile.contact.phone}
              </Typography>
            )}
          </Box>

          {profile.job_title && (
            <Box sx={{ mb: 3 }}>
              <Typography variant="h6" sx={{ mb: 1, color: '#384F52', fontWeight: 'bold' }}>
                Job Title
              </Typography>
              <Typography sx={{ color: '#666' }}>{profile.job_title}</Typography>
            </Box>
          )}

          {profile.education && (profile.education.university || profile.education.major) && (
            <Box sx={{ mb: 3 }}>
              <Typography variant="h6" sx={{ mb: 1, color: '#384F52', fontWeight: 'bold' }}>
                Education
              </Typography>
              {profile.education.university && (
                <Typography sx={{ mb: 1, color: '#666' }}>
                  University: {profile.education.university}
                </Typography>
              )}
              {profile.education.major && (
                <Typography sx={{ color: '#666' }}>
                  Major: {profile.education.major}
                </Typography>
              )}
            </Box>
          )}

          {profile.interests && profile.interests.length > 0 && (
            <Box sx={{ mb: 3 }}>
              <Typography variant="h6" sx={{ mb: 1, color: '#384F52', fontWeight: 'bold' }}>
                Interests
              </Typography>
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                {profile.interests.map((interest, index) => (
                  <Box
                    key={index}
                    sx={{
                      px: 2,
                      py: 1,
                      bgcolor: '#87A98D',
                      color: '#fff',
                      borderRadius: 2,
                      fontSize: '0.9rem',
                    }}
                  >
                    {interest}
                  </Box>
                ))}
              </Box>
            </Box>
          )}

          <Button
            onClick={() => router.back()}
            variant="contained"
            sx={{ mt: 2, bgcolor: '#87A98D', '&:hover': { bgcolor: '#7d9b87' } }}
          >
            Go Back
          </Button>
        </Paper>
      </Box>
    </Box>
  );
}

