import React, { useState, useRef } from 'react';
import {
  Box,
  Avatar,
  IconButton,
  CircularProgress,
  Tooltip,
  Alert,
  Typography,
} from '@mui/material';
import { PhotoCamera, Delete, Close } from '@mui/icons-material';
import { useTranslation } from 'react-i18next';
import { membersApi } from '../api';
import { getGenderColor, getMemberPictureUrl } from '../utils/helpers';

interface MemberPhotoUploadProps {
  memberId: number;
  currentPhoto: string | null;
  memberName: string;
  gender: 'M' | 'F' | 'N';
  version?: number;
  onPhotoChange?: (memberId: number, pictureUrl: string | null) => void;
  size?: number;
  showName?: boolean;
  compact?: boolean;
}

const MemberPhotoUpload: React.FC<MemberPhotoUploadProps> = ({
  memberId,
  currentPhoto,
  memberName,
  gender,
  version,
  onPhotoChange,
  size = 100,
  showName = false,
  compact = false,
}) => {
  const { t } = useTranslation();
  const [uploading, setUploading] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Validate file type
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      setError('Invalid file type. Please upload JPG, PNG, GIF, or WebP');
      return;
    }

    // Validate file size (3MB)
    const maxSize = 3 * 1024 * 1024;
    if (file.size > maxSize) {
      setError('File size exceeds 3MB limit');
      return;
    }

    setError(null);
    setUploading(true);

    try {
      const pictureUrl = await membersApi.uploadPicture(memberId, file);
      onPhotoChange?.(memberId, pictureUrl);
    } catch (err: any) {
      setError(err.response?.data?.error || t('member.failedToUploadPhoto'));
    } finally {
      setUploading(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const handleDelete = async () => {
    if (!confirm(t('member.deletePhotoConfirmation'))) return;

    setError(null);
    setDeleting(true);

    try {
      await membersApi.deletePicture(memberId);
      onPhotoChange?.(memberId, null);
    } catch (err: any) {
      setError(err.response?.data?.error || t('member.failedToDeletePhoto'));
    } finally {
      setDeleting(false);
    }
  };

  const handleUploadClick = () => {
    fileInputRef.current?.click();
  };

  const pictureUrl = getMemberPictureUrl(memberId, currentPhoto, version);

  if (compact) {
    return (
      <Box sx={{ position: 'relative', display: 'inline-block' }}>
        <Avatar
          src={pictureUrl || undefined}
          sx={{
            width: size,
            height: size,
            bgcolor: getGenderColor(gender),
            fontSize: size / 3,
          }}
        >
          {memberName[0]}
        </Avatar>

        {(uploading || deleting) && (
          <Box
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: size,
              height: size,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              bgcolor: 'rgba(0, 0, 0, 0.5)',
              borderRadius: '50%',
            }}
          >
            <CircularProgress size={size / 3} sx={{ color: 'white' }} />
          </Box>
        )}

        <Box
          sx={{
            position: 'absolute',
            bottom: -8,
            right: -8,
            display: 'flex',
            gap: 0.5,
          }}
        >
          <Tooltip title={t('member.uploadPicture')}>
            <IconButton
              size="small"
              onClick={handleUploadClick}
              disabled={uploading || deleting}
              sx={{
                bgcolor: 'primary.main',
                color: 'white',
                '&:hover': { bgcolor: 'primary.dark' },
                width: 28,
                height: 28,
              }}
            >
              <PhotoCamera sx={{ fontSize: 16 }} />
            </IconButton>
          </Tooltip>

          {currentPhoto && (
            <Tooltip title={t('member.deletePicture')}>
              <IconButton
                size="small"
                onClick={handleDelete}
                disabled={uploading || deleting}
                sx={{
                  bgcolor: 'error.main',
                  color: 'white',
                  '&:hover': { bgcolor: 'error.dark' },
                  width: 28,
                  height: 28,
                }}
              >
                <Delete sx={{ fontSize: 16 }} />
              </IconButton>
            </Tooltip>
          )}
        </Box>

        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/jpg,image/png,image/gif,image/webp"
          style={{ display: 'none' }}
          onChange={handleFileSelect}
        />
      </Box>
    );
  }

  const [isHovering, setIsHovering] = useState(false);

  return (
    <Box sx={{ textAlign: 'center' }}>
      <Box
        sx={{ position: 'relative', display: 'inline-block', mb: 2 }}
        onMouseEnter={() => setIsHovering(true)}
        onMouseLeave={() => setIsHovering(false)}
      >
        <Avatar
          src={pictureUrl || undefined}
          sx={{
            width: size,
            height: size,
            bgcolor: getGenderColor(gender),
            fontSize: size / 3,
          }}
        >
          {memberName[0]}
        </Avatar>

        {(uploading || deleting) && (
          <Box
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: size,
              height: size,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              bgcolor: 'rgba(0, 0, 0, 0.5)',
              borderRadius: '50%',
            }}
          >
            <CircularProgress size={size / 3} sx={{ color: 'white' }} />
          </Box>
        )}

        {/* Hover overlay with upload button */}
        {isHovering && !uploading && !deleting && (
          <Box
            sx={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: size,
              height: size,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              bgcolor: 'rgba(0, 0, 0, 0.6)',
              borderRadius: '50%',
              cursor: 'pointer',
            }}
            onClick={handleUploadClick}
          >
            <PhotoCamera sx={{ fontSize: size / 3, color: 'white' }} />
          </Box>
        )}

        {/* Small delete button in corner */}
        {currentPhoto && !uploading && !deleting && (
          <Tooltip title={t('member.deletePicture')}>
            <IconButton
              size="small"
              onClick={handleDelete}
              sx={{
                position: 'absolute',
                bottom: 0,
                right: 0,
                bgcolor: 'error.main',
                color: 'white',
                width: 32,
                height: 32,
                '&:hover': { bgcolor: 'error.dark' },
              }}
            >
              <Delete sx={{ fontSize: 18 }} />
            </IconButton>
          </Tooltip>
        )}
      </Box>

      {showName && (
        <Typography variant="subtitle1" sx={{ mb: 2 }}>
          {memberName}
        </Typography>
      )}

      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/jpg,image/png,image/gif,image/webp"
        style={{ display: 'none' }}
        onChange={handleFileSelect}
      />

      {error && (
        <Alert
          severity="error"
          sx={{ mt: 2 }}
          action={
            <IconButton
              size="small"
              onClick={() => setError(null)}
            >
              <Close fontSize="small" />
            </IconButton>
          }
        >
          {error}
        </Alert>
      )}
    </Box>
  );
};

export default MemberPhotoUpload;
