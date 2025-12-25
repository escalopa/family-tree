import React, { useState, useEffect } from 'react';
import { Autocomplete, TextField, Avatar, Box, Typography, CircularProgress } from '@mui/material';
import { useTranslation } from 'react-i18next';
import { membersApi } from '../api';
import { getMemberPictureUrl } from '../utils/helpers';
import { MemberListItem, MemberInfo } from '../types';

interface ParentAutocompleteProps {
  label: string;
  gender: 'M' | 'F';
  value: number | null;
  onChange: (value: number | null) => void;
  disabled?: boolean;
  initialParent?: MemberInfo | null;
}

const ParentAutocomplete: React.FC<ParentAutocompleteProps> = ({
  label,
  gender,
  value,
  onChange,
  disabled = false,
  initialParent = null,
}) => {
  const { t } = useTranslation();
  const [options, setOptions] = useState<MemberListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [selectedOption, setSelectedOption] = useState<MemberListItem | MemberInfo | null>(null);
  const [initialized, setInitialized] = useState(false);

  // Debounced search
  useEffect(() => {
    if (inputValue.length < 2) {
      // Don't clear options immediately - let user see previous results
      if (inputValue.length === 0) {
        setOptions([]);
      }
      return;
    }

    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const result = await membersApi.searchMembers({
          name: inputValue,
          gender,
          limit: 20,
        });
        console.log('Parent search results:', result); // Debug log
        if (Array.isArray(result.members)) {
          setOptions(result.members);
        } else {
          console.warn('Search results is not an array:', result);
          setOptions([]);
        }
      } catch (error) {
        console.error('search parents:', error);
        setOptions([]);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [inputValue, gender]);

  // Initialize selected option on mount or when value changes
  useEffect(() => {
    if (value && initialParent && initialParent.member_id === value && !initialized) {
      setSelectedOption(initialParent);
      setInitialized(true);
    } else if (!value && selectedOption) {
      // Clear selection if value becomes null/undefined
      setSelectedOption(null);
      setInputValue('');
      setInitialized(false);
    }
  }, [value, initialParent, initialized, selectedOption]);

  // Update selected option from search results
  useEffect(() => {
    if (value && !selectedOption && options.length > 0) {
      const found = options.find(opt => opt.member_id === value);
      if (found) {
        setSelectedOption(found);
        setInitialized(true);
      }
    }
  }, [value, options, selectedOption]);

  return (
    <Autocomplete
      options={options}
      value={selectedOption}
      onChange={(_, newValue) => {
        setSelectedOption(newValue);
        onChange(newValue ? newValue.member_id : null);
      }}
      inputValue={inputValue}
      onInputChange={(_, newInputValue, reason) => {
        // Clear selectedOption when user clears the input
        if (reason === 'clear') {
          setSelectedOption(null);
          onChange(null);
          setInputValue('');
        } else {
          setInputValue(newInputValue);
        }
      }}
      getOptionLabel={(option) => option.name}
      isOptionEqualToValue={(option, value) => option.member_id === value.member_id}
      filterOptions={(x) => x} // Don't filter options - backend already filtered them
      loading={loading}
      disabled={disabled}
      freeSolo={false}
      clearOnEscape
      clearOnBlur={false}
      disableClearable={false} // Make sure clear button is enabled
      renderInput={(params) => (
        <TextField
          {...params}
          label={label}
          placeholder={selectedOption ? '' : t('member.searchByName')}
          helperText={selectedOption ? `${t('language.selected')}: ${selectedOption.name}` : undefined}
          InputProps={{
            ...params.InputProps,
            startAdornment: selectedOption && (
              <Avatar
                src={getMemberPictureUrl(selectedOption.member_id, selectedOption.picture) || undefined}
                sx={{
                  width: 32,
                  height: 32,
                  bgcolor: gender === 'M' ? '#00BCD4' : '#E91E63',
                  mr: 1,
                }}
              >
                {selectedOption.name.charAt(0) || '?'}
              </Avatar>
            ),
            endAdornment: (
              <>
                {loading ? <CircularProgress color="inherit" size={20} /> : null}
                {params.InputProps.endAdornment}
              </>
            ),
          }}
        />
      )}
      renderOption={(props, option) => (
        <Box component="li" {...props} key={option.member_id} sx={{ display: 'flex', gap: 1.5, alignItems: 'center' }}>
          <Avatar
            src={getMemberPictureUrl(option.member_id, option.picture) || undefined}
            sx={{
              width: 40,
              height: 40,
              bgcolor: option.gender === 'M' ? '#00BCD4' : '#E91E63',
            }}
          >
            {option.name.charAt(0) || '?'}
          </Avatar>
          <Box>
            <Typography variant="body2">{option.name}</Typography>
          </Box>
        </Box>
      )}
      noOptionsText={
        loading
          ? t('common.loading')
          : inputValue.length < 2
          ? t('member.searchByName')
          : `${t('member.noMembers')} (${gender === 'M' ? t('member.male') : t('member.female')})`
      }
    />
  );
};

export default ParentAutocomplete;
