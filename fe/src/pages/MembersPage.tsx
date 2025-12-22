import React, { useState, useEffect } from 'react';
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
} from '@mui/material';
import { Add, Edit, Delete, FilterAlt, Clear, Close } from '@mui/icons-material';
import { membersApi } from '../api';
import { Member, MemberSearchQuery, CreateMemberRequest, UpdateMemberRequest } from '../types';
import { formatDate, getMemberPictureUrl } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import MemberPhotoUpload from '../components/MemberPhotoUpload';
import ParentAutocomplete from '../components/ParentAutocomplete';
import SpouseCard from '../components/SpouseCard';
import AddSpouseDialog from '../components/AddSpouseDialog';

const PAGE_SIZE = 10;

const MembersPage: React.FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [openAddSpouseDialog, setOpenAddSpouseDialog] = useState(false);
  const [editingMember, setEditingMember] = useState<Member | null>(null);
  const [memberChildren, setMemberChildren] = useState<Member[]>([]);
  const [originalFormData, setOriginalFormData] = useState<CreateMemberRequest | null>(null);
  const [formData, setFormData] = useState<CreateMemberRequest>({
    arabic_name: '',
    english_name: '',
    gender: 'M',
  });
  const tableRef = React.useRef<HTMLDivElement>(null);

  // Initialize search query from URL params
  const [searchQuery, setSearchQuery] = useState<MemberSearchQuery>(() => {
    const params: MemberSearchQuery = {};
    const name = searchParams.get('name');
    const gender = searchParams.get('gender');
    const married = searchParams.get('married');

    if (name) params.name = name;
    if (gender) params.gender = gender;
    if (married) params.married = Number(married);

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
        cursor: loadMore ? nextCursor || undefined : undefined,
      };

      const response = await membersApi.searchMembers(searchParams);

      if (loadMore) {
        setMembers(prev => [...prev, ...(response.members || [])]);
      } else {
        // Set new members after loading completes for smooth transition
        setMembers(response.members || []);
      }

      setNextCursor(response.next_cursor || null);
      setHasMore(!!response.next_cursor);
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

  const handleClearFilters = () => {
    setSearchQuery({});
  };

  const handleClearFilter = (filterKey: keyof MemberSearchQuery) => {
    const newQuery = { ...searchQuery };
    delete newQuery[filterKey];
    setSearchQuery(newQuery);
  };

  const loadMemberChildren = async (memberId: number) => {
    try {
      const children = await membersApi.getChildren(memberId);
      setMemberChildren(children || []);
    } catch (error) {
      console.error('Failed to load children:', error);
      setMemberChildren([]);
    }
  };

  const handleOpenDialog = async (member?: Member) => {
    if (member) {
      const initialData = {
        arabic_name: member.arabic_name,
        english_name: member.english_name,
        gender: member.gender,
        date_of_birth: member.date_of_birth || undefined,
        date_of_death: member.date_of_death || undefined,
        father_id: member.father_id || undefined,
        mother_id: member.mother_id || undefined,
        nicknames: member.nicknames || [],
        profession: member.profession || undefined,
      };
      setEditingMember(member);
      setFormData(initialData);
      setOriginalFormData(initialData);

      // Load children for this member
      await loadMemberChildren(member.member_id);
    } else {
      setEditingMember(null);
      setOriginalFormData(null);
      setMemberChildren([]);
      setFormData({
        arabic_name: '',
        english_name: '',
        gender: 'M',
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setEditingMember(null);
    setOriginalFormData(null);
    setMemberChildren([]);
  };

  const handleOpenRelatedMember = async (memberId: number) => {
    try {
      const member = await membersApi.getMember(memberId);
      handleCloseDialog(); // Close current dialog first
      setTimeout(() => handleOpenDialog(member), 100); // Open new dialog with slight delay
    } catch (error) {
      console.error('Failed to load member:', error);
      alert('Failed to load member details');
    }
  };

  // Check if form has changes
  const hasChanges = () => {
    if (!editingMember || !originalFormData) return true; // Allow create

    // Compare all fields
    if (formData.arabic_name !== originalFormData.arabic_name) return true;
    if (formData.english_name !== originalFormData.english_name) return true;
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
    try {
      if (editingMember) {
        const updateData: UpdateMemberRequest = {
          ...formData,
          version: editingMember.version,
        };
        await membersApi.updateMember(editingMember.member_id, updateData);
      } else {
        await membersApi.createMember(formData);
      }
      handleCloseDialog();
      performSearch(searchQuery); // Refresh list
    } catch (error) {
      console.error('Failed to save member:', error);
    }
  };

  const handleDelete = async (memberId: number) => {
    if (confirm('Are you sure you want to delete this member?')) {
      try {
        await membersApi.deleteMember(memberId);
        performSearch(searchQuery); // Refresh list
      } catch (error) {
        console.error('Failed to delete member:', error);
      }
    }
  };

  const handlePhotoChange = (memberId: number, pictureUrl: string | null) => {
    // Update the member in the list
    setMembers(prevMembers =>
      prevMembers.map(member =>
        member.member_id === memberId
          ? { ...member, picture: pictureUrl, version: member.version + 1 }
          : member
      )
    );

    // If editing this member, update the dialog state
    if (editingMember && editingMember.member_id === memberId) {
      setEditingMember(prev => prev ? {
        ...prev,
        picture: pictureUrl,
        version: prev.version + 1
      } : null);
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
                    setSearchQuery({ ...searchQuery, gender: e.target.value || undefined })
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
                      married: e.target.value === '' ? undefined : Number(e.target.value),
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
                <TableCell>Avatar</TableCell>
                <TableCell>Arabic Name</TableCell>
                <TableCell>English Name</TableCell>
                <TableCell>Gender</TableCell>
                <TableCell>Date of Birth</TableCell>
                <TableCell>Married</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {(!members || members.length === 0) && !loading && (
                <TableRow>
                  <TableCell colSpan={7} align="center" sx={{ py: 4, color: 'text.secondary' }}>
                    {searchQuery.name || searchQuery.gender || searchQuery.married !== undefined
                      ? 'No members found matching your filters'
                      : 'No members found'}
                  </TableCell>
                </TableRow>
              )}
              {loading && members.length === 0 && (
                <TableRow>
                  <TableCell colSpan={7} align="center" sx={{ py: 8 }}>
                    <CircularProgress />
                    <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                      Loading members...
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
              {members && members.map((member) => (
                <TableRow key={member.member_id}>
                  <TableCell>
                    <Avatar
                      src={getMemberPictureUrl(member.member_id, member.picture, member.version) || undefined}
                      sx={{
                        width: 50,
                        height: 50,
                        bgcolor: member.gender === 'M' ? '#00BCD4' : member.gender === 'F' ? '#E91E63' : '#9E9E9E'
                      }}
                    >
                      {member.english_name[0]}
                    </Avatar>
                  </TableCell>
                  <TableCell>{member.arabic_name}</TableCell>
                  <TableCell>{member.english_name}</TableCell>
                  <TableCell>
                    {member.gender === 'M' ? 'Male' : 'Female'}
                  </TableCell>
                  <TableCell>{formatDate(member.date_of_birth)}</TableCell>
                  <TableCell>
                    {member.is_married ? (
                      <Chip label="Yes" color="primary" size="small" />
                    ) : (
                      <Chip label="No" size="small" />
                    )}
                  </TableCell>
                  <TableCell>
                    <IconButton
                      size="small"
                      onClick={() => handleOpenDialog(member)}
                    >
                      <Edit />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => handleDelete(member.member_id)}
                      color="error"
                    >
                      <Delete />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>

        {/* Load More Button */}
        {hasMore && members && members.length > 0 && (
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 2, mb: 2 }}>
            <Button
              variant="outlined"
              onClick={handleLoadMore}
              disabled={loadingMore}
              startIcon={loadingMore ? <CircularProgress size={20} /> : null}
            >
              {loadingMore ? 'Loading...' : `Load More (${members.length} shown)`}
            </Button>
          </Box>
        )}

        {/* Create/Edit Dialog */}
        <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="md" fullWidth>
          <DialogTitle>
            {editingMember ? 'Edit Member' : 'Add New Member'}
          </DialogTitle>
          <DialogContent>
            <Grid container spacing={2} sx={{ mt: 1 }}>
              {editingMember && (
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', py: 2 }}>
                    <MemberPhotoUpload
                      memberId={editingMember.member_id}
                      currentPhoto={editingMember.picture}
                      memberName={editingMember.english_name}
                      gender={editingMember.gender}
                      version={editingMember.version}
                      onPhotoChange={handlePhotoChange}
                      size={120}
                      showName
                    />
                    {(editingMember.english_full_name || editingMember.arabic_full_name) && (
                      <Box sx={{ mt: 2, textAlign: 'center' }}>
                        <Typography variant="caption" color="text.secondary" gutterBottom>
                          Full Name
                        </Typography>
                        {editingMember.english_full_name && (
                          <Typography variant="body2" fontWeight="medium">
                            {editingMember.english_full_name}
                          </Typography>
                        )}
                        {editingMember.arabic_full_name && (
                          <Typography variant="body2" fontWeight="medium" dir="rtl">
                            {editingMember.arabic_full_name}
                          </Typography>
                        )}
                      </Box>
                    )}
                  </Box>
                </Grid>
              )}
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  required
                  label="Arabic Name"
                  value={formData.arabic_name}
                  onChange={(e) =>
                    setFormData({ ...formData, arabic_name: e.target.value })
                  }
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  required
                  label="English Name"
                  value={formData.english_name}
                  onChange={(e) =>
                    setFormData({ ...formData, english_name: e.target.value })
                  }
                />
              </Grid>
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
                />
              </Grid>

              {/* Parents Display Section */}
              {editingMember && (editingMember.father || editingMember.mother) && (
                <Grid item xs={12}>
                  <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
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
                          {editingMember.father.english_name[0]}
                        </Avatar>
                        <Box sx={{ flex: 1 }}>
                          <Chip label="Father" size="small" color="info" sx={{ mb: 0.5 }} />
                          <Typography variant="body1" fontWeight="bold">
                            {editingMember.father.english_name}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {editingMember.father.arabic_name}
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
                          {editingMember.mother.english_name[0]}
                        </Avatar>
                        <Box sx={{ flex: 1 }}>
                          <Chip label="Mother" size="small" color="secondary" sx={{ mb: 0.5 }} />
                          <Typography variant="body1" fontWeight="bold">
                            {editingMember.mother.english_name}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {editingMember.mother.arabic_name}
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
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mt: 2, mb: 1 }}>
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
                  {editingMember.spouses && editingMember.spouses.length > 0 ? (
                    editingMember.spouses.map((spouse) => (
                      <SpouseCard
                        key={spouse.member_id}
                        spouse={spouse}
                        currentMemberId={editingMember.member_id}
                        onUpdate={() => performSearch(searchQuery)}
                        editable={true}
                        onMemberClick={() => handleOpenRelatedMember(spouse.member_id)}
                      />
                    ))
                  ) : (
                    <Typography variant="body2" color="text.secondary" sx={{ py: 2, textAlign: 'center' }}>
                      No spouses added yet. Click "Add Spouse" to create a relationship.
                    </Typography>
                  )}
                </Grid>
              )}

              {/* Children Section */}
              {editingMember && (
                <Grid item xs={12}>
                  <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
                    Children {memberChildren && memberChildren.length > 0 && `(${memberChildren.length})`}
                  </Typography>
                  {memberChildren && memberChildren.length > 0 ? (
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                      {memberChildren.map((child) => (
                        <Paper
                          key={child.member_id}
                          sx={{
                            p: 2,
                            display: 'flex',
                            alignItems: 'center',
                            gap: 2,
                            border: '1px solid',
                            borderColor: 'divider',
                            cursor: 'pointer',
                            '&:hover': { boxShadow: 3, bgcolor: 'action.hover' },
                          }}
                          onClick={() => handleOpenRelatedMember(child.member_id)}
                        >
                          <Avatar
                            src={getMemberPictureUrl(child.member_id, child.picture, child.version) || undefined}
                            sx={{
                              width: 50,
                              height: 50,
                              bgcolor: child.gender === 'M' ? '#00BCD4' : child.gender === 'F' ? '#E91E63' : '#9E9E9E',
                            }}
                          >
                            {child.english_name[0]}
                          </Avatar>
                          <Box sx={{ flex: 1 }}>
                            <Chip
                              label={child.gender === 'M' ? 'Son' : child.gender === 'F' ? 'Daughter' : 'Child'}
                              size="small"
                              sx={{ mb: 0.5 }}
                            />
                            <Typography variant="body1" fontWeight="bold">
                              {child.english_name}
                            </Typography>
                            <Typography variant="body2" color="text.secondary">
                              {child.arabic_name}
                            </Typography>
                          </Box>
                          {child.age && (
                            <Chip label={`Age: ${child.age}`} size="small" variant="outlined" />
                          )}
                        </Paper>
                      ))}
                    </Box>
                  ) : (
                    <Typography variant="body2" color="text.secondary" sx={{ py: 2, textAlign: 'center' }}>
                      No children found.
                    </Typography>
                  )}
                </Grid>
              )}
            </Grid>
          </DialogContent>
          <DialogActions>
            <Button onClick={handleCloseDialog}>Cancel</Button>
            <Button onClick={handleSubmit} variant="contained">
              {editingMember ? 'Update' : 'Create'}
            </Button>
          </DialogActions>
        </Dialog>

        {/* Add Spouse Dialog */}
        {editingMember && (
          <AddSpouseDialog
            open={openAddSpouseDialog}
            onClose={() => setOpenAddSpouseDialog(false)}
            memberId={editingMember.member_id}
            memberName={editingMember.english_name}
            memberGender={editingMember.gender}
            onSuccess={() => {
              performSearch(searchQuery);
              // Refresh the editing member data
              const refreshMember = async () => {
                try {
                  const updated = await membersApi.getMember(editingMember.member_id);
                  setEditingMember(updated);
                } catch (error) {
                  console.error('Failed to refresh member:', error);
                }
              };
              refreshMember();
            }}
          />
        )}
      </Box>
    </Layout>
  );
};

export default MembersPage;
