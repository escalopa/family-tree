import React from 'react';
import { SnackbarProvider } from 'notistack';
import { styled } from '@mui/material/styles';
import { CheckCircle, Error, Warning, Info } from '@mui/icons-material';
import { Box } from '@mui/material';
import { useTranslation } from 'react-i18next';

const StyledSnackbarProvider = styled(SnackbarProvider)(({ theme }) => ({
  '&.SnackbarItem-variantSuccess': {
    backgroundColor: theme.palette.success.main,
    color: theme.palette.success.contrastText,
    borderRadius: theme.shape.borderRadius,
    boxShadow: theme.shadows[3],
    padding: theme.spacing(1.5, 2),
  },
  '&.SnackbarItem-variantError': {
    backgroundColor: theme.palette.error.main,
    color: theme.palette.error.contrastText,
    borderRadius: theme.shape.borderRadius,
    boxShadow: theme.shadows[3],
    padding: theme.spacing(1.5, 2),
  },
  '&.SnackbarItem-variantWarning': {
    backgroundColor: theme.palette.warning.main,
    color: theme.palette.warning.contrastText,
    borderRadius: theme.shape.borderRadius,
    boxShadow: theme.shadows[3],
    padding: theme.spacing(1.5, 2),
  },
  '&.SnackbarItem-variantInfo': {
    backgroundColor: theme.palette.info.main,
    color: theme.palette.info.contrastText,
    borderRadius: theme.shape.borderRadius,
    boxShadow: theme.shadows[3],
    padding: theme.spacing(1.5, 2),
  },
}));

const IconWrapper: React.FC<{ icon: React.ReactNode }> = ({ icon }) => {
  return (
    <Box sx={{
      display: 'flex',
      alignItems: 'center',
      marginInlineEnd: 1.5
    }}>
      {icon}
    </Box>
  );
};

interface NotificationProviderProps {
  children: React.ReactNode;
}

const NotificationProvider: React.FC<NotificationProviderProps> = ({ children }) => {
  const { i18n } = useTranslation();
  const isRTL = i18n.language === 'ar';

  return (
    <StyledSnackbarProvider
      maxSnack={3}
      anchorOrigin={{
        vertical: 'top',
        horizontal: isRTL ? 'left' : 'right',
      }}
      autoHideDuration={3000}
      iconVariant={{
        success: <IconWrapper icon={<CheckCircle />} />,
        error: <IconWrapper icon={<Error />} />,
        warning: <IconWrapper icon={<Warning />} />,
        info: <IconWrapper icon={<Info />} />,
      }}
      preventDuplicate
      dense
    >
      {children}
    </StyledSnackbarProvider>
  );
};

export default NotificationProvider;
