import React from 'react';
import { Select, MenuItem, FormControl, InputLabel, SelectChangeEvent } from '@mui/material';
import { useTranslation } from 'react-i18next';
import { useInterfaceLanguage } from '../contexts/InterfaceLanguageContext';

const InterfaceLanguageSelector: React.FC = () => {
  const { t } = useTranslation();
  const { interfaceLanguage, supportedLanguagesWithNames, changeInterfaceLanguage, loading } = useInterfaceLanguage();

  const handleChange = (event: SelectChangeEvent<string>) => {
    changeInterfaceLanguage(event.target.value);
  };

  return (
    <FormControl size="small" sx={{ minWidth: 120 }}>
      <InputLabel id="interface-language-select-label">{t('language.interfaceLanguage')}</InputLabel>
      <Select
        labelId="interface-language-select-label"
        id="interface-language-select"
        value={interfaceLanguage}
        label={t('language.interfaceLanguage')}
        onChange={handleChange}
        disabled={loading}
      >
        {supportedLanguagesWithNames.map((lang) => (
          <MenuItem key={lang.language_code} value={lang.language_code}>
            {lang.language_name}
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
};

export default InterfaceLanguageSelector;
