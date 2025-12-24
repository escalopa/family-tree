import React, { useState, useEffect, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  Box,
  Button,
  Typography,
  TextField,
  Grid,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip,
  CircularProgress,
  InputAdornment,
  Avatar,
  Tabs,
  Tab,
  Autocomplete,
  Tooltip,
} from '@mui/material';
import { Add, Delete, FilterAlt, Clear, Close } from '@mui/icons-material';
import { membersApi } from '../api';
import { Member, MemberListItem, MemberSearchQuery, CreateMemberRequest, UpdateMemberRequest, HistoryRecord, Roles } from '../types';
import { formatDateOfBirth, getMemberPictureUrl, formatDateTime, formatRelativeTime, getChangeTypeColor } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import MemberPhotoUpload from '../components/MemberPhotoUpload';
import ParentAutocomplete from '../components/ParentAutocomplete';
import SpouseCard from '../components/SpouseCard';
import AddSpouseDialog from '../components/AddSpouseDialog';
import HistoryDiffDialog from '../components/HistoryDiffDialog';
import { useAuth } from '../contexts/AuthContext';
import { useLanguage } from '../contexts/LanguageContext';

const PAGE_SIZE = 10;

const MembersPage: React.FC = () => {
  const { hasRole } = useAuth();
  const { languages } = useLanguage();
  const isSuperAdmin = hasRole(Roles.SUPER_ADMIN);
  const [searchParams, setSearchParams] = useSearchParams();
  const [members, setMembers] = useState<MemberListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [openAddSpouseDialog, setOpenAddSpouseDialog] = useState(false);
  const [editingMember, setEditingMember] = useState<Member | null>(null);
  const [originalFormData, setOriginalFormData] = useState<CreateMemberRequest | null>(null);
  const [formData, setFormData] = useState<CreateMemberRequest>({
    names: {},
    gender: 'M',
  });
  const [memberHistory, setMemberHistory] = useState<HistoryRecord[]>([]);
  const [displayedHistoryCount, setDisplayedHistoryCount] = useState(10);
  const [selectedHistory, setSelectedHistory] = useState<HistoryRecord | null>(null);
  const [diffDialogOpen, setDiffDialogOpen] = useState(false);
  const [dialogTab, setDialogTab] = useState(0);
  const tableRef = useRef<HTMLDivElement>(null);
  const loadMoreRef = useRef<HTMLDivElement>(null);

  // Initialize search query from URL params
  const [searchQuery, setSearchQuery] = useState<MemberSearchQuery>(() => {
    const params: MemberSearchQuery = {};
    const name = searchParams.get('name');
    const gender = searchParams.get('gender');
    const married = searchParams.get('married');

    if (name) params.name = name;
    if (gender && (gender === 'M' || gender === 'F')) params.gender = gender as 'M' | 'F';
    if (married && (married === '0' || married === '1')) params.married = Number(married) as 0 | 1;

    return params;
  });

  // Perform search (initial load)
  const performSearch = async (query: MemberSearchQuery, loadMore: boolean = false) => {
    if (loadMore) {
      setLoadingMore(true);
    } else {
      setLoading(true);
      // Don't clear members immediately to prevent jumping
    }

    try {
      const searchParams: MemberSearchQuery = {
        ...query,
        limit: PAGE_SIZE,
        // Only include cursor if we're loading more and have a valid cursor
        ...(loadMore && nextCursor ? { cursor: nextCursor } : {}),
      };

      const response = await membersApi.searchMembers(searchParams);

      if (loadMore) {
        setMembers(prev => [...prev, ...(response.members || [])]);
      } else {
        // Set new members after loading completes for smooth transition
        setMembers(response.members || []);
      }

      // Only set cursor if it exists and is not empty
      const validCursor = response.next_cursor && response.next_cursor.trim() !== '' ? response.next_cursor : null;
      setNextCursor(validCursor);
      setHasMore(!!validCursor);
    } catch (error) {
      console.error('Search failed:', error);
      if (!loadMore) {
        setMembers([]);
      }
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  };

  const handleLoadMore = () => {
    performSearch(searchQuery, true);
  };

  // Update URL params when search query changes
  useEffect(() => {
    const params = new URLSearchParams();

    if (searchQuery.name) params.set('name', searchQuery.name);
    if (searchQuery.gender) params.set('gender', searchQuery.gender);
    if (searchQuery.married !== undefined) params.set('married', String(searchQuery.married));

    setSearchParams(params, { replace: true });
  }, [searchQuery, setSearchParams]);

  // Trigger search when query changes
  useEffect(() => {
    performSearch(searchQuery);
  }, [searchQuery]);

  // Initial load - list all members
  useEffect(() => {
    performSearch(searchQuery);
  }, []);

  // Infinite scroll observer
  useEffect(() => {
    if (!loadMoreRef.current || !hasMore || loadingMore) return;

    const currentRef = loadMoreRef.current;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loadingMore) {
          handleLoadMore();
        }
      },
      { threshold: 0.1 }
    );

    observer.observe(currentRef);

    return () => {
      if (currentRef) {
        observer.unobserve(currentRef);
      }
    };
  }, [hasMore, loadingMore, nextCursor, handleLoadMore]);

  const handleClearFilters = () => {
    setSearchQuery({});
  };

  const handleClearFilter = (filterKey: keyof MemberSearchQuery) => {
    const newQuery = { ...searchQuery };
    delete newQuery[filterKey];
    setSearchQuery(newQuery);
  };

  // Fetch full member details when opening dialog (includes children, siblings, spouses, etc.)
  const handleOpenDialog = async (memberIdOrMember?: number | MemberListItem) => {
    setDialogTab(0); // Reset to first tab
    setMemberHistory([]); // Clear previous history
    setDisplayedHistoryCount(10); // Reset pagination

    if (memberIdOrMember) {
      try {
        // If it's a number (member_id), fetch full details from backend
        // If it's a MemberListItem from the list, also fetch full details to get computed fields
        const memberId = typeof memberIdOrMember === 'number'
          ? memberIdOrMember
          : memberIdOrMember.member_id;

        // Fetch fresh full member data including full names, children, siblings, spouses
        const fullMember = await membersApi.getMember(memberId);

        const initialData = {
          names: fullMember.names || {},
          gender: fullMember.gender,
          date_of_birth: fullMember.date_of_birth || undefined,
          date_of_death: fullMember.date_of_death || undefined,
          father_id: fullMember.father_id || undefined,
          mother_id: fullMember.mother_id || undefined,
          nicknames: fullMember.nicknames || [],
          profession: fullMember.profession || undefined,
        };

        setEditingMember(fullMember);
        setFormData(initialData);
        setOriginalFormData(initialData);

        // Fetch member history for super admins
        if (hasRole(Roles.SUPER_ADMIN)) {
          try {
            const historyResponse = await membersApi.getMemberHistory(memberId);
            setMemberHistory(historyResponse.history || []);
          } catch (error) {
            console.error('load member history:', error);
          }
        }
      } catch (error) {
        console.error('load member details:', error);
        alert('Failed to load member details');
        return;
      }
    } else {
      setEditingMember(null);
      setOriginalFormData(null);
      setFormData({
        names: {},
        gender: 'M',
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setEditingMember(null);
    setOriginalFormData(null);
    setMemberHistory([]);
    setDisplayedHistoryCount(10);
    setDialogTab(0);
  };

  const handleViewDiff = (history: HistoryRecord) => {
    setSelectedHistory(history);
    setDiffDialogOpen(true);
  };

  const handleCloseDiff = () => {
    setDiffDialogOpen(false);
    setSelectedHistory(null);
  };

  const handleOpenRelatedMember = async (memberId: number) => {
    handleCloseDialog(); // Close current dialog first
    setTimeout(() => handleOpenDialog(memberId), 100); // Open new dialog with slight delay to fetch fresh data
  };

  // Check if form has changes
  const hasChanges = () => {
    if (!editingMember || !originalFormData) return true; // Allow create

    // Compare all fields
    if (JSON.stringify(formData.names) !== JSON.stringify(originalFormData.names)) return true;
    if (formData.gender !== originalFormData.gender) return true;
    if (formData.date_of_birth !== originalFormData.date_of_birth) return true;
    if (formData.date_of_death !== originalFormData.date_of_death) return true;
    if (formData.father_id !== originalFormData.father_id) return true;
    if (formData.mother_id !== originalFormData.mother_id) return true;
    if (formData.profession !== originalFormData.profession) return true;

    // Compare nicknames
    const oldNicknames = originalFormData.nicknames || [];
    const newNicknames = formData.nicknames || [];
    if (oldNicknames.length !== newNicknames.length) return true;
    const nicknamesSet = new Set(oldNicknames);
    for (const nickname of newNicknames) {
      if (!nicknamesSet.has(nickname)) return true;
    }

    return false;
  };

  const handleSubmit = async () => {
    // Check if there are any changes for update operations
    if (editingMember && !hasChanges()) {
      handleCloseDialog();
      return;
    }

    // Validate that all active languages have names
    const activeLanguages = Array.isArray(languages) ? languages.filter(lang => lang.is_active) : [];
    const missingLanguages = activeLanguages.filter(
      lang => !formData.names[lang.language_code] || formData.names[lang.language_code].trim() === ''
    );

    if (missingLanguages.length > 0) {
      const missingNames = missingLanguages.map(lang => lang.language_name).join(', ');
      alert(`Please provide names for all active languages: ${missingNames}`);
      return;
    }

    try {
      if (editingMember) {
        const updateData: UpdateMemberRequest = {
          ...formData,
          version: editingMember.version,
        };
        await membersApi.updateMember(editingMember.member_id, updateData);

        // Refetch member data after update to get latest computed fields (includes children)
        const updatedMember = await membersApi.getMember(editingMember.member_id);
        setEditingMember(updatedMember);
      } else {
        await membersApi.createMember(formData);
      }
      handleCloseDialog();
      performSearch(searchQuery); // Refresh list
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error || 'Failed to save member';
      alert(errorMsg);
      console.error('save member:', error);
    }
  };

  const handleDelete = async (memberId: number) => {
    const confirmMessage =
      '⚠️ WARNING: This will DELETE this member and clean up all associated data.\n\n' +
      'The following will be removed:\n' +
      '• All member names in all languages\n' +
      '• All spouse relationships\n' +
      '• Member profile picture from storage\n' +
      '• Member will be marked as deleted (soft delete)\n\n' +
      'A history record will be kept for audit purposes.\n\n' +
      'Are you absolutely sure you want to proceed?';

    if (confirm(confirmMessage)) {
      try {
        await membersApi.deleteMember(memberId);
        handleCloseDialog(); // Close dialog after delete
        performSearch(searchQuery); // Refresh list
      } catch (error: any) {
        console.error('delete member:', error);
        const errorMessage = error?.response?.data?.error || 'Failed to delete member. This member may have children or other dependencies.';
        alert(errorMessage);
      }
    }
  };

  const handlePhotoChange = async (memberId: number, pictureUrl: string | null) => {
    // Update the member picture in the list
    setMembers(prevMembers =>
      prevMembers.map(member =>
        member.member_id === memberId
          ? { ...member, picture: pictureUrl }
          : member
      )
    );

    // If editing this member, refetch full member data
    if (editingMember && editingMember.member_id === memberId) {
      try {
        const updatedMember = await membersApi.getMember(memberId);
        setEditingMember(updatedMember);
      } catch (error) {
        console.error('refresh member data after photo change:', error);
      }
    }
  };

  return (
    <Layout>
      <Box>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
          <Typography variant="h4">Members Management</Typography>
          <Button
            variant="contained"
            startIcon={<Add />}
            onClick={() => handleOpenDialog()}
          >
            Add Member
          </Button>
        </Box>

        {/* Search Filters */}
        <Paper sx={{ p: 2, mb: 3 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
            <FilterAlt sx={{ mr: 1, color: 'text.secondary' }} />
            <Typography variant="h6" sx={{ flexGrow: 1 }}>
              Search Filters {!searchQuery.name && !searchQuery.gender && searchQuery.married === undefined && '(Showing all members)'}
            </Typography>
            {(searchQuery.name || searchQuery.gender || searchQuery.married !== undefined) && (
              <Button
                size="small"
                startIcon={<Clear />}
                onClick={handleClearFilters}
                color="secondary"
              >
                Clear Filters
              </Button>
            )}
          </Box>
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} md={4}>
              <TextField
                fullWidth
                label="Name"
                placeholder="Search by name (Arabic or English)"
                value={searchQuery.name || ''}
                onChange={(e) =>
                  setSearchQuery({ ...searchQuery, name: e.target.value || undefined })
                }
                InputProps={{
                  endAdornment: searchQuery.name && (
                    <InputAdornment position="end">
                      <IconButton
                        size="small"
                        onClick={() => handleClearFilter('name')}
                        edge="end"
                      >
                        <Close fontSize="small" />
                      </IconButton>
                    </InputAdornment>
                  ),
                }}
              />
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <FormControl fullWidth>
                <InputLabel>Gender</InputLabel>
                <Select
                  value={searchQuery.gender || ''}
                  label="Gender"
                  onChange={(e) =>
                    setSearchQuery({ ...searchQuery, gender: (e.target.value as 'M' | 'F') || undefined })
                  }
                  endAdornment={
                    searchQuery.gender && (
                      <InputAdornment position="end" sx={{ mr: 3 }}>
                        <IconButton
                          size="small"
                          onClick={() => handleClearFilter('gender')}
                          edge="end"
                        >
                          <Close fontSize="small" />
                        </IconButton>
                      </InputAdornment>
                    )
                  }
                >
                  <MenuItem value="">All</MenuItem>
                  <MenuItem value="M">Male</MenuItem>
                  <MenuItem value="F">Female</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6} md={4}>
              <FormControl fullWidth>
                <InputLabel>Married</InputLabel>
                <Select
                  value={searchQuery.married ?? ''}
                  label="Married"
                  onChange={(e) =>
                    setSearchQuery({
                      ...searchQuery,
                      married: e.target.value === '' ? undefined : Number(e.target.value) as 0 | 1,
                    })
                  }
                  endAdornment={
                    searchQuery.married !== undefined && (
                      <InputAdornment position="end" sx={{ mr: 3 }}>
                        <IconButton
                          size="small"
                          onClick={() => handleClearFilter('married')}
                          edge="end"
                        >
                          <Close fontSize="small" />
                        </IconButton>
                      </InputAdornment>
                    )
                  }
                >
                  <MenuItem value="">All</MenuItem>
                  <MenuItem value={1}>Yes</MenuItem>
                  <MenuItem value={0}>No</MenuItem>
                </Select>
              </FormControl>
            </Grid>
          </Grid>
          {loading && (
            <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
              <CircularProgress size={24} />
            </Box>
          )}
        </Paper>

        {/* Members Table */}
        <TableContainer
          component={Paper}
          ref={tableRef}
          sx={{
            position: 'relative',
            minHeight: '400px',
            transition: 'opacity 0.3s ease-in-out',
            opacity: loading && members.length === 0 ? 0.6 : 1,
          }}
        >
          {loading && members.length > 0 && (
            <Box
              sx={{
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                height: '4px',
                bgcolor: 'primary.main',
                animation: 'loading 1s ease-in-out infinite',
                '@keyframes loading': {
                  '0%': { transform: 'translateX(-100%)' },
                  '100%': { transform: 'translateX(100%)' },
                },
                zIndex: 1,
              }}
            />
          )}
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>Avatar</TableCell>
                <TableCell>Name</TableCell>
                <TableCell>Gender</TableCell>
                <TableCell>Date of Birth</TableCell>
                <TableCell>Married</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {(!members || members.length === 0) && !loading && (
                <TableRow>
                  <TableCell colSpan={6} align="center" sx={{ py: 4, color: 'text.secondary' }}>
                    {searchQuery.name || searchQuery.gender || searchQuery.married !== undefined
                      ? 'No members found matching your filters'
                      : 'No members found'}
                  </TableCell>
                </TableRow>
              )}
              {loading && members.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} align="center" sx={{ py: 8 }}>
                    <CircularProgress />
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                      Loading members...
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
              {members && members.map((member) => (
                <TableRow
                  key={member.member_id}
                  hover
                  sx={{ cursor: 'pointer' }}
                  onClick={() => handleOpenDialog(member)}
                >
                  <TableCell>
                    <Typography variant="body2" color="text.secondary">
                      #{member.member_id}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Avatar
                      src={getMemberPictureUrl(member.member_id, member.picture) || undefined}
                      sx={{
                        width: 50,
                        height: 50,
                        bgcolor: member.gender === 'M' ? '#00BCD4' : member.gender === 'F' ? '#E91E63' : '#9E9E9E'
                      }}
                    >
                      {member.name?.[0] || '?'}
                    </Avatar>
                  </TableCell>
                  <TableCell>{member.name}</TableCell>
                  <TableCell>
                    {member.gender === 'M' ? 'Male' : 'Female'}
                  </TableCell>
                  <TableCell>{formatDateOfBirth(member.date_of_birth, isSuperAdmin)}</TableCell>
                  <TableCell>
                    {member.is_married ? (
                      <Chip label="Yes" color="primary" size="small" />
                    ) : (
                      <Chip label="No" size="small" />
                    )}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>

        {/* Load More Button */}
        {hasMore && members && members.length > 0 && (
          <Box
            ref={loadMoreRef}
            sx={{
              display: 'flex',
              justifyContent: 'center',
              py: 3,
              minHeight: '60px'
            }}
          >
            {loadingMore ? (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <CircularProgress size={24} />
                <Typography variant="body2" color="text.secondary">
                  Loading more members...
                </Typography>
              </Box>
            ) : (
              <Button
                variant="outlined"
                onClick={handleLoadMore}
                size="large"
              >
                Load More Members
              </Button>
            )}
          </Box>
        )}

        {/* Create/Edit Dialog */}
        <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="lg" fullWidth>
          <DialogTitle>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Typography variant="h6">
                {editingMember ? 'Edit Member' : 'Add New Member'}
              </Typography>
              <IconButton onClick={handleCloseDialog} size="small">
                <Close />
              </IconButton>
            </Box>
          </DialogTitle>
          {editingMember && hasRole(Roles.SUPER_ADMIN) && (
            <Box sx={{ borderBottom: 1, borderColor: 'divider', px: 3 }}>
              <Tabs value={dialogTab} onChange={(_, v) => setDialogTab(v)}>
                <Tab label="Details" />
                <Tab label={`Change History (${memberHistory.length})`} />
              </Tabs>
            </Box>
          )}
          <DialogContent>
            {/* Details Tab */}
            {(!editingMember || dialogTab === 0) && (
              <Grid container spacing={2} sx={{ mt: 1 }}>
              {editingMember && (
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', py: 2 }}>
                    <MemberPhotoUpload
                      memberId={editingMember.member_id}
                      currentPhoto={editingMember.picture}
                      memberName={editingMember.name}
                      gender={editingMember.gender}
                      version={editingMember.version}
                      onPhotoChange={handlePhotoChange}
                      size={120}
                      showName
                    />
                    {editingMember.full_name && (
                      <Box sx={{ mt: 2, textAlign: 'center' }}>
                        <Typography variant="caption" color="text.secondary" gutterBottom>
                          Full Name
                        </Typography>
                        <Typography variant="body2" fontWeight="medium">
                          {editingMember.full_name}
                        </Typography>
                      </Box>
                    )}
                    {editingMember.age !== undefined && editingMember.age !== null && (
                      <Box sx={{ mt: 1, textAlign: 'center' }}>
                        <Typography variant="caption" color="text.secondary">
                          Age: <strong>{editingMember.age} years</strong>
                        </Typography>
                      </Box>
                    )}
                  </Box>
                </Grid>
              )}
              {/* Multi-language name inputs */}
              {Array.isArray(languages) && languages.map((lang) => (
                <Grid item xs={12} sm={6} key={lang.language_code}>
                  <TextField
                    fullWidth
                    label={`${lang.language_name} Name`}
                    value={formData.names?.[lang.language_code] || ''}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        names: {
                          ...formData.names,
                          [lang.language_code]: e.target.value,
                        },
                      })
                    }
                    helperText={`Enter member's name in ${lang.language_name}`}
                  />
                </Grid>
              ))}
              <Grid item xs={12} sm={6}>
                <FormControl fullWidth required>
                  <InputLabel>Gender</InputLabel>
                  <Select
                    value={formData.gender}
                    label="Gender"
                    onChange={(e) =>
                      setFormData({ ...formData, gender: e.target.value as 'M' | 'F' })
                    }
                  >
                    <MenuItem value="M">Male</MenuItem>
                    <MenuItem value="F">Female</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label="Date of Birth"
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  value={formData.date_of_birth || ''}
                  onChange={(e) =>
                    setFormData({ ...formData, date_of_birth: e.target.value })
                  }
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label="Date of Death"
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  value={formData.date_of_death || ''}
                  onChange={(e) =>
                    setFormData({ ...formData, date_of_death: e.target.value })
                  }
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label="Profession"
                  value={formData.profession || ''}
                  onChange={(e) =>
                    setFormData({ ...formData, profession: e.target.value })
                  }
                />
              </Grid>
              <Grid item xs={12}>
                <Autocomplete
                  multiple
                  freeSolo
                  options={[]}
                  value={formData.nicknames || []}
                  onChange={(_, newValue) => {
                    setFormData({ ...formData, nicknames: newValue });
                  }}
                  renderTags={(value, getTagProps) =>
                    value.map((option, index) => (
                      <Chip
                        label={option}
                        {...getTagProps({ index })}
                        key={index}
                      />
                    ))
                  }
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      label="Nicknames"
                      placeholder="Type a nickname and press Enter"
                      helperText="Press Enter to add a nickname, click X to remove"
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <ParentAutocomplete
                  label="Father"
                  gender="M"
                  value={formData.father_id || null}
                  onChange={(value) =>
                    setFormData({
                      ...formData,
                      father_id: value || undefined,
                    })
                  }
                  initialParent={editingMember?.father || null}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <ParentAutocomplete
                  label="Mother"
                  gender="F"
                  value={formData.mother_id || null}
                  onChange={(value) =>
                    setFormData({
                      ...formData,
                      mother_id: value || undefined,
                    })
                  }
                  initialParent={editingMember?.mother || null}
                />
              </Grid>


              {/* Parents Display Section */}
              {editingMember && (editingMember.father || editingMember.mother) && (
                <Grid item xs={12}>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Parents
                  </Typography>
                  <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
                    {editingMember.father && (
                      <Paper
                        sx={{
                          p: 2,
                          display: 'flex',
                          alignItems: 'center',
                          gap: 2,
                          flex: '1 1 300px',
                          border: '1px solid #00BCD4',
                          cursor: 'pointer',
                          '&:hover': { boxShadow: 3, bgcolor: 'action.hover' },
                        }}
                        onClick={() => handleOpenRelatedMember(editingMember.father!.member_id)}
                      >
                        <Avatar
                          src={getMemberPictureUrl(editingMember.father.member_id, editingMember.father.picture) || undefined}
                          sx={{ width: 50, height: 50, bgcolor: '#00BCD4' }}
                        >
                          {editingMember.father.name?.[0] || '?'}
                        </Avatar>
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="body1" fontWeight="bold">
                            {editingMember.father.name}
                          </Typography>
                        </Box>
                      </Paper>
                    )}
                    {editingMember.mother && (
                      <Paper
                        sx={{
                          p: 2,
                          display: 'flex',
                          alignItems: 'center',
                          gap: 2,
                          flex: '1 1 300px',
                          border: '1px solid #E91E63',
                          cursor: 'pointer',
                          '&:hover': { boxShadow: 3, bgcolor: 'action.hover' },
                        }}
                        onClick={() => handleOpenRelatedMember(editingMember.mother!.member_id)}
                      >
                        <Avatar
                          src={getMemberPictureUrl(editingMember.mother.member_id, editingMember.mother.picture) || undefined}
                          sx={{ width: 50, height: 50, bgcolor: '#E91E63' }}
                        >
                          {editingMember.mother.name?.[0] || '?'}
                        </Avatar>
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="body1" fontWeight="bold">
                            {editingMember.mother.name}
                          </Typography>
                        </Box>
                      </Paper>
                    )}
                  </Box>
                </Grid>
              )}

              {/* Spouses Section */}
              {editingMember && (
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
                    <Typography variant="h6">
                      Spouses {editingMember.spouses && editingMember.spouses.length > 0 && `(${editingMember.spouses.length})`}
                    </Typography>
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<Add />}
                      onClick={() => setOpenAddSpouseDialog(true)}
                    >
                      Add Spouse
                    </Button>
                  </Box>
                  <Grid container spacing={2}>
                    {editingMember.spouses && editingMember.spouses.length > 0 ? (
                      editingMember.spouses.map((spouse) => (
                        <Grid item xs={12} md={6} key={spouse.member_id}>
                          <SpouseCard
                            spouse={spouse}
                            currentMemberId={editingMember.member_id}
                            onUpdate={async () => {
                              performSearch(searchQuery);
                              // Refetch member data after spouse update/delete
                              try {
                                const updated = await membersApi.getMember(editingMember.member_id);
                                setEditingMember(updated);
                              } catch (error) {
                                console.error('refresh member after spouse update:', error);
                              }
                            }}
                            editable={true}
                            onMemberClick={() => handleOpenRelatedMember(spouse.member_id)}
                          />
                        </Grid>
                      ))
                    ) : (
                      <Grid item xs={12}>
                        <Paper sx={{ p: 2, textAlign: 'center', bgcolor: 'action.hover' }}>
                          <Typography variant="body2" color="text.secondary">
                            No spouses added yet. Click "Add Spouse" to create a relationship.
                          </Typography>
                        </Paper>
                      </Grid>
                    )}
                  </Grid>
                </Grid>
              )}

              {/* Children Section */}
              {editingMember && editingMember.children && editingMember.children.length > 0 && (
                <Grid item xs={12}>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Children ({editingMember.children.length})
                  </Typography>
                  <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2 }}>
                    {editingMember.children.map((child) => (
                      <Paper
                        key={child.member_id}
                        sx={{
                          p: 2,
                          display: 'flex',
                          alignItems: 'center',
                          gap: 2,
                          border: '2px solid',
                          borderColor: '#9C27B0', // Purple border for children
                          cursor: 'pointer',
                          '&:hover': { boxShadow: 3, bgcolor: 'action.hover' },
                        }}
                        onClick={() => handleOpenRelatedMember(child.member_id)}
                      >
                        <Avatar
                          src={getMemberPictureUrl(child.member_id, child.picture) || undefined}
                          sx={{ width: 50, height: 50, bgcolor: '#9C27B0' }}
                        >
                          {child.name?.[0] || '?'}
                        </Avatar>
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="body1" fontWeight="bold">
                            {child.name}
                          </Typography>
                        </Box>
                      </Paper>
                    ))}
                  </Box>
                </Grid>
              )}

              {/* Siblings Section */}
              {editingMember && editingMember.siblings && editingMember.siblings.length > 0 && (
                <Grid item xs={12}>
                  <Typography variant="h6" sx={{ mb: 1 }}>
                    Siblings ({editingMember.siblings.length})
                  </Typography>
                  <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2 }}>
                    {editingMember.siblings.map((sibling) => (
                      <Paper
                        key={sibling.member_id}
                        sx={{
                          p: 2,
                          display: 'flex',
                          alignItems: 'center',
                          gap: 2,
                          border: '1px solid',
                          borderColor: 'warning.main',
                          cursor: 'pointer',
                          '&:hover': { boxShadow: 3, bgcolor: 'action.hover' },
                        }}
                        onClick={() => handleOpenRelatedMember(sibling.member_id)}
                      >
                        <Avatar
                          src={getMemberPictureUrl(sibling.member_id, sibling.picture) || undefined}
                          sx={{
                            width: 50,
                            height: 50,
                            bgcolor: '#FF9800',
                          }}
                        >
                          {sibling.name?.[0] || '?'}
                        </Avatar>
                        <Box sx={{ flex: 1 }}>
                          <Typography variant="body1" fontWeight="bold">
                            {sibling.name}
                          </Typography>
                        </Box>
                      </Paper>
                    ))}
                  </Box>
                </Grid>
              )}
            </Grid>
            )}

            {/* Change History Tab (Super Admin Only) */}
            {editingMember && hasRole(Roles.SUPER_ADMIN) && dialogTab === 1 && (
              <Box sx={{ mt: 2 }}>
                {memberHistory.length === 0 ? (
                  <Box sx={{ textAlign: 'center', py: 4 }}>
                    <Typography variant="body2" color="text.secondary">
                      No change history available for this member
                    </Typography>
                  </Box>
                ) : (
                  <>
                    <TableContainer>
                      <Table>
                        <TableHead>
                          <TableRow>
                            <TableCell>Change Type</TableCell>
                            <TableCell>User</TableCell>
                            <TableCell>Date</TableCell>
                            <TableCell>Version</TableCell>
                          </TableRow>
                        </TableHead>
                        <TableBody>
                          {memberHistory.slice(0, displayedHistoryCount).map((change) => (
                            <TableRow
                              key={change.history_id}
                              hover
                              sx={{ cursor: 'pointer' }}
                              onClick={() => handleViewDiff(change)}
                            >
                              <TableCell>
                                <Chip
                                  label={change.change_type}
                                  size="small"
                                  color={getChangeTypeColor(change.change_type)}
                                />
                              </TableCell>
                              <TableCell>
                                <Box>
                                  <Typography variant="body2">{change.user_full_name}</Typography>
                                  <Typography variant="caption" color="text.secondary">
                                    {change.user_email}
                                  </Typography>
                                </Box>
                              </TableCell>
                              <TableCell>
                                <Box>
                                  <Typography variant="body2">{formatDateTime(change.changed_at)}</Typography>
                                  <Typography variant="caption" color="text.secondary">
                                    {formatRelativeTime(change.changed_at)}
                                  </Typography>
                                </Box>
                              </TableCell>
                              <TableCell>{change.member_version}</TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </TableContainer>
                    {memberHistory.length > displayedHistoryCount && (
                      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2 }}>
                        <Button
                          variant="outlined"
                          onClick={() => setDisplayedHistoryCount(prev => prev + 10)}
                        >
                          Load More ({memberHistory.length - displayedHistoryCount} remaining)
                        </Button>
                      </Box>
                    )}
                  </>
                )}
              </Box>
            )}
          </DialogContent>
          <DialogActions sx={{ px: 3, py: 2 }}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', width: '100%', alignItems: 'center' }}>
              <Box>
                {editingMember && isSuperAdmin && (
                  <Tooltip title="Delete this member permanently (Super Admin only)">
                    <IconButton
                      onClick={() => handleDelete(editingMember.member_id)}
                      color="error"
                      size="small"
                    >
                      <Delete />
                    </IconButton>
                  </Tooltip>
                )}
              </Box>
              <Box sx={{ display: 'flex', gap: 1 }}>
                <Button onClick={handleCloseDialog}>Cancel</Button>
                {(!editingMember || dialogTab === 0) && (
                  <Button onClick={handleSubmit} variant="contained">
                    {editingMember ? 'Update' : 'Create'}
                  </Button>
                )}
              </Box>
            </Box>
          </DialogActions>
        </Dialog>

        {/* Add Spouse Dialog */}
        {editingMember && (
          <AddSpouseDialog
            open={openAddSpouseDialog}
            onClose={() => setOpenAddSpouseDialog(false)}
            memberId={editingMember.member_id}
            memberName={editingMember.name}
            memberGender={editingMember.gender}
            onSuccess={() => {
              performSearch(searchQuery);
              // Refresh the editing member data
              const refreshMember = async () => {
                try {
                  const updated = await membersApi.getMember(editingMember.member_id);
                  setEditingMember(updated);
                } catch (error) {
                  console.error('refresh member:', error);
                }
              };
              refreshMember();
            }}
          />
        )}

        {/* History Diff Dialog */}
        <HistoryDiffDialog
          open={diffDialogOpen}
          onClose={handleCloseDiff}
          history={selectedHistory}
        />
      </Box>
    </Layout>
  );
};

export default MembersPage;
