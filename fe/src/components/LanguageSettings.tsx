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
} from '@mui/material';
import { Save } from '@mui/icons-material';
import { useLanguage } from '../contexts/LanguageContext';

interface LanguageSettingsProps {
  onSave?: () => void;
}

const LanguageSettings: React.FC<LanguageSettingsProps> = ({ onSave }) => {
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

  if (loading && !preferences) {
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
          Language Preference
        </Typography>
        <Typography variant="body2" color="text.secondary" paragraph>
          Choose your preferred language for member names in tree view and avatar initials.
          In list views, all names will be displayed.
        </Typography>

        {error && !preferences && (
          <Alert severity="warning" sx={{ mb: 2 }}>
            {error}. Using default setting (Arabic).
          </Alert>
        )}

        {saveSuccess && (
          <Alert severity="success" sx={{ mb: 2 }}>
            Language preference saved successfully!
          </Alert>
        )}

        {saveError && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {saveError}
          </Alert>
        )}

        <Grid container spacing={2}>
          <Grid item xs={12}>
            <FormControl fullWidth>
              <InputLabel>Preferred Language</InputLabel>
              <Select
                value={preferredLanguage}
                onChange={(e) => setPreferredLanguage(e.target.value)}
                label="Preferred Language"
                disabled={saving}
              >
                {languages.map((lang) => (
                  <MenuItem
                    key={lang.language_code}
                    value={lang.language_code}
                  >
                    {lang.language_name} ({lang.language_code.toUpperCase()})
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>

          <Grid item xs={12}>
            <Button
              variant="contained"
              startIcon={saving ? <CircularProgress size={20} /> : <Save />}
              onClick={handleSave}
              disabled={saving || !preferredLanguage}
              fullWidth
            >
              {saving ? 'Saving...' : 'Save Preference'}
            </Button>
          </Grid>
        </Grid>

        <Box sx={{ mt: 2, p: 2, bgcolor: 'background.default', borderRadius: 1 }}>
          <Typography variant="caption" color="text.secondary">
            <strong>Note:</strong> Your preferred language will be used for tree view and avatar initials.
            In list views (members list, tree list), all available names will be displayed regardless of this setting.
          </Typography>
        </Box>
      </CardContent>
    </Card>
  );
};

export default LanguageSettings;
