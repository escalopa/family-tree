import React from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  IconButton,
  Avatar,
  Menu,
  MenuItem,
  Box,
  Divider,
} from '@mui/material';
import {
  AccountTree,
  People,
  SupervisorAccount,
  Leaderboard,
  AccountCircle,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { authApi } from '../../api';
import { Roles } from '../../types';

const Header: React.FC = () => {
  const navigate = useNavigate();
  const { user, setUser, hasRole, isActive } = useAuth();
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
      setUser(null);
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
    handleClose();
  };

  const handleLogoutAll = async () => {
    try {
      await authApi.logoutAll();
      setUser(null);
      navigate('/login');
    } catch (error) {
      console.error('Logout from all devices failed:', error);
    }
    handleClose();
  };

  const handleProfile = () => {
    if (user) {
      navigate(`/users/${user.user_id}`);
    }
    handleClose();
  };

  return (
    <AppBar position="static">
      <Toolbar>
        <AccountTree sx={{ mr: 2 }} />
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          Family Tree
        </Typography>

        {user && isActive && (
          <Box sx={{ display: 'flex', gap: 1 }}>
            <Button
              color="inherit"
              startIcon={<AccountTree />}
              onClick={() => navigate('/tree')}
            >
              Tree
            </Button>
            <Button
              color="inherit"
              startIcon={<Leaderboard />}
              onClick={() => navigate('/leaderboard')}
            >
              Leaderboard
            </Button>
            {hasRole(Roles.ADMIN) && (
              <Button
                color="inherit"
                startIcon={<People />}
                onClick={() => navigate('/members')}
              >
                Members
              </Button>
            )}
            {hasRole(Roles.SUPER_ADMIN) && (
              <Button
                color="inherit"
                startIcon={<SupervisorAccount />}
                onClick={() => navigate('/users')}
              >
                Users
              </Button>
            )}
          </Box>
        )}

        {user ? (
          <>
            <IconButton
              size="large"
              aria-label="account of current user"
              aria-controls="menu-appbar"
              aria-haspopup="true"
              onClick={handleMenu}
              color="inherit"
              sx={{ ml: 2 }}
            >
              {user.avatar ? (
                <Avatar src={user.avatar} alt={user.full_name} />
              ) : (
                <AccountCircle />
              )}
            </IconButton>
            <Menu
              id="menu-appbar"
              anchorEl={anchorEl}
              anchorOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              keepMounted
              transformOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              open={Boolean(anchorEl)}
              onClose={handleClose}
            >
              <MenuItem onClick={handleProfile}>Profile</MenuItem>
              <Divider />
              <MenuItem onClick={handleLogout}>Logout</MenuItem>
              <MenuItem onClick={handleLogoutAll} sx={{ color: 'error.main' }}>
                Logout from all devices
              </MenuItem>
            </Menu>
          </>
        ) : (
          <Button color="inherit" onClick={() => navigate('/login')}>
            Login
          </Button>
        )}
      </Toolbar>
    </AppBar>
  );
};

export default Header;
