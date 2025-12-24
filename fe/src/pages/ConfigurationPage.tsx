import React from 'react';
import {
  Container,
  Typography,
  Box,
  Paper,
  Tabs,
  Tab,
  Divider,
} from '@mui/material';
import { Settings, Language as LanguageIcon, AdminPanelSettings } from '@mui/icons-material';
import Layout from '../components/Layout/Layout';
import LanguageSettings from '../components/LanguageSettings';
import LanguageManagement from '../components/LanguageManagement';
import { useAuth } from '../contexts/AuthContext';
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
      id={`config-tabpanel-${index}`}
      aria-labelledby={`config-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

const ConfigurationPage: React.FC = () => {
  const [tabValue, setTabValue] = React.useState(0);
  const { hasRole } = useAuth();
  const isSuperAdmin = hasRole(Roles.SUPER_ADMIN);

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  return (
    <Layout>
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ mb: 4 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
            <Settings color="primary" fontSize="large" />
            <Typography variant="h4" component="h1">
              Configuration
            </Typography>
          </Box>
          <Typography variant="body1" color="text.secondary">
            Manage your application settings and preferences
          </Typography>
        </Box>

        <Paper sx={{ width: '100%' }}>
          <Tabs
            value={tabValue}
            onChange={handleTabChange}
            aria-label="configuration tabs"
            sx={{ borderBottom: 1, borderColor: 'divider' }}
          >
            <Tab
              icon={<LanguageIcon />}
              iconPosition="start"
              label="Language Preferences"
              id="config-tab-0"
              aria-controls="config-tabpanel-0"
            />
            {isSuperAdmin && (
              <Tab
                icon={<AdminPanelSettings />}
                iconPosition="start"
                label="Language Management"
                id="config-tab-1"
                aria-controls="config-tabpanel-1"
              />
            )}
            {/* Future tabs can be added here */}
            {/* <Tab icon={<NotificationsIcon />} iconPosition="start" label="Notifications" /> */}
            {/* <Tab icon={<SecurityIcon />} iconPosition="start" label="Security" /> */}
          </Tabs>

          <TabPanel value={tabValue} index={0}>
            <Typography variant="h6" gutterBottom>
              Language Preferences
            </Typography>
            <Typography variant="body2" color="text.secondary" paragraph>
              Configure how member names are displayed throughout the application.
            </Typography>
            <Divider sx={{ my: 2 }} />
            <LanguageSettings />
          </TabPanel>

          {isSuperAdmin && (
            <TabPanel value={tabValue} index={1}>
              <Typography variant="h6" gutterBottom>
                Language Management
              </Typography>
              <Typography variant="body2" color="text.secondary" paragraph>
                Add, edit, or deactivate languages available in the system.
              </Typography>
              <Divider sx={{ my: 2 }} />
              <LanguageManagement />
            </TabPanel>
          )}

          {/* Future tab panels */}
          {/* <TabPanel value={tabValue} index={2}>
            Notification settings content
          </TabPanel> */}
        </Paper>
      </Container>
    </Layout>
  );
};

export default ConfigurationPage;
