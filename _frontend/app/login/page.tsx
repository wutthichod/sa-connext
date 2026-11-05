'use client'

import React, { useState, useEffect } from 'react';
import { Box, Button, Checkbox, FormControlLabel, TextField, Typography } from '@mui/material';
import { Gloock } from 'next/font/google';
import Image from 'next/image';
import { useRouter } from 'next/navigation';

const gloock = Gloock({ subsets: ['latin'], weight: '400' });

const LoginPage: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [keepLoggedIn, setKeepLoggedIn] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  // Allow access to login page even when logged in (for logout/login as different user)
  // Removed auto-redirect to allow users to logout

  const validate = () => {
    if (!email.trim() || !password.trim()) {
      setError('Please provide both email and password.');
      return false;
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email.trim())) {
      setError('Please provide a valid email address.');
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!validate()) return;
    setLoading(true);

    try {
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: email.trim(), password }),
      });

      if (!res.ok) {
        let message = `Login failed (status ${res.status})`;
        try {
          const ct = res.headers.get('content-type') || '';
          if (ct.includes('application/json')) {
            const payload = await res.json();
            message = payload?.message || payload?.error || message;
          } else {
            const text = await res.text();
            message = text || message;
          }
        } catch {}
        throw new Error(message);
      }

      const contentType = res.headers.get('content-type') || '';
      if (!contentType.includes('application/json')) {
        throw new Error('Server returned invalid response format');
      }

      // <-- Important: your API returns jwtToken, so read that key:
      const data = await res.json();
      const token = data?.jwtToken; // <---- use jwtToken
      if (!token) throw new Error('No token received from server');

      // Save token as 'token' so /join reads the right key
      const storage = keepLoggedIn ? window.localStorage : window.sessionStorage;
      storage.setItem('token', token);

      // Keep quick flag for client checks (optional)
      localStorage.setItem('isLoggedIn', 'true');

      // Dispatch custom event to trigger WebSocket connection
      window.dispatchEvent(new Event('token-set'));

      // redirect to /join
      router.replace('/join');
    } catch (err: any) {
      setError(err?.message || 'Unknown error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box className={gloock.className} display="flex" height="100vh" width="100vw" overflow="hidden" sx={{ margin: 0, padding: 0 }}>
      {/* Left */}
      <Box flex={1} bgcolor="#384F52" color="white" display="flex" flexDirection="row" justifyContent="center" alignItems="center" p={2} minHeight={0} boxSizing="border-box" minWidth="60%">
        <Image src="/images/secondary_logo.png" alt="Logo" width={250} height={250} />
        <Box display="flex" flexDirection="column" alignItems="flex-start" >
          <Typography sx={{ mb: -2, fontFamily: 'Gloock, sans-serif', fontSize: '3rem', fontWeight: "normal"}}>Connect with</Typography>
          <Typography sx={{ mb: 0, fontFamily: 'Gloock, sans-serif', fontSize: '3rem' }}>The Next Generation</Typography>
          <Typography sx={{ fontFamily: 'Gloock, sans-serif', fontSize: '1.5rem' }}>Find the people</Typography>
        </Box>
      </Box>

      {/* Right */}
      <Box flex={1} display="flex" flexDirection="column" justifyContent="flex-start" alignItems="center" p={5} pt={15} minHeight={0} boxSizing="border-box">
        <Box width="100%" maxWidth="700px" display="flex" flexDirection="column" justifyContent="center" maxHeight="100%" overflow="auto" alignItems="center">
          <Image src="/images/primary_logo.png" alt="Logo" width={250} height={250}/>
          <Typography variant="h4" textAlign="center" sx={{ mb: 5, fontFamily: 'Gloock, sans-serif', fontSize: '3rem', mt: 5, color:"#384F52"}}>Connext</Typography>

          <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '12px', minWidth:'70%' }} aria-live="polite">
            <TextField label="Email" variant="outlined" fullWidth value={email} onChange={(e) => setEmail(e.target.value)} />
            <TextField label="Password" variant="outlined" type="password" fullWidth value={password} onChange={(e) => setPassword(e.target.value)} />
            <FormControlLabel control={<Checkbox checked={keepLoggedIn} onChange={(e)=>setKeepLoggedIn(e.target.checked)} />} label="Keep me logged in" />

            {error && <Typography role="alert" sx={{ color: '#d32f2f', backgroundColor: '#ffebee', padding: '12px', borderRadius: '4px' }}>{error}</Typography>}

            <Button type="submit" variant="contained" disabled={loading} sx={{ bgcolor: '#56796F', height:'60px', width:"70%", alignSelf:"center", color: '#FFFFFF', textTransform: 'none', mt:"70px"}}>
              {loading ? 'Logging in...' : 'Log in'}
            </Button>
          </form>

          <Typography variant="body2" textAlign="center" sx={{ mb: 5, mt: 5, color: '#384F52' }}>
            Don't have an account? <a href="/register">Sign up here</a>
          </Typography>
        </Box>
      </Box>
    </Box>
  );
};

export default LoginPage;
