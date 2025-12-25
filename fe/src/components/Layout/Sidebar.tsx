import React from 'react';
import {
  Box,
  Drawer,
  List,
  ListItem,
  IconButton,
  Avatar,
  Typography,
  Divider,
  useTheme,
  useMediaQuery,
  Tooltip,
} from '@mui/material';
import {
  AccountTree,
  People,
  SupervisorAccount,
  Leaderboard,
  AccountCircle,
  Menu as MenuIcon,
  ChevronLeft,
  ChevronRight,
  LightMode,
  DarkMode,
} from '@mui/icons-material';
import { useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../../contexts/AuthContext';
import { useTheme as useCustomTheme } from '../../contexts/ThemeContext';
import { Roles } from '../../types';

const DRAWER_WIDTH = 260;
const COLLAPSED_WIDTH = 72;

interface SidebarProps {
  open: boolean;
  onToggle: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ open, onToggle }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();
  const { t, i18n } = useTranslation();
  const isRTL = i18n.dir() === 'rtl';
  const { mode, toggleTheme } = useCustomTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { user, hasRole, isActive } = useAuth();

  const menuItems = [
    {
      text: t('navigation.tree'),
      icon: <AccountTree />,
      path: '/tree',
      show: user && isActive,
    },
    {
      text: t('navigation.leaderboard'),
      icon: <Leaderboard />,
      path: '/leaderboard',
      show: user && isActive,
    },
    {
      text: t('navigation.members'),
      icon: <People />,
      path: '/members',
      show: user && isActive && hasRole(Roles.ADMIN),
    },
    {
      text: t('navigation.users'),
      icon: <SupervisorAccount />,
      path: '/users',
      show: user && isActive && hasRole(Roles.SUPER_ADMIN),
    },
  ];

  const handleNavigation = (path: string) => {
    navigate(path);
    if (isMobile) {
      onToggle();
    }
  };

  const drawerContent = (
    <Box
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        bgcolor: 'background.paper',
      }}
    >
      {/* Header */}
      <Box
        sx={{
          p: 2,
          display: 'flex',
          alignItems: 'center',
          justifyContent: open || isMobile ? 'space-between' : 'center',
          minHeight: 64,
        }}
      >
        {(open || isMobile) && (
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <AccountTree color="primary" />
            <Typography variant="h6" fontWeight="bold" noWrap>
              {t('navigation.familyTree')}
            </Typography>
          </Box>
        )}
        <IconButton onClick={onToggle} size="small">
          {open ? (isRTL ? <ChevronRight /> : <ChevronLeft />) : <MenuIcon />}
        </IconButton>
      </Box>

      <Divider />

      {/* User Profile Section */}
      {user && (
        <>
          <Box sx={{ p: 2, display: 'flex', justifyContent: 'center' }}>
            <Tooltip
              title={!open && !isMobile ? user.full_name : ''}
              placement="right"
            >
              <IconButton
                onClick={() => handleNavigation('/profile')}
                sx={{
                  p: 0,
                  '&:hover': {
                    bgcolor: 'transparent',
                  },
                }}
              >
                <Box
                  sx={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 2,
                    p: 1.5,
                    borderRadius: open || isMobile ? 2 : '50%',
                    bgcolor:
                      location.pathname === '/profile'
                        ? 'primary.main'
                        : 'transparent',
                    color:
                      location.pathname === '/profile'
                        ? 'primary.contrastText'
                        : 'inherit',
                    '&:hover': {
                      bgcolor:
                        location.pathname === '/profile'
                          ? 'primary.dark'
                          : 'action.hover',
                    },
                    transition: 'all 0.2s ease-in-out',
                    width: open || isMobile ? '100%' : 'auto',
                  }}
                >
                  {user.avatar ? (
                    <Avatar
                      src={user.avatar}
                      alt={user.full_name}
                      sx={{ width: 36, height: 36 }}
                    />
                  ) : (
                    <AccountCircle sx={{ fontSize: 36 }} />
                  )}
                  {(open || isMobile) && (
                    <Box sx={{ textAlign: 'left' }}>
                      <Typography
                        variant="body2"
                        fontWeight={600}
                        noWrap
                        sx={{ color: 'inherit' }}
                      >
                        {user.full_name}
                      </Typography>
                      <Typography
                        variant="caption"
                        noWrap
                        sx={{
                          color: location.pathname === '/profile'
                            ? 'inherit'
                            : 'text.secondary',
                          opacity: 0.8
                        }}
                      >
                        {user.email}
                      </Typography>
                    </Box>
                  )}
                </Box>
              </IconButton>
            </Tooltip>
          </Box>
          <Divider />
        </>
      )}

      {/* Navigation Menu */}
      <List sx={{ flexGrow: 1, px: 2, py: 1 }}>
        {menuItems
          .filter((item) => item.show)
          .map((item) => (
            <Tooltip
              key={item.path}
              title={!open && !isMobile ? item.text : ''}
              placement="right"
            >
                <ListItem
                  disablePadding
                  sx={{
                    mb: 0.5,
                    display: 'flex',
                    justifyContent: 'center',
                  }}
                >
                  <IconButton
                    onClick={() => handleNavigation(item.path)}
                    sx={{
                      p: 0,
                      width: '100%',
                      '&:hover': {
                        bgcolor: 'transparent',
                      },
                    }}
                  >
                    <Box
                      sx={{
                        borderRadius: open || isMobile ? 2 : '50%',
                        minHeight: 48,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: open || isMobile ? 'flex-start' : 'center',
                        px: open || isMobile ? 2.5 : 1.5,
                        py: open || isMobile ? 0 : 1.5,
                        width: open || isMobile ? '100%' : 'auto',
                        bgcolor:
                          location.pathname === item.path
                            ? 'primary.main'
                            : 'transparent',
                        color:
                          location.pathname === item.path
                            ? 'primary.contrastText'
                            : 'inherit',
                        '&:hover': {
                          bgcolor:
                            location.pathname === item.path
                              ? 'primary.dark'
                              : 'action.hover',
                          transform: open || isMobile ? 'translateX(4px)' : 'none',
                        },
                        transition: 'all 0.2s ease-in-out',
                      }}
                    >
                      <Box
                        sx={{
                          display: 'flex',
                          justifyContent: 'center',
                          alignItems: 'center',
                          color: 'inherit',
                        }}
                      >
                        {item.icon}
                      </Box>
                      {(open || isMobile) && (
                        <Typography
                          variant="body2"
                          fontWeight={500}
                          sx={{ ml: 2, color: 'inherit' }}
                        >
                          {item.text}
                        </Typography>
                      )}
                    </Box>
                  </IconButton>
                </ListItem>
              </Tooltip>
          ))}
      </List>

      <Divider />

      {/* Theme Toggle */}
      <Box sx={{ p: 2, display: 'flex', justifyContent: 'center' }}>
        <Tooltip
          title={!open && !isMobile ? 'Toggle Theme' : ''}
          placement="right"
        >
          <IconButton
            onClick={toggleTheme}
            sx={{
              p: 0,
              '&:hover': {
                bgcolor: 'transparent',
              },
            }}
          >
            <Box
              sx={{
                borderRadius: open || isMobile ? 2 : '50%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: open || isMobile ? 'flex-start' : 'center',
                px: open || isMobile ? 2.5 : 1.5,
                py: open || isMobile ? 1.5 : 1.5,
                width: open || isMobile ? '100%' : 'auto',
                '&:hover': {
                  bgcolor: 'action.hover',
                },
                transition: 'all 0.2s ease-in-out',
              }}
            >
              <Box
                sx={{
                  display: 'flex',
                  justifyContent: 'center',
                  alignItems: 'center',
                }}
              >
                {mode === 'dark' ? <LightMode /> : <DarkMode />}
              </Box>
              {(open || isMobile) && (
                <Typography
                  variant="body2"
                  fontWeight={500}
                  sx={{ ml: 2 }}
                >
                  {mode === 'dark' ? t('theme.lightMode') : t('theme.darkMode')}
                </Typography>
              )}
            </Box>
          </IconButton>
        </Tooltip>
      </Box>
    </Box>
  );

  return (
    <>
      {/* Desktop Drawer */}
      {!isMobile && (
        <Drawer
          variant="permanent"
          anchor={isRTL ? 'right' : 'left'}
          sx={{
            width: open ? DRAWER_WIDTH : COLLAPSED_WIDTH,
            flexShrink: 0,
            '& .MuiDrawer-paper': {
              width: open ? DRAWER_WIDTH : COLLAPSED_WIDTH,
              boxSizing: 'border-box',
              transition: theme.transitions.create('width', {
                easing: theme.transitions.easing.sharp,
                duration: theme.transitions.duration.enteringScreen,
              }),
              overflowX: 'hidden',
            },
          }}
        >
          {drawerContent}
        </Drawer>
      )}

      {/* Mobile Drawer */}
      {isMobile && (
        <Drawer
          variant="temporary"
          anchor={isRTL ? 'right' : 'left'}
          open={open}
          onClose={onToggle}
          ModalProps={{
            keepMounted: true, // Better mobile performance
          }}
          sx={{
            '& .MuiDrawer-paper': {
              width: DRAWER_WIDTH,
              boxSizing: 'border-box',
            },
          }}
        >
          {drawerContent}
        </Drawer>
      )}
    </>
  );
};

export default Sidebar;
