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
} from '@mui/material';
import {
  AccountCircle,
  Leaderboard,
  ExitToApp,
  PowerSettingsNew,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import Layout from '../components/Layout/Layout';
import SettingsContent from '../components/SettingsContent';
import { useAuth } from '../contexts/AuthContext';
import { authApi } from '../api';
import { getRoleName } from '../utils/helpers';

const ProfilePage: React.FC = () => {
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
        <Typography>Please log in</Typography>
      </Layout>
    );
  }

  return (
    <Layout>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        {/* User Info Section */}
        <Paper sx={{ p: 4, mb: 3 }}>
          <Grid container spacing={3}>
            <Grid item xs={12} md={4} sx={{ textAlign: 'center' }}>
              <Avatar
                src={user.avatar || undefined}
                sx={{ width: 120, height: 120, mx: 'auto', mb: 2 }}
              >
                {user.full_name[0]}
              </Avatar>
            </Grid>
            <Grid item xs={12} md={8}>
              <Typography variant="h4" gutterBottom>
                {user.full_name}
              </Typography>
              <Typography variant="body1" color="text.secondary" gutterBottom>
                {user.email}
              </Typography>
              <Box sx={{ mt: 2, display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                <Chip
                  label={getRoleName(user.role_id)}
                  color="primary"
                  variant="outlined"
                />
                <Chip
                  label={user.is_active ? 'Active' : 'Inactive'}
                  color={user.is_active ? 'success' : 'default'}
                  variant="outlined"
                />
                {user.total_score !== undefined && (
                  <Chip
                    label={`${user.total_score} points`}
                    color="secondary"
                    variant="outlined"
                  />
                )}
              </Box>

              <Box sx={{ mt: 3 }}>
                <Button
                  variant="contained"
                  startIcon={<Leaderboard />}
                  onClick={() => navigate(`/users/${user.user_id}`)}
                  fullWidth
                >
                  View Progress & Scores
                </Button>
              </Box>
            </Grid>
          </Grid>
        </Paper>

        {/* Settings Section */}
        <Paper sx={{ p: 4, mb: 3 }}>
          <Typography variant="h5" gutterBottom sx={{ mb: 3 }}>
            Settings
          </Typography>
          <SettingsContent />
        </Paper>

        {/* Logout Section */}
        <Paper sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>
            Session Management
          </Typography>
          <Divider sx={{ my: 2 }} />
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6}>
              <Button
                variant="outlined"
                color="primary"
                startIcon={<ExitToApp />}
                onClick={handleLogout}
                fullWidth
              >
                Logout
              </Button>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Button
                variant="outlined"
                color="error"
                startIcon={<PowerSettingsNew />}
                onClick={handleLogoutAll}
                fullWidth
              >
                Logout from All Devices
              </Button>
            </Grid>
          </Grid>
        </Paper>
      </Container>
    </Layout>
  );
};

export default ProfilePage;
