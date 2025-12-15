import React, { ReactNode } from 'react';
import { Box, Container } from '@mui/material';
import Header from './Header';

interface LayoutProps {
  children: ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <Header />
      <Container
        maxWidth="xl"
        sx={{
          flexGrow: 1,
          py: 3,
          display: 'flex',
          flexDirection: 'column'
        }}
      >
        {children}
      </Container>
    </Box>
  );
};

export default Layout;
