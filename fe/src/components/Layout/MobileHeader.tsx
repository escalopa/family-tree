import React, { useState } from 'react';
import {
  AppBar,
  Toolbar,
  IconButton,
  Typography,
  Box,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
} from '@mui/material';
import {
  Menu as MenuIcon,
  AccountTree,
  Language as LanguageIcon,
  LightMode,
  DarkMode,
  Check,
  Logout,
  ExitToApp,
  PowerSettingsNew,
} from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { useInterfaceLanguage } from '../../contexts/InterfaceLanguageContext';
import { useTheme } from '../../contexts/ThemeContext';
import { useAuth } from '../../contexts/AuthContext';
import { authApi } from '../../api';

interface MobileHeaderProps {
  onMenuClick: () => void;
}

const MobileHeader: React.FC<MobileHeaderProps> = ({ onMenuClick }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { user, setUser } = useAuth();
  const { interfaceLanguage, supportedLanguagesWithNames, changeInterfaceLanguage } = useInterfaceLanguage();
  const { mode, toggleTheme } = useTheme();
  const [languageAnchorEl, setLanguageAnchorEl] = useState<null | HTMLElement>(null);
  const [logoutAnchorEl, setLogoutAnchorEl] = useState<null | HTMLElement>(null);

  const handleLanguageMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setLanguageAnchorEl(event.currentTarget);
  };

  const handleLanguageMenuClose = () => {
    setLanguageAnchorEl(null);
  };

  const handleLanguageChange = (languageCode: string) => {
    changeInterfaceLanguage(languageCode);
    handleLanguageMenuClose();
  };

  const handleLogoutMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setLogoutAnchorEl(event.currentTarget);
  };

  const handleLogoutMenuClose = () => {
    setLogoutAnchorEl(null);
  };

  const handleLogout = async () => {
    try {
      await authApi.logout();
      setUser(null);
      navigate('/login');
    } catch (error) {

    }
    handleLogoutMenuClose();
  };

  const handleLogoutAll = async () => {
    try {
      await authApi.logoutAll();
      setUser(null);
      navigate('/login');
    } catch (error) {

    }
    handleLogoutMenuClose();
  };

  return (
    <AppBar
      position="fixed"
      sx={{
        display: { xs: 'block', md: 'none' },
        zIndex: (theme) => theme.zIndex.drawer + 1,
      }}
    >
      <Toolbar>
        <IconButton
          color="inherit"
          aria-label="open drawer"
          edge="start"
          onClick={onMenuClick}
          sx={{ marginInlineEnd: 2 }}
        >
          <MenuIcon />
        </IconButton>
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.3 }}
          style={{ flexGrow: 1 }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <AccountTree />
            <Typography variant="h6" noWrap component="div">
              Family Tree
            </Typography>
          </Box>
        </motion.div>

        {/* Theme Toggle */}
        <IconButton
          color="inherit"
          onClick={toggleTheme}
          aria-label={t('theme.toggleTheme')}
          sx={{ marginInlineEnd: 1 }}
        >
          {mode === 'dark' ? <LightMode /> : <DarkMode />}
        </IconButton>

        {/* Language Selector */}
        <IconButton
          color="inherit"
          onClick={handleLanguageMenuOpen}
          aria-label={t('language.interfaceLanguage')}
          sx={{ marginInlineEnd: 1 }}
        >
          <LanguageIcon />
        </IconButton>

        {/* Logout Menu */}
        {user && (
          <IconButton
            color="inherit"
            onClick={handleLogoutMenuOpen}
            aria-label={t('common.logout')}
          >
            <Logout />
          </IconButton>
        )}

        {/* Language Menu */}
        <Menu
          anchorEl={languageAnchorEl}
          open={Boolean(languageAnchorEl)}
          onClose={handleLanguageMenuClose}
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'right',
          }}
          transformOrigin={{
            vertical: 'top',
            horizontal: 'right',
          }}
        >
          {supportedLanguagesWithNames.map((lang) => (
            <MenuItem
              key={lang.language_code}
              onClick={() => handleLanguageChange(lang.language_code)}
              selected={interfaceLanguage === lang.language_code}
            >
              <ListItemText>{lang.language_name}</ListItemText>
              {interfaceLanguage === lang.language_code && (
                <ListItemIcon sx={{ minWidth: 'auto', marginInlineStart: 2 }}>
                  <Check fontSize="small" />
                </ListItemIcon>
              )}
            </MenuItem>
          ))}
        </Menu>

        {/* Logout Menu */}
        <Menu
          anchorEl={logoutAnchorEl}
          open={Boolean(logoutAnchorEl)}
          onClose={handleLogoutMenuClose}
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'right',
          }}
          transformOrigin={{
            vertical: 'top',
            horizontal: 'right',
          }}
        >
          <MenuItem onClick={handleLogout}>
            <ListItemIcon>
              <ExitToApp fontSize="small" />
            </ListItemIcon>
            <ListItemText>{t('common.logout')}</ListItemText>
          </MenuItem>
          <MenuItem onClick={handleLogoutAll}>
            <ListItemIcon>
              <PowerSettingsNew fontSize="small" />
            </ListItemIcon>
            <ListItemText>{t('auth.logoutFromAllDevices')}</ListItemText>
          </MenuItem>
        </Menu>
      </Toolbar>
    </AppBar>
  );
};

export default MobileHeader;
