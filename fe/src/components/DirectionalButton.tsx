import React from 'react';
import { Button, ButtonProps, Box } from '@mui/material';
import { useIsRTL } from '../utils/rtl';

interface DirectionalButtonProps extends Omit<ButtonProps, 'startIcon' | 'endIcon'> {
  icon?: React.ReactNode;
  iconPosition?: 'start' | 'end';
}

/**
 * Button component that automatically handles icon positioning based on text direction
 * In RTL mode, icons are placed at the end; in LTR mode, at the start
 * Ensures proper spacing between icon and text to prevent overlap
 */
export const DirectionalButton: React.FC<DirectionalButtonProps> = ({
  icon,
  iconPosition = 'start',
  children,
  sx,
  ...props
}) => {
  const isRTL = useIsRTL();

  if (!icon) {
    return <Button sx={sx} {...props}>{children}</Button>;
  }

  // Determine actual icon position based on RTL and user preference
  const shouldPlaceAtEnd = (iconPosition === 'start' && isRTL) || (iconPosition === 'end' && !isRTL);

  const iconProps = shouldPlaceAtEnd
    ? { endIcon: icon }
    : { startIcon: icon };

  // Enhanced styles to ensure proper spacing
  const buttonStyles = {
    ...sx,
    '& .MuiButton-startIcon': {
      marginInlineEnd: '10px',
      marginInlineStart: '-2px',
      display: 'flex',
      alignItems: 'center',
    },
    '& .MuiButton-endIcon': {
      marginInlineStart: '10px',
      marginInlineEnd: '-2px',
      display: 'flex',
      alignItems: 'center',
    },
    paddingLeft: '20px',
    paddingRight: '20px',
    gap: '10px',
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
  };

  return (
    <Button sx={buttonStyles} {...props} {...iconProps}>
      {children}
    </Button>
  );
};

export default DirectionalButton;
