import React, { useEffect, useState } from 'react';
import { Box, Button, Card, CardContent, Typography, Container, Stack, CircularProgress, SvgIcon } from '@mui/material';
import {
  Google,
  GitHub,
  Facebook,
  LinkedIn,
  Code,
  Language,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { authApi } from '../api';

// Custom SVG Icons for providers not in Material-UI
const YandexIcon = () => (
  <SvgIcon viewBox="0 0 24 24">
    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1.93 15.88h-2.12l-2.25-5.77H9.3V17.88H7.5V6.12h3.84c2.37 0 3.78 1.17 3.78 3.18 0 1.56-.84 2.58-2.19 2.97l2.49 5.61h.51zm-2.49-7.59c1.17 0 1.86-.63 1.86-1.71 0-1.08-.69-1.68-1.86-1.68h-1.8v3.39h1.8z" />
  </SvgIcon>
);

const VKIcon = () => (
  <SvgIcon viewBox="0 0 24 24">
    <path d="M15.07 2H8.93C3.33 2 2 3.33 2 8.93v6.14C2 20.67 3.33 22 8.93 22h6.14c5.6 0 6.93-1.33 6.93-6.93V8.93C22 3.33 20.67 2 15.07 2zm3.83 14.2h-1.43c-.52 0-.68-.42-1.61-1.35-.81-.79-1.17-.89-1.38-.89-.28 0-.36.08-.36.47v1.23c0 .33-.11.53-1 .53-1.47 0-3.1-.89-4.24-2.55-1.71-2.4-2.18-4.2-2.18-4.57 0-.21.08-.4.47-.4h1.43c.35 0 .48.16.62.53.68 1.97 1.82 3.69 2.29 3.69.18 0 .26-.08.26-.52v-2.03c-.06-.98-.58-1.06-.58-1.41 0-.17.14-.33.36-.33h2.24c.3 0 .4.16.4.5v2.73c0 .3.14.4.22.4.18 0 .33-.1.66-.43 1.02-1.14 1.75-2.9 1.75-2.9.1-.2.25-.4.64-.4h1.43c.43 0 .52.22.43.52-.16.75-1.88 3.19-1.88 3.19-.15.24-.18.35 0 .62.13.2.57.56.86.9.53.6.94 1.1 1.05 1.45.11.35-.08.53-.51.53z" />
  </SvgIcon>
);

const InstagramIcon = () => (
  <SvgIcon viewBox="0 0 24 24">
    <path d="M7.8 2h8.4C19.4 2 22 4.6 22 7.8v8.4c0 3.2-2.6 5.8-5.8 5.8H7.8C4.6 22 2 19.4 2 16.2V7.8C2 4.6 4.6 2 7.8 2m-.2 2C5.6 4 4 5.6 4 7.6v8.8C4 18.39 5.61 20 7.6 20h8.8c1.99 0 3.6-1.61 3.6-3.6V7.6C20 5.61 18.39 4 16.4 4H7.6m9.65 1.5a1.25 1.25 0 0 1 1.25 1.25A1.25 1.25 0 0 1 17.25 8 1.25 1.25 0 0 1 16 6.75a1.25 1.25 0 0 1 1.25-1.25M12 7a5 5 0 0 1 5 5 5 5 0 0 1-5 5 5 5 0 0 1-5-5 5 5 0 0 1 5-5m0 2a3 3 0 0 0-3 3 3 3 0 0 0 3 3 3 3 0 0 0 3-3 3 3 0 0 0-3-3z" />
  </SvgIcon>
);

// Provider icon and display name mapping
const providerConfig: Record<string, { icon: React.ReactElement; name: string; color?: string }> = {
  google: { icon: <Google />, name: 'Google', color: '#4285F4' },
  github: { icon: <GitHub />, name: 'GitHub', color: '#24292e' },
  gitlab: { icon: <Code />, name: 'GitLab', color: '#FC6D26' },
  facebook: { icon: <Facebook />, name: 'Facebook', color: '#1877F2' },
  instagram: { icon: <InstagramIcon />, name: 'Instagram', color: '#E4405F' },
  linkedin: { icon: <LinkedIn />, name: 'LinkedIn', color: '#0A66C2' },
  yandex: { icon: <YandexIcon />, name: 'Yandex', color: '#FF0000' },
  vk: { icon: <VKIcon />, name: 'VK', color: '#0077FF' },
};

const LoginPage: React.FC = () => {
  const { user, loading } = useAuth();
  const navigate = useNavigate();
  const [providers, setProviders] = useState<string[]>([]);
  const [loadingProviders, setLoadingProviders] = useState(true);

  useEffect(() => {
    // Redirect to home if already logged in
    if (!loading && user) {
      navigate('/tree', { replace: true });
    }
  }, [user, loading, navigate]);

  useEffect(() => {
    // Fetch available providers
    const fetchProviders = async () => {
      try {
        const availableProviders = await authApi.getProviders();
        setProviders(availableProviders);
      } catch (error) {
        console.error('Failed to fetch providers:', error);
        // Fallback to google if API fails
        setProviders(['google']);
      } finally {
        setLoadingProviders(false);
      }
    };

    fetchProviders();
  }, []);

  const handleProviderLogin = async (provider: string) => {
    try {
      const { url } = await authApi.getAuthURL(provider);
      window.location.href = url;
    } catch (error) {
      console.error(`Failed to get auth URL for ${provider}:`, error);
    }
  };

  // Show nothing while checking auth status
  if (loading) {
    return null;
  }

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
        <Card sx={{ width: '100%' }}>
          <CardContent sx={{ textAlign: 'center', p: 4 }}>
            <Typography variant="h4" component="h1" gutterBottom>
              Family Tree
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
              Sign in to access your family tree
            </Typography>

            {loadingProviders ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
                <CircularProgress />
              </Box>
            ) : (
              <Stack spacing={2}>
                {providers.map((provider) => {
                  const config = providerConfig[provider] || {
                    icon: <Language />,
                    name: provider.charAt(0).toUpperCase() + provider.slice(1),
                  };

                  return (
                    <Button
                      key={provider}
                      variant="contained"
                      size="large"
                      startIcon={config.icon}
                      onClick={() => handleProviderLogin(provider)}
                      fullWidth
                      sx={{
                        py: 1.5,
                        ...(config.color && {
                          backgroundColor: config.color,
                          '&:hover': {
                            backgroundColor: config.color,
                            filter: 'brightness(0.9)',
                          },
                        }),
                      }}
                    >
                      Sign in with {config.name}
                    </Button>
                  );
                })}
              </Stack>
            )}
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
};

export default LoginPage;
