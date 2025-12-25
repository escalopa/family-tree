import React, { useState, useEffect } from 'react';
import { Box, Card, CardContent, Typography, Container, Button, CircularProgress, Alert } from '@mui/material';
import { Info, Refresh } from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../contexts/AuthContext';

const InactivePage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const { refreshUser, isActive, user } = useAuth();
  const [checking, setChecking] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isRTL = i18n.dir() === 'rtl';

  // Automatically redirect if user becomes active
  useEffect(() => {
    if (isActive) {
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
        setError(t('inactive.accountStillInactive'));
      }
      // If user is active, the useEffect above will redirect
    } catch (err: any) {

      setError(t('inactive.failedToCheckStatus'));
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
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.4 }}
        >
        <Card>
          <CardContent sx={{ textAlign: 'center', p: 4 }}>
            <Info color="warning" sx={{ fontSize: 60, mb: 2 }} />
            <Typography variant="h5" gutterBottom>
              {t('inactive.accountPendingActivation')}
            </Typography>
            <Typography variant="body1" color="text.secondary" paragraph>
              {t('inactive.accountCreatedSuccessfully')}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              {t('inactive.pleaseContactAdmin')}
            </Typography>

            {error && (
              <Alert
                severity="warning"
                sx={{
                  mb: 2,
                  textAlign: isRTL ? 'right' : 'left',
                  '& .MuiAlert-icon': {
                    marginInlineEnd: 1.5,
                    marginInlineStart: 0,
                  }
                }}
              >
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
                {checking ? t('inactive.checkingStatus') : t('inactive.checkAccountStatus')}
              </Button>
            </Box>
          </CardContent>
        </Card>
        </motion.div>
      </Box>
    </Container>
  );
};

export default InactivePage;
