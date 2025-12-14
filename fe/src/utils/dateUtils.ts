import { format, parseISO } from 'date-fns';

export const formatDate = (date: string | null | undefined): string => {
  if (!date) return 'N/A';
  try {
    return format(parseISO(date), 'MMM dd, yyyy');
  } catch {
    return 'Invalid date';
  }
};

export const calculateAge = (birthDate: string | null, deathDate?: string | null): number | null => {
  if (!birthDate) return null;
  try {
    const birth = parseISO(birthDate);
    const end = deathDate ? parseISO(deathDate) : new Date();
    const age = end.getFullYear() - birth.getFullYear();
    return age;
  } catch {
    return null;
  }
};

export const toISODate = (date: Date | null): string | null => {
  if (!date) return null;
  return date.toISOString().split('T')[0];
};



