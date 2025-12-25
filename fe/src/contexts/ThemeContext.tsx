import React, { createContext, useContext, useState, useEffect, useMemo } from 'react';
import { ThemeProvider as MUIThemeProvider, createTheme, PaletteMode } from '@mui/material';
import { CssBaseline } from '@mui/material';
import { useTranslation } from 'react-i18next';

interface ThemeContextType {
  mode: PaletteMode;
  toggleTheme: () => void;
  setThemeMode: (mode: PaletteMode) => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export const useTheme = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
};

interface ThemeProviderProps {
  children: React.ReactNode;
}

export const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const { i18n } = useTranslation();
  const [mode, setMode] = useState<PaletteMode>(() => {
    const savedMode = localStorage.getItem('themeMode');
    return (savedMode as PaletteMode) || 'light';
  });

  useEffect(() => {
    localStorage.setItem('themeMode', mode);
  }, [mode]);

  const toggleTheme = () => {
    setMode((prevMode) => (prevMode === 'light' ? 'dark' : 'light'));
  };

  const setThemeMode = (newMode: PaletteMode) => {
    setMode(newMode);
  };

  const theme = useMemo(
    () =>
      createTheme({
        direction: i18n.language === 'ar' ? 'rtl' : 'ltr',
        palette: {
          mode,
          ...(mode === 'light'
            ? {
                // Light mode colors
                primary: {
                  main: '#6366f1', // Indigo
                  light: '#818cf8',
                  dark: '#4f46e5',
                  contrastText: '#ffffff',
                },
                secondary: {
                  main: '#ec4899', // Pink
                  light: '#f472b6',
                  dark: '#db2777',
                  contrastText: '#ffffff',
                },
                background: {
                  default: '#f8fafc',
                  paper: '#ffffff',
                },
                text: {
                  primary: '#1e293b',
                  secondary: '#64748b',
                },
                success: {
                  main: '#10b981',
                  light: '#34d399',
                  dark: '#059669',
                },
                warning: {
                  main: '#f59e0b',
                  light: '#fbbf24',
                  dark: '#d97706',
                },
                error: {
                  main: '#ef4444',
                  light: '#f87171',
                  dark: '#dc2626',
                },
                info: {
                  main: '#3b82f6',
                  light: '#60a5fa',
                  dark: '#2563eb',
                },
              }
            : {
                // Dark mode colors
                primary: {
                  main: '#818cf8', // Lighter indigo for dark mode
                  light: '#a5b4fc',
                  dark: '#6366f1',
                  contrastText: '#000000',
                },
                secondary: {
                  main: '#f472b6', // Lighter pink for dark mode
                  light: '#f9a8d4',
                  dark: '#ec4899',
                  contrastText: '#000000',
                },
                background: {
                  default: '#0f172a', // Slate 900
                  paper: '#1e293b', // Slate 800
                },
                text: {
                  primary: '#f1f5f9',
                  secondary: '#94a3b8',
                },
                success: {
                  main: '#34d399',
                  light: '#6ee7b7',
                  dark: '#10b981',
                },
                warning: {
                  main: '#fbbf24',
                  light: '#fcd34d',
                  dark: '#f59e0b',
                },
                error: {
                  main: '#f87171',
                  light: '#fca5a5',
                  dark: '#ef4444',
                },
                info: {
                  main: '#60a5fa',
                  light: '#93c5fd',
                  dark: '#3b82f6',
                },
              }),
        },
        typography: {
          fontFamily: [
            '-apple-system',
            'BlinkMacSystemFont',
            '"Segoe UI"',
            'Roboto',
            '"Helvetica Neue"',
            'Arial',
            'sans-serif',
          ].join(','),
          h1: {
            fontWeight: 700,
            letterSpacing: '-0.02em',
          },
          h2: {
            fontWeight: 700,
            letterSpacing: '-0.01em',
          },
          h3: {
            fontWeight: 600,
            letterSpacing: '-0.01em',
          },
          h4: {
            fontWeight: 600,
            letterSpacing: '-0.005em',
          },
          h5: {
            fontWeight: 600,
          },
          h6: {
            fontWeight: 600,
          },
          button: {
            fontWeight: 500,
            textTransform: 'none',
          },
        },
        shape: {
          borderRadius: 12,
        },
        shadows: [
          'none',
          mode === 'light'
            ? '0 1px 2px 0 rgba(0, 0, 0, 0.05)'
            : '0 1px 2px 0 rgba(0, 0, 0, 0.3)',
          mode === 'light'
            ? '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)'
            : '0 1px 3px 0 rgba(0, 0, 0, 0.4), 0 1px 2px 0 rgba(0, 0, 0, 0.2)',
          mode === 'light'
            ? '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)'
            : '0 4px 6px -1px rgba(0, 0, 0, 0.4), 0 2px 4px -1px rgba(0, 0, 0, 0.2)',
          mode === 'light'
            ? '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)'
            : '0 10px 15px -3px rgba(0, 0, 0, 0.5), 0 4px 6px -2px rgba(0, 0, 0, 0.3)',
          mode === 'light'
            ? '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)'
            : '0 20px 25px -5px rgba(0, 0, 0, 0.5), 0 10px 10px -5px rgba(0, 0, 0, 0.3)',
          mode === 'light'
            ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)'
            : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
          mode === 'light' ? '0 25px 50px -12px rgba(0, 0, 0, 0.25)' : '0 25px 50px -12px rgba(0, 0, 0, 0.6)',
        ],
        components: {
          MuiButton: {
            styleOverrides: {
              root: {
                borderRadius: 8,
                padding: '8px 16px',
                boxShadow: 'none',
                '&:hover': {
                  boxShadow: 'none',
                },
              },
              contained: {
                '&:hover': {
                  transform: 'translateY(-1px)',
                  transition: 'transform 0.2s ease-in-out',
                },
              },
            },
          },
          MuiPaper: {
            styleOverrides: {
              root: {
                backgroundImage: 'none',
              },
              elevation1: {
                boxShadow:
                  mode === 'light'
                    ? '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)'
                    : '0 1px 3px 0 rgba(0, 0, 0, 0.4), 0 1px 2px 0 rgba(0, 0, 0, 0.2)',
              },
            },
          },
          MuiCard: {
            styleOverrides: {
              root: {
                borderRadius: 12,
                transition: 'transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out',
                '&:hover': {
                  transform: 'translateY(-2px)',
                  boxShadow:
                    mode === 'light'
                      ? '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)'
                      : '0 10px 15px -3px rgba(0, 0, 0, 0.5), 0 4px 6px -2px rgba(0, 0, 0, 0.3)',
                },
              },
            },
          },
          MuiAppBar: {
            styleOverrides: {
              root: {
                boxShadow: 'none',
                borderBottom: `1px solid ${mode === 'light' ? '#e2e8f0' : '#334155'}`,
              },
            },
          },
          MuiDrawer: {
            styleOverrides: {
              paper: {
                backgroundImage: 'none',
                borderRight: `1px solid ${mode === 'light' ? '#e2e8f0' : '#334155'}`,
              },
            },
          },
          MuiTextField: {
            styleOverrides: {
              root: {
                '& .MuiOutlinedInput-root': {
                  transition: 'all 0.2s ease-in-out',
                  '&:hover': {
                    transform: 'translateY(-1px)',
                  },
                  '&.Mui-focused': {
                    transform: 'translateY(-1px)',
                  },
                },
              },
            },
          },
        },
      }),
    [mode, i18n.language]
  );

  return (
    <ThemeContext.Provider value={{ mode, toggleTheme, setThemeMode }}>
      <MUIThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </MUIThemeProvider>
    </ThemeContext.Provider>
  );
};
