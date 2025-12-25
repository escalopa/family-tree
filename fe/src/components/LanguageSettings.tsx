import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Button,
  Alert,
  CircularProgress,
  Grid,
  useTheme,
} from '@mui/material';
import { Save } from '@mui/icons-material';
import { useTranslation } from 'react-i18next';
import { useLanguage } from '../contexts/LanguageContext';
import DirectionalButton from './DirectionalButton';

interface LanguageSettingsProps {
  onSave?: () => void;
}

const LanguageSettings: React.FC<LanguageSettingsProps> = ({ onSave }) => {
  const { t } = useTranslation();
  const { languages, preferences, loading, error, updatePreferences } = useLanguage();
  const [preferredLanguage, setPreferredLanguage] = useState('');
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveSuccess, setSaveSuccess] = useState(false);

  // Initialize with current preference
  useEffect(() => {
    if (preferences) {
      setPreferredLanguage(preferences.preferred_language);
    }
  }, [preferences]);

  const handleSave = async () => {
    // Validate
    if (!preferredLanguage) {
      setSaveError('Please select a preferred language');
      return;
    }

    try {
      setSaving(true);
      setSaveError(null);
      setSaveSuccess(false);

      await updatePreferences(preferredLanguage);

      setSaveSuccess(true);
      if (onSave) {
        onSave();
      }

      // Clear success message after 3 seconds
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err: any) {
      setSaveError(err?.response?.data?.error || 'Failed to save preference');
    } finally {
      setSaving(false);
    }
  };

  if (loading && languages.length === 0) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          {t('language.namesLanguage')}
        </Typography>
        <Typography variant="body2" color="text.secondary" paragraph>
          {t('language.namesLanguageDescription')}
        </Typography>

        {error && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        {saveSuccess && (
          <Alert severity="success" sx={{ mb: 2 }}>
            {t('language.languageUpdated')}
          </Alert>
        )}

        {saveError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {saveError}
          </Alert>
        )}

        {!Array.isArray(languages) || languages.length === 0 ? (
          <Alert severity="info" sx={{ mb: 2 }}>
            No languages available. Please contact an administrator to activate languages.
          </Alert>
        ) : null}

        <Grid container spacing={2}>
          <Grid item xs={12}>
            <FormControl fullWidth>
              <InputLabel>{t('language.selectLanguage')}</InputLabel>
              <Select
                value={preferredLanguage}
                onChange={(e) => setPreferredLanguage(e.target.value)}
                label={t('language.selectLanguage')}
                disabled={saving || !Array.isArray(languages) || languages.length === 0}
              >
                {Array.isArray(languages) && languages.map((lang) => (
                  <MenuItem
                    key={lang.language_code}
                    value={lang.language_code}
                  >
                    {lang.language_name}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>

          <Grid item xs={12}>
            <DirectionalButton
              variant="contained"
              icon={saving ? <CircularProgress size={20} /> : <Save />}
              onClick={handleSave}
              disabled={saving || !preferredLanguage}
              fullWidth
            >
              {saving ? t('common.loading') : t('common.save')}
            </DirectionalButton>
          </Grid>
        </Grid>
      </CardContent>
    </Card>
  );
};

export default LanguageSettings;
