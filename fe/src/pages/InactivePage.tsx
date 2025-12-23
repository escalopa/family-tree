import React, { useState, useEffect } from 'react';
import { Box, Card, CardContent, Typography, Container, Button, CircularProgress, Alert } from '@mui/material';
import { Info, Refresh } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const InactivePage: React.FC = () => {
  const navigate = useNavigate();
  const { refreshUser, isActive, user } = useAuth();
  const [checking, setChecking] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Automatically redirect if user becomes active
  useEffect(() => {
    if (isActive) {
      console.log('[InactivePage] User is now active, redirecting to home...');
      navigate('/', { replace: true });
    }
  }, [isActive, navigate]);

  const handleCheckStatus = async () => {
    setChecking(true);
    setError(null);
    try {
      // Refresh user data from backend
      await refreshUser();
      // Wait a tick for state to update
      await new Promise(resolve => setTimeout(resolve, 100));
      // Check if user is still inactive after refresh
      if (user && !user.is_active) {
        setError('Your account is still inactive. Please contact an administrator.');
      }
      // If user is active, the useEffect above will redirect
    } catch (err: any) {
      console.error('Failed to refresh user data:', err);
      setError('Failed to check account status. Please try again.');
    } finally {
      setChecking(false);
    }
  };

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
        <Card>
          <CardContent sx={{ textAlign: 'center', p: 4 }}>
            <Info color="warning" sx={{ fontSize: 60, mb: 2 }} />
            <Typography variant="h5" gutterBottom>
              Account Pending Activation
            </Typography>
            <Typography variant="body1" color="text.secondary" paragraph>
              Your account has been created successfully, but it needs to be activated by an
              administrator before you can access the system.
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Please contact an administrator to activate your account.
            </Typography>

            {error && (
              <Alert severity="warning" sx={{ mb: 2 }}>
                {error}
              </Alert>
            )}

            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 3 }}>
              <Button
                variant="contained"
                color="primary"
                startIcon={checking ? <CircularProgress size={20} color="inherit" /> : <Refresh />}
                onClick={handleCheckStatus}
                disabled={checking}
                fullWidth
              >
                {checking ? 'Checking Status...' : 'Check Account Status'}
              </Button>
            </Box>
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
};

export default InactivePage;
