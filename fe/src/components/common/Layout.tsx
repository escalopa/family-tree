import React from 'react';
import { Box, AppBar, Toolbar, Typography, Button, Avatar, Menu, MenuItem } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/authStore';
import { authApi } from '../../api/auth';
import { ROLE_LABELS } from '../../utils/constants';

interface LayoutProps {
  children: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const { user, logout } = useAuthStore();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

  const handleMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = async () => {
    try {
      await authApi.logout();
      logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1, cursor: 'pointer' }} onClick={() => navigate('/')}>
            Family Tree
          </Typography>

          {user && (
            <>
              <Button color="inherit" onClick={() => navigate('/tree')}>
                Tree
              </Button>
              {user.role_id >= 300 && (
                <Button color="inherit" onClick={() => navigate('/members')}>
                  Members
                </Button>
              )}
              {user.role_id >= 400 && (
                <Button color="inherit" onClick={() => navigate('/users')}>
                  Users
                </Button>
              )}
              <Button color="inherit" onClick={() => navigate('/leaderboard')}>
                Leaderboard
              </Button>

              <Box sx={{ ml: 2 }}>
                <Avatar
                  src={user.avatar || undefined}
                  alt={user.full_name}
                  onClick={handleMenu}
                  sx={{ cursor: 'pointer' }}
                />
                <Menu
                  anchorEl={anchorEl}
                  open={Boolean(anchorEl)}
                  onClose={handleClose}
                >
                  <MenuItem disabled>
                    <Box>
                      <Typography variant="body2">{user.full_name}</Typography>
                      <Typography variant="caption" color="textSecondary">
                        {ROLE_LABELS[user.role_id as keyof typeof ROLE_LABELS]}
                      </Typography>
                    </Box>
                  </MenuItem>
                  <MenuItem onClick={() => { handleClose(); navigate(`/profile/${user.user_id}`); }}>
                    Profile
                  </MenuItem>
                  <MenuItem onClick={handleLogout}>Logout</MenuItem>
                </Menu>
              </Box>
            </>
          )}
        </Toolbar>
      </AppBar>

      <Box component="main" sx={{ flexGrow: 1, p: 3, bgcolor: 'background.default' }}>
        {children}
      </Box>
    </Box>
  );
};



