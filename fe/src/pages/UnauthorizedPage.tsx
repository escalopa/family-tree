import React from 'react';
import { Box, Card, CardContent, Typography, Container, Button } from '@mui/material';
import { Block } from '@mui/icons-material';
import { motion } from 'framer-motion';
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
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.4 }}
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
        </motion.div>
      </Box>
    </Container>
  );
};

export default UnauthorizedPage;
