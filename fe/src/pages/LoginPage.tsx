import React, { useEffect, useState } from 'react';
import { Box, Card, CardContent, Typography, Container, CircularProgress } from '@mui/material';
import { Language } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../contexts/AuthContext';
import { useTheme } from '../contexts/ThemeContext';
import { authApi } from '../api';

import googleIcon from '../assets/icons/google.svg';
import githubIcon from '../assets/icons/github.svg';
import gitlabIcon from '../assets/icons/gitlab.svg';
import yandexIcon from '../assets/icons/yandex.svg';

// Provider icon and display name mapping
const providerConfig: Record<string, { icon: React.ReactElement; name: string }> = {
  google: {
    icon: <img src={googleIcon} alt="Google" style={{ width: '1.8rem', height: '1.8rem' }} />,
    name: 'Google'
  },
  github: {
    icon: <img src={githubIcon} alt="GitHub" style={{ width: '1.8rem', height: '1.8rem' }} />,
    name: 'GitHub'
  },
  gitlab: {
    icon: <img src={gitlabIcon} alt="GitLab" style={{ width: '1.8rem', height: '1.8rem' }} />,
    name: 'GitLab'
  },
  yandex: {
    icon: <img src={yandexIcon} alt="Yandex" style={{ width: '1.8rem', height: '1.8rem' }} />,
    name: 'Yandex'
  },
};

