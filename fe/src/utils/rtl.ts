import { useTheme } from '@mui/material';

/**
 * Hook to check if current direction is RTL
 */
export const useIsRTL = (): boolean => {
  const theme = useTheme();
  return theme.direction === 'rtl';
};

/**
 * Get button icon props based on direction
 */
export const getDirectionalIconProps = (icon: React.ReactNode, isRTL: boolean) => {
  return isRTL ? { endIcon: icon } : { startIcon: icon };
};

/**
 * Get text direction for content
 * Use 'auto' for mixed content, specific direction for known content
 */
export const getTextDirection = (content: string | null | undefined): 'ltr' | 'rtl' | 'auto' => {
  if (!content) return 'auto';

  // Check if content starts with RTL character (Arabic, Hebrew, etc.)
  const rtlRegex = /[\u0591-\u07FF\uFB1D-\uFDFD\uFE70-\uFEFC]/;
  const ltrRegex = /[A-Za-z]/;

  const hasRTL = rtlRegex.test(content);
  const hasLTR = ltrRegex.test(content);

  // If mixed content, use auto
  if (hasRTL && hasLTR) return 'auto';
  if (hasRTL) return 'rtl';
  if (hasLTR) return 'ltr';

  return 'auto';
};

/**
 * CSS class for cells with LTR content in RTL tables
 */
export const getLTRCellClass = (content: string | null | undefined): string => {
  const dir = getTextDirection(content);
  return dir === 'ltr' ? 'force-ltr' : '';
};
