import { Roles } from '../types';

export const getRoleName = (roleId: number): string => {
  switch (roleId) {
    case Roles.NONE:
      return 'None';
    case Roles.GUEST:
      return 'Guest';
    case Roles.ADMIN:
      return 'Admin';
    case Roles.SUPER_ADMIN:
      return 'Super Admin';
    default:
      return 'Unknown';
  }
};

export const formatDate = (dateString: string | null): string => {
  if (!dateString) return '-';
  return new Date(dateString).toLocaleDateString();
};

export const calculateAge = (dateOfBirth: string | null, dateOfDeath: string | null): number | null => {
  if (!dateOfBirth) return null;

  const birth = new Date(dateOfBirth);
  const end = dateOfDeath ? new Date(dateOfDeath) : new Date();

  const age = end.getFullYear() - birth.getFullYear();
  const monthDiff = end.getMonth() - birth.getMonth();

  if (monthDiff < 0 || (monthDiff === 0 && end.getDate() < birth.getDate())) {
    return age - 1;
  }

  return age;
};

export const getGenderColor = (gender: 'M' | 'F' | 'N'): string => {
  switch (gender) {
    case 'M':
      return '#00BCD4'; // Cyan
    case 'F':
      return '#E91E63'; // Pink
    default:
      return '#9E9E9E'; // Grey
  }
};

export const getDefaultAvatar = (gender: 'M' | 'F' | 'N'): string => {
  // Return a data URI or path to default avatar based on gender
  return gender === 'M'
    ? '/default-male-avatar.png'
    : gender === 'F'
    ? '/default-female-avatar.png'
    : '/default-avatar.png';
};
