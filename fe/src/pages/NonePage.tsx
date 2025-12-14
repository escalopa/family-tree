import React from 'react';
import { Box, Card, CardContent, Typography, Container } from '@mui/material';
import HourglassEmptyIcon from '@mui/icons-material/HourglassEmpty';

export const NonePage: React.FC = () => {
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
          <CardContent sx={{ p: 4, textAlign: 'center' }}>
            <HourglassEmptyIcon sx={{ fontSize: 64, color: 'warning.main', mb: 2 }} />
            <Typography variant="h5" gutterBottom>
              Account Pending Activation
            </Typography>
            <Typography variant="body1" color="textSecondary">
              Your account has been created but needs to be activated by an administrator.
              Please contact the admin to grant you access.
            </Typography>
          </CardContent>
        </Card>
      </Box>
    </Container>
  );
};


