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

export const formatDateHideYear = (dateString: string | null): string => {
  if (!dateString) return '-';
  const date = new Date(dateString);
  return date.toLocaleDateString(undefined, { month: 'long', day: 'numeric' });
};

export const formatDateOfBirth = (dateString: string | null, isSuperAdmin: boolean): string => {
  if (!dateString) return '-';
  if (isSuperAdmin) {
    return formatDate(dateString);
  }
  return formatDateHideYear(dateString);
};

export const formatDateTime = (dateTimeString: string | null): string => {
  if (!dateTimeString) return '-';
  const date = new Date(dateTimeString);
  return `${date.toLocaleDateString()} ${date.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })}`;
};

export const formatRelativeTime = (dateTimeString: string | null): string => {
  if (!dateTimeString) return '-';

  const date = new Date(dateTimeString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSeconds = Math.floor(diffMs / 1000);
  const diffMinutes = Math.floor(diffSeconds / 60);
  const diffHours = Math.floor(diffMinutes / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSeconds < 60) {
    return 'just now';
  } else if (diffMinutes < 60) {
    return `${diffMinutes} minute${diffMinutes !== 1 ? 's' : ''} ago`;
  } else if (diffHours < 24) {
    return `${diffHours} hour${diffHours !== 1 ? 's' : ''} ago`;
  } else if (diffDays < 7) {
    return `${diffDays} day${diffDays !== 1 ? 's' : ''} ago`;
  } else {
    return formatDateTime(dateTimeString);
  }
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

export const getMemberPictureUrl = (memberId: number, pictureKey: string | null, version?: number): string | null => {
  if (!pictureKey) return null;

  // Get the API base URL from environment or default to current origin
  const apiUrl = import.meta.env.VITE_API_URL || window.location.origin;
  const versionParam = version ? `?v=${version}` : '';
  return `${apiUrl}/api/members/${memberId}/picture${versionParam}`;
};