const LoginPage: React.FC = () => {
  const { t } = useTranslation();
  const { user, loading } = useAuth();
  const { mode } = useTheme();
  const navigate = useNavigate();
  const [providers, setProviders] = useState<string[]>([]);
  const [loadingProviders, setLoadingProviders] = useState(true);
  const isDarkMode = mode === 'dark';

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

    }
  };

  // Show nothing while checking auth status
  if (loading) {
    return null;
  }

  return (
    <Box
      sx={{
        minHeight: '100vh',
        position: 'relative',
        overflow: 'hidden',
        background: isDarkMode
          ? 'linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)'
          : 'linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%)',
        '&::before': {
          content: '""',
          position: 'absolute',
          top: '-50%',
          right: '-50%',
          width: '100%',
          height: '100%',
          background: isDarkMode
            ? 'radial-gradient(circle, rgba(99, 102, 241, 0.1) 0%, transparent 70%)'
            : 'radial-gradient(circle, rgba(255, 255, 255, 0.3) 0%, transparent 70%)',
          animation: 'float 20s ease-in-out infinite',
        },
        '&::after': {
          content: '""',
          position: 'absolute',
          bottom: '-50%',
          left: '-50%',
          width: '100%',
          height: '100%',
          background: isDarkMode
            ? 'radial-gradient(circle, rgba(139, 92, 246, 0.1) 0%, transparent 70%)'
            : 'radial-gradient(circle, rgba(255, 255, 255, 0.2) 0%, transparent 70%)',
          animation: 'float 25s ease-in-out infinite reverse',
        },
        '@keyframes float': {
          '0%, 100%': {
            transform: 'translate(0, 0) scale(1)',
          },
          '33%': {
            transform: 'translate(30px, -50px) scale(1.1)',
          },
          '66%': {
            transform: 'translate(-20px, 20px) scale(0.9)',
          },
        },
      }}
    >
      <Container maxWidth="sm" sx={{ position: 'relative', zIndex: 1 }}>
        <Box
          sx={{
            minHeight: '100vh',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5, ease: [0.4, 0.0, 0.2, 1] }}
          style={{ width: '100%' }}
        >
          <Card
            sx={{
              width: '100%',
              background: isDarkMode
                ? 'rgba(255, 255, 255, 0.05)'
                : 'rgba(255, 255, 255, 0.25)',
              backdropFilter: 'blur(20px)',
              border: isDarkMode
                ? '1px solid rgba(255, 255, 255, 0.1)'
                : '1px solid rgba(255, 255, 255, 0.3)',
              boxShadow: isDarkMode
                ? '0 8px 32px rgba(0, 0, 0, 0.5)'
                : '0 8px 32px rgba(31, 38, 135, 0.37)',
            }}
          >
            <CardContent sx={{ textAlign: 'center', p: 4 }}>
              <motion.div
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.2 }}
              >
                <Typography
                  variant="h4"
                  component="h1"
                  gutterBottom
                  fontWeight="bold"
                  sx={{
                    color: isDarkMode ? '#fff' : '#fff',
                    textShadow: isDarkMode
                      ? '0 2px 10px rgba(0, 0, 0, 0.3)'
                      : '0 2px 10px rgba(0, 0, 0, 0.1)',
                  }}
                >
                  {t('navigation.familyTree')}
                </Typography>
                <Typography
                  variant="body1"
                  sx={{
                    mb: 4,
                    color: isDarkMode ? 'rgba(255, 255, 255, 0.8)' : 'rgba(255, 255, 255, 0.9)',
                    textShadow: isDarkMode
                      ? '0 1px 5px rgba(0, 0, 0, 0.2)'
                      : '0 1px 5px rgba(0, 0, 0, 0.1)',
                  }}
                >
                  {t('auth.signInToAccess')}
                </Typography>
              </motion.div>

              {loadingProviders ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
                  <CircularProgress />
                </Box>
              ) : (
                <Box sx={{ display: 'flex', justifyContent: 'center', gap: 3, flexWrap: 'wrap' }}>
                  {providers.map((provider, index) => {
                    const config = providerConfig[provider] || {
                      icon: <Language style={{ width: '1.8rem', height: '1.8rem' }} />,
                      name: provider.charAt(0).toUpperCase() + provider.slice(1),
                    };

                    return (
                      <motion.div
                        key={provider}
                        initial={{ opacity: 0, scale: 0.8 }}
                        animate={{ opacity: 1, scale: 1 }}
                        transition={{ duration: 0.4, delay: 0.3 + index * 0.1 }}
                      >
                        <Box
                          onClick={() => handleProviderLogin(provider)}
                          sx={{
                            position: 'relative',
                            width: 64,
                            height: 64,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            borderRadius: '50%',
                            border: isDarkMode ? '2px solid rgba(255, 255, 255, 0.15)' : '2px solid rgba(255, 255, 255, 0.4)',
                            backgroundColor: isDarkMode ? 'rgba(255, 255, 255, 0.08)' : 'rgba(255, 255, 255, 0.5)',
                            backdropFilter: 'blur(10px)',
                            cursor: 'pointer',
                            overflow: 'visible',
                            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                            '&::before': {
                              content: '""',
                              position: 'absolute',
                              bottom: -4,
                              left: '50%',
                              transform: 'translateX(-50%)',
                              width: '0%',
                              height: '3px',
                              borderRadius: '2px',
                              background: isDarkMode
                                ? 'linear-gradient(90deg, transparent, #fff, transparent)'
                                : 'linear-gradient(90deg, transparent, #4285F4, transparent)',
                              transition: 'width 0.4s cubic-bezier(0.4, 0, 0.2, 1)',
                              boxShadow: isDarkMode
                                ? '0 0 20px rgba(255, 255, 255, 0.8)'
                                : '0 0 20px rgba(66, 133, 244, 0.6)',
                            },
                            '&:hover': {
                              borderColor: isDarkMode ? 'rgba(255, 255, 255, 0.3)' : 'rgba(255, 255, 255, 0.6)',
                              backgroundColor: isDarkMode ? 'rgba(255, 255, 255, 0.12)' : 'rgba(255, 255, 255, 0.7)',
                              boxShadow: isDarkMode
                                ? '0 8px 24px rgba(0, 0, 0, 0.4), 0 0 0 1px rgba(255, 255, 255, 0.1)'
                                : '0 8px 24px rgba(0, 0, 0, 0.12)',
                              '&::before': {
                                width: '80%',
                              },
                            },
                            '&:active': {
                              transform: 'scale(0.95)',
                            },
                          }}
                        >
                          {config.icon}
                        </Box>
                      </motion.div>
                    );
                  })}
                </Box>
              )}
            </CardContent>
          </Card>
        </motion.div>
        </Box>
      </Container>
    </Box>
  );
};

export default LoginPage;
