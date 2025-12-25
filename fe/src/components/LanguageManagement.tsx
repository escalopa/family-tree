import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Switch,
  FormControlLabel,
  Chip,
  Alert,
  CircularProgress,
  Typography,
} from '@mui/material';
import { Add } from '@mui/icons-material';
import { useTranslation } from 'react-i18next';
import { languageApi } from '../api';
import { Language } from '../types';

const LanguageManagement: React.FC = () => {
  const { t } = useTranslation();
  const [languages, setLanguages] = useState<Language[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Dialog state
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingLanguage, setEditingLanguage] = useState<Language | null>(null);
  const [formData, setFormData] = useState({
    language_code: '',
    language_name: '',
    is_active: true,
    display_order: 0,
  });

  useEffect(() => {
    loadLanguages();
  }, []);

  const loadLanguages = async () => {
    try {
      setLoading(true);
      setError(null);
      // Get all languages (including inactive)
      const langs = await languageApi.getLanguages(false);
      setLanguages(langs.sort((a, b) => a.display_order - b.display_order));
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to load languages');
      console.error('load languages:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (language?: Language) => {
    if (language) {
      setEditingLanguage(language);
      setFormData({
        language_code: language.language_code,
        language_name: language.language_name,
        is_active: language.is_active,
        display_order: language.display_order,
      });
    } else {
      setEditingLanguage(null);
      setFormData({
        language_code: '',
        language_name: '',
        is_active: true,
        display_order: languages.length,
      });
    }
    setDialogOpen(true);
    setError(null);
    setSuccess(null);
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
    setEditingLanguage(null);
    setFormData({
      language_code: '',
      language_name: '',
      is_active: true,
      display_order: 0,
    });
  };

  const handleSave = async () => {
    try {
      setLoading(true);
      setError(null);

      if (editingLanguage) {
        // Update existing language
        await languageApi.updateLanguage(editingLanguage.language_code, {
          language_name: formData.language_name,
          is_active: formData.is_active,
          display_order: formData.display_order,
        });
        setSuccess(`Language "${formData.language_name}" updated successfully`);
      } else {
        // Create new language
        await languageApi.createLanguage({
          language_code: formData.language_code,
          language_name: formData.language_name,
          display_order: formData.display_order,
        });
        setSuccess(`Language "${formData.language_name}" created successfully`);
      }

      await loadLanguages();
      handleCloseDialog();

      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to save language');
      console.error('save language:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleActive = async (language: Language, event: React.MouseEvent) => {
    // Stop propagation to prevent opening the edit dialog
    event.stopPropagation();
    try {
      setLoading(true);
      setError(null);

      await languageApi.updateLanguage(language.language_code, {
        language_name: language.language_name,
        is_active: !language.is_active,
        display_order: language.display_order,
      });

      setSuccess(`Language "${language.language_name}" ${!language.is_active ? 'activated' : 'deactivated'}`);
      await loadLanguages();

      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to update language');
      console.error('update language:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h6">{t('language.languageManagement')}</Typography>
        <Button
          variant="contained"
          startIcon={<Add />}
          onClick={() => handleOpenDialog()}
          disabled={loading}
        >
          {t('language.addLanguage')}
        </Button>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess(null)}>
          {success}
        </Alert>
      )}

      {loading && !dialogOpen ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>{t('language.code')}</TableCell>
                <TableCell>{t('language.name')}</TableCell>
                <TableCell>{t('language.status')}</TableCell>
                <TableCell>{t('language.displayOrder')}</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {!Array.isArray(languages) || languages.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} align="center">
                    <Typography color="text.secondary">{t('language.noLanguagesFound')}</Typography>
                  </TableCell>
                </TableRow>
              ) : (
                languages.map((language) => (
                  <TableRow
                    key={language.language_code}
                    onClick={() => handleOpenDialog(language)}
                    sx={{
                      cursor: 'pointer',
                      '&:hover': {
                        backgroundColor: 'action.hover',
                      },
                    }}
                  >
                    <TableCell>
                      <Chip label={language.language_code} size="small" />
                    </TableCell>
                    <TableCell>{language.language_name}</TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', alignItems: 'center' }}>
                        <Switch
                          checked={language.is_active}
                          onChange={(e) => handleToggleActive(language, e as any)}
                          onClick={(e) => e.stopPropagation()}
                          color="primary"
                          disabled={loading}
                        />
                        <Chip
                          label={language.is_active ? t('language.active') : t('language.inactive')}
                          color={language.is_active ? 'success' : 'default'}
                          size="small"
                          sx={{ marginInlineStart: 1 }}
                        />
                      </Box>
                    </TableCell>
                    <TableCell>{language.display_order}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>{editingLanguage ? t('language.editLanguage') : t('language.addLanguage')}</DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
            <TextField
              label={t('language.code')}
              value={formData.language_code}
              onChange={(e) => setFormData({ ...formData, language_code: e.target.value.toLowerCase() })}
              disabled={!!editingLanguage}
              required
              helperText={t('language.codeHelperText')}
              inputProps={{ maxLength: 10 }}
            />
            <TextField
              label={t('language.name')}
              value={formData.language_name}
              onChange={(e) => setFormData({ ...formData, language_name: e.target.value })}
              required
              helperText={t('language.nameHelperText')}
            />
            <TextField
              label={t('language.displayOrder')}
              type="number"
              value={formData.display_order}
              onChange={(e) => setFormData({ ...formData, display_order: parseInt(e.target.value) || 0 })}
              required
              helperText={t('language.displayOrderHelperText')}
            />
            {editingLanguage && (
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.is_active}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                  />
                }
                label={t('language.active')}
              />
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog} disabled={loading}>
            {t('common.cancel')}
          </Button>
          <Button
            onClick={handleSave}
            variant="contained"
            disabled={loading || !formData.language_code || !formData.language_name}
          >
            {loading ? t('common.saving') : t('common.save')}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default LanguageManagement;
