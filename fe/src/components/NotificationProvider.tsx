import React from 'react';
import { SnackbarProvider, closeSnackbar } from 'notistack';
import { styled } from '@mui/material/styles';
import { CheckCircle, Error, Warning, Info, Close } from '@mui/icons-material';
import { IconButton } from '@mui/material';
import { useTranslation } from 'react-i18next';

const StyledSnackbarProvider = styled(SnackbarProvider)(({ theme }) => ({
  '&.SnackbarItem-variantSuccess': {
    backgroundColor: theme.palette.success.main,
    color: theme.palette.success.contrastText,
    borderRadius: theme.shape.borderRadius,
    boxShadow: theme.shadows[3],
    padding: theme.spacing(1.5, 2),
    '& .MuiSvgIcon-root': {
      marginInlineEnd: theme.spacing(1.5),
    },
  },
  '&.SnackbarItem-variantError': {
    backgroundColor: theme.palette.error.main,
    color: theme.palette.error.contrastText,
    borderRadius: theme.shape.borderRadius,
    boxShadow: theme.shadows[3],
    padding: theme.spacing(1.5, 2),
    '& .MuiSvgIcon-root': {
      marginInlineEnd: theme.spacing(1.5),
    },
  },
  '&.SnackbarItem-variantWarning': {
    backgroundColor: theme.palette.warning.main,
    color: theme.palette.warning.contrastText,
    borderRadius: theme.shape.borderRadius,
    boxShadow: theme.shadows[3],
    padding: theme.spacing(1.5, 2),
    '& .MuiSvgIcon-root': {
      marginInlineEnd: theme.spacing(1.5),
    },
  },
  '&.SnackbarItem-variantInfo': {
    backgroundColor: theme.palette.info.main,
    color: theme.palette.info.contrastText,
    borderRadius: theme.shape.borderRadius,
    boxShadow: theme.shadows[3],
    padding: theme.spacing(1.5, 2),
    '& .MuiSvgIcon-root': {
      marginInlineEnd: theme.spacing(1.5),
    },
  },
}));

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
      autoHideDuration={6000}
      iconVariant={{
        success: <CheckCircle />,
        error: <Error />,
        warning: <Warning />,
        info: <Info />,
      }}
      action={(snackbarId) => (
        <IconButton
          size="small"
          aria-label="close"
          color="inherit"
          onClick={() => closeSnackbar(snackbarId)}
        >
          <Close fontSize="small" />
        </IconButton>
      )}
      preventDuplicate
      dense
    >
      {children}
    </StyledSnackbarProvider>
  );
};

export default NotificationProvider;
