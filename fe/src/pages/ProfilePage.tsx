import React from 'react';
import {
  Container,
  Box,
  Paper,
  Typography,
  Avatar,
  Button,
  Divider,
  Grid,
  Chip,
  useTheme,
} from '@mui/material';
import {
  Leaderboard,
  ExitToApp,
  PowerSettingsNew,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import Layout from '../components/Layout/Layout';
import SettingsContent from '../components/SettingsContent';
import { useAuth } from '../contexts/AuthContext';
import { authApi } from '../api';
import { getRoleName } from '../utils/helpers';
import DirectionalButton from '../components/DirectionalButton';

const ProfilePage: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { user, setUser } = useAuth();

  const handleLogout = async () => {
    try {
      await authApi.logout();
      setUser(null);
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const handleLogoutAll = async () => {
    try {
      await authApi.logoutAll();
      setUser(null);
      navigate('/login');
    } catch (error) {
      console.error('Logout from all devices failed:', error);
    }
  };

  if (!user) {
    return (
      <Layout>
        <Typography>{t('common.login')}</Typography>
      </Layout>
    );
  }

  return (
    <Layout>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        {/* User Info Section */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
        >
          <Paper sx={{ p: 4, mb: 3 }}>
            <Grid container spacing={3}>
              <Grid item xs={12} md={4} sx={{ textAlign: 'center' }}>
                <motion.div
                  initial={{ scale: 0.8, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  transition={{ duration: 0.5, delay: 0.2 }}
                >
                  <Avatar
                    src={user.avatar || undefined}
                    sx={{ width: 120, height: 120, mx: 'auto', mb: 2 }}
                  >
                    {user.full_name[0]}
                  </Avatar>
                </motion.div>
              </Grid>
              <Grid item xs={12} md={8}>
                <motion.div
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ duration: 0.4, delay: 0.1 }}
                >
                  <Typography variant="h4" gutterBottom>
                    {user.full_name}
                  </Typography>
                  <Typography variant="body1" color="text.secondary" gutterBottom>
                    {user.email}
                  </Typography>
                  <Box sx={{ mt: 2, display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                <Chip
                  label={getRoleName(user.role_id, t)}
                  color="primary"
                  variant="outlined"
                />
                <Chip
                  label={user.is_active ? t('user.active') : t('user.inactive')}
                  color={user.is_active ? 'success' : 'default'}
                  variant="outlined"
                />
                {user.total_score !== undefined && (
                  <Chip
                    label={`${user.total_score} ${t('user.points')}`}
                    color="secondary"
                    variant="outlined"
                  />
                )}
                  </Box>

              <Box sx={{ mt: 3 }}>
                <DirectionalButton
                  variant="contained"
                  icon={<Leaderboard />}
                  onClick={() => navigate(`/users/${user.user_id}`)}
                  fullWidth
                >
                  {t('profile.viewProgressAndScores')}
                </DirectionalButton>
              </Box>
                </motion.div>
              </Grid>
            </Grid>
          </Paper>
        </motion.div>

        {/* Settings Section */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.2 }}
        >
          <Paper sx={{ p: 4, mb: 3 }}>
            <Typography variant="h5" gutterBottom sx={{ mb: 3 }}>
              {t('settings.title')}
            </Typography>
            <SettingsContent />
          </Paper>
        </motion.div>

        {/* Logout Section */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.3 }}
        >
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              {t('profile.sessionManagement')}
            </Typography>
            <Divider sx={{ my: 2 }} />
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <DirectionalButton
                  variant="outlined"
                  color="primary"
                  icon={<ExitToApp />}
                  onClick={handleLogout}
                  fullWidth
                >
                  {t('common.logout')}
                </DirectionalButton>
              </Grid>
              <Grid item xs={12} sm={6}>
                <DirectionalButton
                  variant="outlined"
                  color="error"
                  icon={<PowerSettingsNew />}
                  onClick={handleLogoutAll}
                  fullWidth
                >
                  {t('auth.logoutFromAllDevices')}
                </DirectionalButton>
              </Grid>
            </Grid>
          </Paper>
        </motion.div>
      </Container>
    </Layout>
  );
};

export default ProfilePage;
