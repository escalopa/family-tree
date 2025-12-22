import React, { useEffect } from 'react';
import { Box, Button, Card, CardContent, Typography, Container } from '@mui/material';
import { Google } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { authApi } from '../api';

const LoginPage: React.FC = () => {
  const { user, loading } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    // Redirect to home if already logged in
    if (!loading && user) {
      navigate('/tree', { replace: true });
    }
  }, [user, loading, navigate]);

  const handleGoogleLogin = async () => {
    try {
      const { url } = await authApi.getAuthURL('google');
      window.location.href = url;
    } catch (error) {
      console.error('Failed to get auth URL:', error);
    }
  };

  // Show nothing while checking auth status
  if (loading) {
    return null;
  }

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Card sx={{ width: '100%' }}>
          <CardContent sx={{ textAlign: 'center', p: 4 }}>
            <Typography variant="h4" component="h1" gutterBottom>
              Family Tree
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
              Sign in to access your family tree
            </Typography>
            <Button
              variant="contained"
              size="large"
              startIcon={<Google />}
              onClick={handleGoogleLogin}
              fullWidth
              sx={{ py: 1.5 }}
            >
              Sign in with Google
            </Button>
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
};

export default LoginPage;
