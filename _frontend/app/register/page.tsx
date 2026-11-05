'use client'

import React, { useState } from 'react';
import { Box, Button, TextField, Typography } from '@mui/material';
import { Gloock } from 'next/font/google';
import Image from 'next/image';
import { useRouter } from 'next/navigation';

const gloock = Gloock({ subsets: ['latin'], weight: '400' });

const RegisterPage = () => {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [university, setUniversity] = useState('');
  const [major, setMajor] = useState('');
  const [jobTitle, setJobTitle] = useState('');
  const [interests, setInterests] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    // Validation
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      setLoading(false);
      return;
    }

    if (!username || !email || !password) {
      setError('Username, email, and password are required');
      setLoading(false);
      return;
    }

    try {
      // Parse interests from comma-separated string to array
      const interestsArray = interests
        .split(',')
        .map(i => i.trim())
        .filter(i => i.length > 0);

      const requestBody = {
        username,
        password,
        contact: {
          email,
          phone: phone || undefined,
        },
        education: {
          university: university || undefined,
          major: major || undefined,
        },
        jobTitle: jobTitle || undefined,
        interests: interestsArray.length > 0 ? interestsArray : undefined,
      };

      const res = await fetch('/api/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      });

      if (!res.ok) {
        let errorMessage = `Registration failed (status ${res.status})`;
        
        try {
          const contentType = res.headers.get('content-type');
          if (contentType && contentType.includes('application/json')) {
            const errorData = await res.json();
            errorMessage = errorData?.message || errorData?.error || errorMessage;
          } else {
            const textError = await res.text();
            errorMessage = textError || errorMessage;
          }
        } catch (parseError) {
          // Silently handle parse errors
        }
        
        throw new Error(errorMessage);
      }

      const contentType = res.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        throw new Error('Server returned invalid response format');
      }

      const data = await res.json();

      // Registration successful
      setSuccess(true);
      
      // Redirect to login after 2 seconds
      setTimeout(() => {
        router.push('/login');
      }, 2000);

    } catch (err: any) {
      setError(err?.message || 'Unknown error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      className={gloock.className} 
      display="flex"
      height="100vh"
      width="100vw"
      overflow="hidden"
      sx={{ margin: 0, padding: 0 }}
    >
      {/* Left side */}
      <Box
        flex={1}
        bgcolor="#384F52"
        color="white"
        display="flex"
        flexDirection="row"
        justifyContent="center"
        alignItems="center"
        p={2}
        minHeight={0}
        boxSizing="border-box"
        minWidth="60%"
      >
        <Image src="/images/secondary_logo.png" alt="Logo" width={250} height={250} />
        <Box display="flex" flexDirection="column" alignItems="flex-start" ml={4}>
          <Typography sx={{ mb: -2, fontFamily: 'Gloock, sans-serif', fontSize: '3rem', fontWeight: "normal"}}>
            Join
          </Typography>
          <Typography sx={{ mb: 0, fontFamily: 'Gloock, sans-serif', fontSize: '3rem' }}>
            Connext Today
          </Typography>
          <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1.5rem' }}>
            Connect with the next generation
          </Typography>
        </Box>
      </Box>

      {/* Right side */}
      <Box
        flex={1}
        display="flex"
        flexDirection="column"
        justifyContent="flex-start"
        alignItems="center"
        p={3}
        minHeight={0}
        boxSizing="border-box"
      >
        <Box
          width="100%"
          maxWidth="700px"
          display="flex"
          flexDirection="column"
          justifyContent="flex-start"
          maxHeight="100%"
          overflow="auto"
          alignItems="center"
        >
          <Image src="/images/primary_logo.png" alt="Logo" width={250} height={250}/>
          <Typography variant="h4" textAlign="center" sx={{ mb: 5, fontFamily: 'Gloock, sans-serif', fontSize: '3rem', mt: 5, color:"#384F52"}}>
            Register
          </Typography>

          {success && (
            <Typography 
              sx={{ 
                color: '#2e7d32', 
                fontSize: '0.9rem',
                backgroundColor: '#e8f5e9',
                padding: '12px',
                borderRadius: '4px',
                border: '1px solid #81c784',
                textAlign: 'center',
                mb: 2,
                width: '70%'
              }}
            >
              Registration successful! Redirecting to login...
            </Typography>
          )}

          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '20px', minWidth:'70%' }}>

            {/* Username */}
            <Box display="flex" flexDirection="column" gap="4px">
              <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1rem', fontWeight: 500, color: '#384F52' }}>
                Username *
              </Typography>
              <TextField
                variant="outlined"
                fullWidth
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
              />
            </Box>

            {/* Email */}
            <Box display="flex" flexDirection="column" gap="4px">
              <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1rem', fontWeight: 500, color: '#384F52' }}>
                Email *
              </Typography>
              <TextField
                type="email"
                variant="outlined"
                fullWidth
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </Box>

            {/* Phone */}
            <Box display="flex" flexDirection="column" gap="4px">
              <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1rem', fontWeight: 500, color: '#384F52' }}>
                Phone
              </Typography>
              <TextField
                variant="outlined"
                fullWidth
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="+1234567890"
              />
            </Box>

            {/* Password */}
            <Box display="flex" flexDirection="column" gap="4px">
              <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1rem', fontWeight: 500, color: '#384F52' }}>
                Password *
              </Typography>
              <TextField
                variant="outlined"
                type="password"
                fullWidth
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </Box>

            {/* Confirm Password */}
            <Box display="flex" flexDirection="column" gap="4px">
              <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1rem', fontWeight: 500, color: '#384F52' }}>
                Confirm Password *
              </Typography>
              <TextField
                variant="outlined"
                type="password"
                fullWidth
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
              />
            </Box>

            {/* University */}
            <Box display="flex" flexDirection="column" gap="4px">
              <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1rem', fontWeight: 500, color: '#384F52' }}>
                University
              </Typography>
              <TextField
                variant="outlined"
                fullWidth
                value={university}
                onChange={(e) => setUniversity(e.target.value)}
              />
            </Box>

            {/* Major */}
            <Box display="flex" flexDirection="column" gap="4px">
              <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1rem', fontWeight: 500, color: '#384F52' }}>
                Major
              </Typography>
              <TextField
                variant="outlined"
                fullWidth
                value={major}
                onChange={(e) => setMajor(e.target.value)}
              />
            </Box>

            {/* Job Title */}
            <Box display="flex" flexDirection="column" gap="4px">
              <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1rem', fontWeight: 500, color: '#384F52' }}>
                Job Title
              </Typography>
              <TextField
                variant="outlined"
                fullWidth
                value={jobTitle}
                onChange={(e) => setJobTitle(e.target.value)}
              />
            </Box>

            {/* Interests */}
            <Box display="flex" flexDirection="column" gap="4px">
              <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1rem', fontWeight: 500, color: '#384F52' }}>
                Interests (comma-separated)
              </Typography>
              <TextField
                variant="outlined"
                fullWidth
                value={interests}
                onChange={(e) => setInterests(e.target.value)}
                placeholder="coding, music, travel"
              />
            </Box>

            {error && (
              <Typography 
                sx={{ 
                  color: '#d32f2f', 
                  fontSize: '0.9rem',
                  backgroundColor: '#ffebee',
                  padding: '12px',
                  borderRadius: '4px',
                  border: '1px solid #ef9a9a',
                  textAlign: 'center'
                }}
              >
                {error}
              </Typography>
            )}

            {/* Register Button */}
            <Button
              type="submit"
              variant="contained"
              disabled={loading || success}
              sx={{
                bgcolor: '#56796F',
                height:'60px',
                width:"70%",
                alignSelf:"center",
                color: '#FFFFFF',
                fontFamily: 'Gloock, sans-serif',
                fontSize: '1.5rem',
                textTransform: 'none',
                mt:"20px"
              }}
            >
              {loading ? 'Registering...' : success ? 'Success!' : 'Register'}
            </Button>
          </form>

          {/* Footer */}
          <Typography   
            variant="body2"
            textAlign="center"
            sx={{
              mb: 5,
              fontFamily: 'Gloock, sans-serif',
              fontSize: '1rem',
              mt: 5,
              color: '#384F52',
              '& a': {
                color: '#56796F',      
                textDecoration: 'none',
                transition: 'color 0.2s',
                '&:hover': {
                  color: '#87A98D',     
                },
              },
            }}
          >
            Already have an account?<a href="/login"> Sign in here</a>
          </Typography>
        </Box>
      </Box>
    </Box>
  );
};

export default RegisterPage;