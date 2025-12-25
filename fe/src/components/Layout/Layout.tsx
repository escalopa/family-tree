import React, { ReactNode, useState, useEffect } from 'react';
import { Box, Container, useMediaQuery, useTheme } from '@mui/material';
import Sidebar from './Sidebar';
import MobileHeader from './MobileHeader';

interface LayoutProps {
  children: ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));

  // Initialize sidebar state from localStorage or default based on screen size
  const [sidebarOpen, setSidebarOpen] = useState(() => {
    const saved = localStorage.getItem('sidebarOpen');
    if (saved !== null) {
      return saved === 'true';
    }
    return !isMobile; // Default to open on desktop, closed on mobile
  });

  // Update localStorage when sidebar state changes
  useEffect(() => {
    localStorage.setItem('sidebarOpen', String(sidebarOpen));
  }, [sidebarOpen]);

  const handleSidebarToggle = () => {
    setSidebarOpen(!sidebarOpen);
  };

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh' }}>
      <Sidebar open={sidebarOpen} onToggle={handleSidebarToggle} />
      {isMobile && <MobileHeader onMenuClick={handleSidebarToggle} />}

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          display: 'flex',
          flexDirection: 'column',
          minHeight: '100vh',
          width: '100%',
          mt: { xs: 7, md: 0 },
          transition: theme.transitions.create(['margin', 'width'], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
          }),
        }}
      >
        <Container
          maxWidth="xl"
          sx={{
            flexGrow: 1,
            py: { xs: 2, md: 4 },
            px: { xs: 2, md: 3 },
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', width: '100%' }}>
            {children}
          </Box>
        </Container>
      </Box>
    </Box>
  );
};

export default Layout;
