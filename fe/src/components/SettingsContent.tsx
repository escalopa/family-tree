import React, { useEffect } from 'react';
import {
  Box,
  Tabs,
  Tab,
  Card,
  CardContent,
  Typography,
  Stack,
  ToggleButtonGroup,
  ToggleButton,
} from '@mui/material';
import {
  Language as LanguageIcon,
  AdminPanelSettings,
  Palette,
  LightMode,
  DarkMode,
} from '@mui/icons-material';
import { useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import LanguageSettings from './LanguageSettings';
import LanguageManagement from './LanguageManagement';
import UILanguageSelector from './UILanguageSelector';
import { useAuth } from '../contexts/AuthContext';
import { useLanguage } from '../contexts/LanguageContext';
import { useTheme } from '../contexts/ThemeContext';
import { Roles } from '../types';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`settings-tabpanel-${index}`}
      aria-labelledby={`settings-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ py: 2 }}>{children}</Box>}
    </div>
  );
}

const SettingsContent: React.FC = () => {
  const { t } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const { hasRole } = useAuth();
  const { loadLanguages, loadPreferences } = useLanguage();
  const { mode, setThemeMode } = useTheme();
  const isSuperAdmin = hasRole(Roles.SUPER_ADMIN);

  // Get tab from query params, default to 0
  const tabParam = searchParams.get('tab');
  const tabValue = tabParam ? parseInt(tabParam, 10) : 0;

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setSearchParams({ tab: newValue.toString() });
  };

  // Reload data when tab changes
  useEffect(() => {
    if (tabValue === 0) {
      // Appearance tab
    } else if (tabValue === 1) {
      // Language Preferences tab - reload languages and preferences
      loadLanguages();
      loadPreferences();
    } else if (tabValue === 2 && isSuperAdmin) {
      // Language Management tab - reload languages
      loadLanguages();
    }
  }, [tabValue, isSuperAdmin]);

  const handleLanguagePreferenceSaved = () => {
    // Reload languages in case a new language was activated
    loadLanguages();
    loadPreferences();
  };

  const handleThemeChange = (_event: React.MouseEvent<HTMLElement>, newMode: string | null) => {
    if (newMode) {
      setThemeMode(newMode as 'light' | 'dark');
    }
  };

  return (
    <Box>
      <Tabs
        value={tabValue}
        onChange={handleTabChange}
        aria-label="settings tabs"
        sx={{ borderBottom: 1, borderColor: 'divider' }}
      >
        <Tab
          icon={<Palette />}
          iconPosition="start"
          label={t('settings.appearance')}
          id="settings-tab-0"
          aria-controls="settings-tabpanel-0"
        />
        <Tab
          icon={<LanguageIcon />}
          iconPosition="start"
          label={t('settings.languagePreferences')}
          id="settings-tab-1"
          aria-controls="settings-tabpanel-1"
        />
        {isSuperAdmin && (
          <Tab
            icon={<AdminPanelSettings />}
            iconPosition="start"
            label={t('settings.languageManagement')}
            id="settings-tab-2"
            aria-controls="settings-tabpanel-2"
          />
        )}
      </Tabs>

      <TabPanel value={tabValue} index={0}>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                {t('settings.theme')}
              </Typography>
              <Typography variant="body2" color="text.secondary" paragraph>
                {t('settings.themeDescription')}
              </Typography>
              <ToggleButtonGroup
                value={mode}
                exclusive
                onChange={handleThemeChange}
                aria-label="theme mode"
                fullWidth
                sx={{ mt: 2 }}
              >
                <ToggleButton value="light" aria-label="light mode">
                  <LightMode sx={{ mr: 1 }} />
                  {t('settings.lightMode')}
                </ToggleButton>
                <ToggleButton value="dark" aria-label="dark mode">
                  <DarkMode sx={{ mr: 1 }} />
                  {t('settings.darkMode')}
                </ToggleButton>
              </ToggleButtonGroup>
              <Typography variant="caption" color="text.secondary" sx={{ mt: 2, display: 'block' }}>
                {t('settings.themeLocalNote')}
              </Typography>
            </CardContent>
          </Card>
        </motion.div>
      </TabPanel>

      <TabPanel value={tabValue} index={1}>
        <Stack spacing={3}>
          {/* UI Language Selector */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3 }}
          >
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  {t('language.uiLanguage')}
                </Typography>
                <Typography variant="body2" color="text.secondary" paragraph>
                  {t('language.uiLanguageDescription')}
                </Typography>
                <UILanguageSelector />
              </CardContent>
            </Card>
          </motion.div>

          {/* Names Language Selector */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.3, delay: 0.1 }}
          >
            <LanguageSettings onSave={handleLanguagePreferenceSaved} />
          </motion.div>
        </Stack>
      </TabPanel>

      {isSuperAdmin && (
        <TabPanel value={tabValue} index={2}>
          <LanguageManagement />
        </TabPanel>
      )}
    </Box>
  );
};

export default SettingsContent;
