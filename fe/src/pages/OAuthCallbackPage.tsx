import { useEffect, useState } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { Box, Typography, CircularProgress, Alert } from '@mui/material';
import { authApi, OAuthProvider } from '../api/auth';
import { useAuthStore } from '../store/authStore';

export const OAuthCallbackPage = () => {
  const navigate = useNavigate();
  const { provider } = useParams<{ provider: string }>();
  const [searchParams] = useSearchParams();
  const [error, setError] = useState<string | null>(null);
  const setUser = useAuthStore((state) => state.setUser);

  useEffect(() => {
    const handleCallback = async () => {
      try {
        // Get the authorization code from URL
        const code = searchParams.get('code');
        const errorParam = searchParams.get('error');

        // Check for OAuth error
        if (errorParam) {
          setError(`Authentication failed: ${errorParam}`);
          setTimeout(() => navigate('/login'), 3000);
          return;
        }

        // Validate provider
        if (!provider) {
          setError('Invalid OAuth provider');
          setTimeout(() => navigate('/login'), 3000);
          return;
        }

        // Validate code
        if (!code) {
          setError('Authorization code not found');
          setTimeout(() => navigate('/login'), 3000);
          return;
        }

        // Send code to backend
        const response = await authApi.handleCallback(provider as OAuthProvider, code);

        // Update auth store
        setUser(response.user);

        // Redirect to dashboard on success
        navigate('/', { replace: true });
      } catch (err: any) {
        console.error('OAuth callback error:', err);
        const errorMessage = err.response?.data?.error || 'Authentication failed. Please try again.';
        setError(errorMessage);
        setTimeout(() => navigate('/login'), 3000);
      }
    };

    handleCallback();
  }, [provider, searchParams, navigate, setUser]);

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '100vh',
        gap: 2,
        px: 2,
      }}
    >
      {error ? (
        <>
          <Alert severity="error" sx={{ maxWidth: 500 }}>
            {error}
          </Alert>
          <Typography color="text.secondary">
            Redirecting to login...
          </Typography>
        </>
      ) : (
        <>
          <CircularProgress size={60} />
          <Typography variant="h6" color="text.secondary">
            Completing authentication...
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Please wait while we verify your credentials
          </Typography>
        </>
      )}
    </Box>
  );
};
