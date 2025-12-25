import React, { useState, useEffect } from 'react';
import {
  Box,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Alert,
  CircularProgress,
  Typography,
  Switch,
  Button,
} from '@mui/material';
import {
  DragIndicator,
  Save,
} from '@mui/icons-material';
import { useTranslation } from 'react-i18next';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { languageApi } from '../api';
import { Language } from '../types';

interface SortableRowProps {
  language: Language;
  onToggleActive: (language: Language, event: React.ChangeEvent<HTMLInputElement>) => void;
  loading: boolean;
  t: (key: string) => string;
}

function SortableRow({ language, onToggleActive, loading, t }: SortableRowProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: language.language_code });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
    backgroundColor: isDragging ? 'action.hover' : 'inherit',
  };

  return (
    <TableRow ref={setNodeRef} style={style} {...attributes}>
      <TableCell sx={{ width: 50, cursor: 'grab' }} {...listeners}>
        <DragIndicator />
      </TableCell>
      <TableCell>
        <Chip label={language.language_code} size="small" />
      </TableCell>
      <TableCell>{language.language_name}</TableCell>
      <TableCell>
        <Box sx={{ display: 'flex', alignItems: 'center' }}>
          <Switch
            checked={language.is_active}
            onChange={(e) => onToggleActive(language, e)}
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
    </TableRow>
  );
}

const LanguageManagement: React.FC = () => {
  const { t } = useTranslation();
  const [languages, setLanguages] = useState<Language[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [hasChanges, setHasChanges] = useState(false);

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

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
      setHasChanges(false);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to load languages');
      console.error('load languages:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleActive = async (language: Language, event: React.ChangeEvent<HTMLInputElement>) => {
    event.stopPropagation();
    try {
      setLoading(true);
      setError(null);

      await languageApi.toggleLanguageActive(language.language_code, !language.is_active);

      setSuccess(`Language "${language.language_name}" ${!language.is_active ? 'enabled' : 'disabled'}`);
      await loadLanguages();

      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to toggle language');
      console.error('toggle language:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      setLanguages((items) => {
        const oldIndex = items.findIndex((item) => item.language_code === active.id);
        const newIndex = items.findIndex((item) => item.language_code === over.id);

        const newItems = arrayMove(items, oldIndex, newIndex);

        // Update display_order for all items
        const updatedItems = newItems.map((item, index) => ({
          ...item,
          display_order: index,
        }));

        setHasChanges(true);
        return updatedItems;
      });
    }
  };

  const handleSaveOrder = async () => {
    try {
      setSaving(true);
      setError(null);

      // Prepare the order data
      const orderData = languages.map((lang) => ({
        language_code: lang.language_code,
        display_order: lang.display_order,
      }));

      await languageApi.updateLanguageOrder(orderData);

      setSuccess('Language display order updated successfully');
      setHasChanges(false);
      await loadLanguages();

      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to update language order');
      console.error('update language order:', err);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h6">{t('language.languageManagement')}</Typography>
        {hasChanges && (
          <Button
            variant="contained"
            color="primary"
            startIcon={<Save />}
            onClick={handleSaveOrder}
            disabled={saving || loading}
          >
            {saving ? t('common.saving') : t('common.save')}
          </Button>
        )}
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

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell sx={{ width: 50 }}></TableCell>
                  <TableCell>{t('language.code')}</TableCell>
                  <TableCell>{t('language.name')}</TableCell>
                  <TableCell>{t('language.status')}</TableCell>
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
                  <SortableContext
                    items={languages.map((l) => l.language_code)}
                    strategy={verticalListSortingStrategy}
                  >
                    {languages.map((language) => (
                      <SortableRow
                        key={language.language_code}
                        language={language}
                        onToggleActive={handleToggleActive}
                        loading={loading}
                        t={t}
                      />
                    ))}
                  </SortableContext>
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </DndContext>
      )}
    </Box>
  );
};

export default LanguageManagement;
