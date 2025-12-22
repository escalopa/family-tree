import React, { useEffect, useState, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Box, CircularProgress, Typography } from '@mui/material';
import { authApi } from '../api';
import { useAuth } from '../contexts/AuthContext';

const CallbackPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { setUser } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const hasCalledCallback = useRef(false);

  useEffect(() => {
    if (hasCalledCallback.current) {
      return;
    }

    const handleCallback = async () => {
      const code = searchParams.get('code');
      const state = searchParams.get('state');
      const provider = window.location.pathname.split('/')[2];

      if (!code || !state) {
        setError('Invalid callback parameters');
        return;
      }

      hasCalledCallback.current = true;

      try {
        const response = await authApi.handleCallback(provider, code, state);
        setUser(response.user);

        if (response.user.is_active) {
          navigate('/tree');
        } else {
          navigate('/inactive');
        }
      } catch (err) {
        console.error('Auth callback failed:', err);
        setError('Authentication failed');
        setTimeout(() => navigate('/login'), 2000);
      }
    };

    handleCallback();
  }, [searchParams, navigate, setUser]);

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      {error ? (
        <Typography color="error">{error}</Typography>
      ) : (
        <>
          <CircularProgress />
          <Typography sx={{ mt: 2 }}>Signing you in...</Typography>
        </>
      )}
    </Box>
  );
};

export default CallbackPage;
