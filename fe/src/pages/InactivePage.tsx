import React from 'react';
import { Box, Card, CardContent, Typography, Container } from '@mui/material';
import { Info } from '@mui/icons-material';

const InactivePage: React.FC = () => {
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
            <Info color="warning" sx={{ fontSize: 60, mb: 2 }} />
            <Typography variant="h5" gutterBottom>
              Account Pending Activation
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Your account has been created successfully, but it needs to be activated by an
              administrator before you can access the system.
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
              Please contact an administrator to activate your account.
            </Typography>
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
};

export default InactivePage;
