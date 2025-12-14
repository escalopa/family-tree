import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Box, Button, Card, CardContent, Typography, Container, Stack, CircularProgress } from '@mui/material';
import GoogleIcon from '@mui/icons-material/Google';
import GitHubIcon from '@mui/icons-material/GitHub';
import FacebookIcon from '@mui/icons-material/Facebook';
import { authApi, OAuthProvider } from '../api/auth';
import { useAuthStore } from '../store/authStore';

// OAuth provider configuration
interface OAuthProviderConfig {
  id: OAuthProvider;
  name: string;
  icon: React.ReactElement;
  enabled: boolean;
}

const OAUTH_PROVIDERS: OAuthProviderConfig[] = [
  {
    id: 'google',
    name: 'Google',
    icon: <GoogleIcon />,
    enabled: true, // Can be controlled via environment variables
  },
  {
    id: 'facebook',
    name: 'Facebook',
    icon: <FacebookIcon />,
    enabled: false, // Set to true when Facebook OAuth is configured
  },
  {
    id: 'github',
    name: 'GitHub',
    icon: <GitHubIcon />,
    enabled: false, // Set to true when GitHub OAuth is configured
  },
];

export const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { isAuthenticated } = useAuthStore();
  const [loading, setLoading] = useState<OAuthProvider | null>(null);

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/');
    }
  }, [isAuthenticated, navigate]);

  const handleOAuthLogin = async (provider: OAuthProvider) => {
    try {
      setLoading(provider);
      const { url } = await authApi.getAuthURL(provider);
      window.location.href = url;
    } catch (error) {
      console.error(`Failed to get ${provider} auth URL:`, error);
      setLoading(null);
    }
  };

  const enabledProviders = OAUTH_PROVIDERS.filter(p => p.enabled);

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
        <Card sx={{ width: '100%', maxWidth: 400 }}>
          <CardContent sx={{ p: 4 }}>
            <Typography variant="h4" align="center" gutterBottom>
              Family Tree
            </Typography>
            <Typography variant="body1" align="center" color="textSecondary" sx={{ mb: 4 }}>
              Sign in to manage your family tree
            </Typography>

            <Stack spacing={2}>
              {enabledProviders.map((provider) => (
                <Button
                  key={provider.id}
                  fullWidth
                  variant="contained"
                  size="large"
                  startIcon={loading === provider.id ? <CircularProgress size={20} color="inherit" /> : provider.icon}
                  onClick={() => handleOAuthLogin(provider.id)}
                  disabled={loading !== null}
                  sx={{ py: 1.5 }}
                >
                  Sign in with {provider.name}
                </Button>
              ))}
            </Stack>

            {searchParams.get('error') && (
              <Typography variant="body2" color="error" align="center" sx={{ mt: 2 }}>
                Authentication failed. Please try again.
              </Typography>
            )}
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
};


