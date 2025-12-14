export const GENDER_LABELS = {
  M: 'Male',
  F: 'Female',
  N: 'Other',
} as const;

export const ROLE_LABELS = {
  100: 'None',
  200: 'Guest',
  300: 'Admin',
  400: 'Super Admin',
} as const;

export const GENDER_COLORS = {
  M: '#00bcd4', // Cyan
  F: '#e91e63', // Pink
  N: '#9e9e9e', // Grey
} as const;

export const DEFAULT_AVATAR_MALE = '/avatars/default-male.png';
export const DEFAULT_AVATAR_FEMALE = '/avatars/default-female.png';
export const DEFAULT_AVATAR_OTHER = '/avatars/default-other.png';

export const getDefaultAvatar = (gender: 'M' | 'F' | 'N'): string => {
  switch (gender) {
    case 'M':
      return DEFAULT_AVATAR_MALE;
    case 'F':
      return DEFAULT_AVATAR_FEMALE;
    default:
      return DEFAULT_AVATAR_OTHER;
  }
};



