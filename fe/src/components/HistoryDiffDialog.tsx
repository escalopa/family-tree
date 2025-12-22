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
      arabic_name: 'Arabic Name',
      english_name: 'English Name',
      gender: 'Gender',
      picture: 'Picture',
      date_of_birth: 'Date of Birth',
      date_of_death: 'Date of Death',
      father_id: 'Father ID',
      mother_id: 'Mother ID',
      nicknames: 'Nicknames',
      profession: 'Profession',
    };
    return labels[field] || field;
  };

  const formatValue = (value: any): string => {
    if (value === null || value === undefined) return '-';
    if (typeof value === 'boolean') return value ? 'Yes' : 'No';
    if (Array.isArray(value)) return value.join(', ');
    if (typeof value === 'object') return JSON.stringify(value);
    return String(value);
  };

  const getChangedFields = (): Array<{ field: string; oldValue: any; newValue: any }> => {
    const changes: Array<{ field: string; oldValue: any; newValue: any }> = [];

    if (history.change_type === 'INSERT') {
      // For INSERT, show all new values
      Object.entries(history.new_values || {}).forEach(([field, value]) => {
        if (field !== 'member_id' && field !== 'version' && field !== 'deleted_at') {
          changes.push({ field, oldValue: null, newValue: value });
        }
      });
    } else if (history.change_type === 'DELETE') {
      // For DELETE, show all old values
      Object.entries(history.old_values || {}).forEach(([field, value]) => {
        if (field !== 'member_id' && field !== 'version' && field !== 'deleted_at') {
          changes.push({ field, oldValue: value, newValue: null });
        }
      });
    } else if (history.change_type === 'UPDATE') {
      // For UPDATE, show only changed fields
      const oldValues = history.old_values || {};
      const newValues = history.new_values || {};
      const allFields = new Set([...Object.keys(oldValues), ...Object.keys(newValues)]);

      allFields.forEach((field) => {
        if (field !== 'member_id' && field !== 'version' && field !== 'deleted_at') {
          const oldValue = oldValues[field];
          const newValue = newValues[field];

          // Only show if values are different
          if (JSON.stringify(oldValue) !== JSON.stringify(newValue)) {
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
