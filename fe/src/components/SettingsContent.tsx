import React, { useEffect } from 'react';
import {
  Box,
  Tabs,
  Tab,
} from '@mui/material';
import {
  Language as LanguageIcon,
  Translate,
} from '@mui/icons-material';
import { useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { motion } from 'framer-motion';
import LanguageSettings from './LanguageSettings';
import LanguageManagement from './LanguageManagement';
import { useAuth } from '../contexts/AuthContext';
import { useLanguage } from '../contexts/LanguageContext';
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
      // Language Preferences tab - reload languages and preferences
      loadLanguages();
      loadPreferences();
    } else if (tabValue === 1 && isSuperAdmin) {
      // Language Management tab - reload languages
      loadLanguages();
    }
  }, [tabValue, isSuperAdmin]);

  const handleLanguagePreferenceSaved = () => {
    // Reload languages in case a new language was activated
    loadLanguages();
    loadPreferences();
  };

  return (
    <Box>
      <Tabs
        value={tabValue}
        onChange={handleTabChange}
        aria-label="settings tabs"
        sx={{
          borderBottom: 1,
          borderColor: 'divider',
          '& .MuiTab-root': {
            '& .MuiSvgIcon-root': {
              marginInlineEnd: 1,
            }
          }
        }}
      >
        <Tab
          icon={<LanguageIcon />}
          iconPosition="start"
          label={t('settings.languagePreferences')}
          id="settings-tab-0"
          aria-controls="settings-tabpanel-0"
        />
        {isSuperAdmin && (
          <Tab
            icon={<Translate />}
            iconPosition="start"
            label={t('settings.languageManagement')}
            id="settings-tab-1"
            aria-controls="settings-tabpanel-1"
          />
        )}
      </Tabs>

      <TabPanel value={tabValue} index={0}>
        {/* Names Language Selector */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          <LanguageSettings onSave={handleLanguagePreferenceSaved} />
        </motion.div>
      </TabPanel>

      {isSuperAdmin && (
        <TabPanel value={tabValue} index={1}>
          <LanguageManagement />
        </TabPanel>
      )}
    </Box>
  );
};

export default SettingsContent;
