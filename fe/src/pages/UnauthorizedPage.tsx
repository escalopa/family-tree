import React from 'react';
import { Box, Card, CardContent, Typography, Container, Button } from '@mui/material';
import { Block } from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';

const UnauthorizedPage: React.FC = () => {
  const navigate = useNavigate();

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
        <Card>
          <CardContent sx={{ textAlign: 'center', p: 4 }}>
            <Block color="error" sx={{ fontSize: 60, mb: 2 }} />
            <Typography variant="h5" gutterBottom>
              Unauthorized Access
            </Typography>
            <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
              You don't have permission to access this page.
            </Typography>
            <Button variant="contained" onClick={() => navigate('/tree')}>
              Go to Home
            </Button>
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
};

export default UnauthorizedPage;
