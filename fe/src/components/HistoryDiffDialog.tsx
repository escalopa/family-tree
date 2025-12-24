import React from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
} from '@mui/material';
import { HistoryRecord } from '../types';
import { formatDateTime, formatRelativeTime } from '../utils/helpers';

interface HistoryDiffDialogProps {
  open: boolean;
  onClose: () => void;
  history: HistoryRecord | null;
}

const HistoryDiffDialog: React.FC<HistoryDiffDialogProps> = ({ open, onClose, history }) => {
  if (!history) return null;

  const getFieldLabel = (field: string): string => {
    const labels: Record<string, string> = {
      gender: 'Gender',
      picture: 'Picture',
      date_of_birth: 'Date of Birth',
      date_of_death: 'Date of Death',
      father_id: 'Father ID',
      mother_id: 'Mother ID',
      nicknames: 'Nicknames',
      profession: 'Profession',
      spouse_id: 'Spouse ID',
      marriage_date: 'Marriage Date',
      divorce_date: 'Divorce Date',
    };

    // Handle dynamic name fields (e.g., name_ar, name_en, name_ru)
    if (field.startsWith('name_')) {
      const langCode = field.substring(5).toUpperCase();
      return `Name (${langCode})`;
    }

    return labels[field] || field;
  };

  const formatValue = (value: any): string => {
    if (value === null || value === undefined) return '-';
    if (typeof value === 'boolean') return value ? 'Yes' : 'No';
    if (Array.isArray(value)) {
      if (value.length === 0) return '-';
      return value.join(', ');
    }
    if (typeof value === 'object') return JSON.stringify(value);
    return String(value);
  };

  // Normalize values for comparison (treat empty arrays, null, undefined as equivalent)
  const normalizeValue = (value: any): any => {
    if (value === null || value === undefined) return null;
    if (Array.isArray(value) && value.length === 0) return null;
    return value;
  };

  // Check if two values are actually different (considering normalization)
  const valuesAreDifferent = (oldValue: any, newValue: any): boolean => {
    const normalizedOld = normalizeValue(oldValue);
    const normalizedNew = normalizeValue(newValue);

    // Both are null/empty - not different
    if (normalizedOld === null && normalizedNew === null) return false;

    // Compare using JSON.stringify for complex objects
    return JSON.stringify(normalizedOld) !== JSON.stringify(normalizedNew);
  };

  const getChangedFields = (): Array<{ field: string; oldValue: any; newValue: any }> => {
    const changes: Array<{ field: string; oldValue: any; newValue: any }> = [];

    if (history.change_type === 'INSERT' || history.change_type === 'ADD_SPOUSE' || history.change_type === 'ADD_PICTURE') {
      // For INSERT/ADD operations, show all new values (except empty ones)
      Object.entries(history.new_values || {}).forEach(([field, value]) => {
        if (field !== 'member_id' && field !== 'version' && field !== 'deleted_at') {
          // Only show if value is not empty/null
          if (normalizeValue(value) !== null) {
            changes.push({ field, oldValue: null, newValue: value });
          }
        }
      });
    } else if (history.change_type === 'DELETE' || history.change_type === 'REMOVE_SPOUSE' || history.change_type === 'DELETE_PICTURE') {
      // For DELETE/REMOVE operations, show all old values (except empty ones)
      Object.entries(history.old_values || {}).forEach(([field, value]) => {
        if (field !== 'member_id' && field !== 'version' && field !== 'deleted_at') {
          // Only show if value is not empty/null
          if (normalizeValue(value) !== null) {
            changes.push({ field, oldValue: value, newValue: null });
          }
        }
      });
    } else if (history.change_type === 'UPDATE' || history.change_type === 'UPDATE_SPOUSE') {
      // For UPDATE operations, show only changed fields
      const oldValues = history.old_values || {};
      const newValues = history.new_values || {};
      const allFields = new Set([...Object.keys(oldValues), ...Object.keys(newValues)]);

      allFields.forEach((field) => {
        if (field !== 'member_id' && field !== 'version' && field !== 'deleted_at') {
          const oldValue = oldValues[field];
          const newValue = newValues[field];

          // Only show if values are actually different (after normalization)
          if (valuesAreDifferent(oldValue, newValue)) {
            changes.push({ field, oldValue, newValue });
          }
        }
      });
    }

    return changes;
  };

  const changes = getChangedFields();

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>
        <Box>
          <Typography variant="h6">Change Details</Typography>
          <Box sx={{ mt: 1, display: 'flex', gap: 1, alignItems: 'center' }}>
            <Chip label={history.change_type} size="small" color="primary" />
            <Typography variant="body2" color="text.secondary">
              by {history.user_full_name} ({history.user_email})
            </Typography>
          </Box>
          <Typography variant="caption" color="text.secondary">
            {formatDateTime(history.changed_at)} • {formatRelativeTime(history.changed_at)} • Version {history.member_version}
          </Typography>
        </Box>
      </DialogTitle>
      <DialogContent>
        {changes.length === 0 ? (
          <Typography color="text.secondary">No field changes to display</Typography>
        ) : (
          <TableContainer component={Paper} variant="outlined">
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell sx={{ fontWeight: 'bold' }}>Field</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>Old Value</TableCell>
                  <TableCell sx={{ fontWeight: 'bold' }}>New Value</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {changes.map((change, index) => (
                  <TableRow key={index}>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontWeight: 'medium' }}>
                        {getFieldLabel(change.field)}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography
                        variant="body2"
                        sx={{
                          color: change.oldValue !== null ? 'error.main' : 'text.secondary',
                          textDecoration: change.oldValue !== null ? 'line-through' : 'none',
                        }}
                      >
                        {formatValue(change.oldValue)}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography
                        variant="body2"
                        sx={{
                          color: change.newValue !== null ? 'success.main' : 'text.secondary',
                          fontWeight: change.newValue !== null ? 'medium' : 'normal',
                        }}
                      >
                        {formatValue(change.newValue)}
                      </Typography>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
};

export default HistoryDiffDialog;
