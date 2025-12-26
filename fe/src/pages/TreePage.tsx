import React, { useEffect, useState, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  Box,
  Paper,
  Typography,
  ToggleButtonGroup,
  ToggleButton,
  Divider,
  Drawer,
  Button,
  Avatar,
  Chip,
  Alert,
  CircularProgress,
  TextField,
  Grid,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  InputAdornment,
  IconButton,
  TableContainer,
  Table,
  TableHead,
  TableRow,
  TableCell,
  TableBody,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Tabs,
  Tab,
  Autocomplete,
  Tooltip,
} from '@mui/material';
import {
  AccountTree,
  TableChart,
  Close,
  Refresh,
  FilterAlt,
  Clear,
  Add,
  Delete,
} from '@mui/icons-material';
// Animations removed for stability
import { useTranslation } from 'react-i18next';
import { enqueueSnackbar } from 'notistack';
import { treeApi, membersApi } from '../api';
import { MemberListItem, MemberSearchQuery, CreateMemberRequest, UpdateMemberRequest, HistoryRecord, Language } from '../types';
import { TreeNode, Member } from '../types';
import { getGenderColor, formatDate, formatDateOfBirth, getMemberPictureUrl, formatDateTime, formatRelativeTime, getChangeTypeColor, getLocalizedLanguageName } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import TreeVisualization from '../components/TreeVisualization';
import RelationFinder from '../components/RelationFinder';
import MemberPhotoUpload from '../components/MemberPhotoUpload';
import ParentAutocomplete from '../components/ParentAutocomplete';
import SpouseCard from '../components/SpouseCard';
import AddSpouseDialog from '../components/AddSpouseDialog';
import HistoryDiffDialog from '../components/HistoryDiffDialog';
import DirectionalButton from '../components/DirectionalButton';
import { useAuth } from '../contexts/AuthContext';
import { useLanguage } from '../contexts/LanguageContext';
import { Roles } from '../types';

type ViewMode = 'tree' | 'list' | 'relation';

const TreePage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { hasRole } = useAuth();
  const { getPreferredName, getAllNamesFormatted, languages } = useLanguage();
  const isSuperAdmin = hasRole(Roles.SUPER_ADMIN);
  const isRTL = i18n.dir() === 'rtl';
  const [searchParams, setSearchParams] = useSearchParams();

  // Initialize from URL params
  const initialViewMode = (searchParams.get('view') as ViewMode) || 'tree';
  const initialRootId = searchParams.get('root') ? parseInt(searchParams.get('root')!) : undefined;

  // View state
  const [viewMode, setViewMode] = useState<ViewMode>(initialViewMode);
  const [rootId, setRootId] = useState<number | undefined>(initialRootId);

  // Data state
  const [treeData, setTreeData] = useState<TreeNode | null>(null);
  const [listMembers, setListMembers] = useState<MemberListItem[]>([]);
  const [relationTree, setRelationTree] = useState<TreeNode | null>(null);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);

  // Search/filter state for list view - Initialize from URL params
  const [searchQuery, setSearchQuery] = useState<MemberSearchQuery>(() => {
    const params: MemberSearchQuery = {};
    const name = searchParams.get('name');
    const gender = searchParams.get('gender');
    const married = searchParams.get('married');

    if (name) params.name = name;
    if (gender && (gender === 'M' || gender === 'F')) params.gender = gender;
    if (married !== null) params.married = married === '1';

    return params;
  });

  // UI state
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedMember, setSelectedMember] = useState<Member | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [relationLoading, setRelationLoading] = useState(false);

  // Member management dialog state
  const [openDialog, setOpenDialog] = useState(false);
  const [openAddSpouseDialog, setOpenAddSpouseDialog] = useState(false);
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false);
  const [memberToDelete, setMemberToDelete] = useState<number | null>(null);
  const [deleting, setDeleting] = useState(false);
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

  // Ref for intersection observer
  const loadMoreRef = useRef<HTMLDivElement>(null);

  // Update URL params when state changes
  useEffect(() => {
    const params = new URLSearchParams();
    params.set('view', viewMode);
    if (rootId !== undefined) {
      params.set('root', rootId.toString());
    }
    // Add search query params for list view
    if (viewMode === 'list') {
      if (searchQuery.name) params.set('name', searchQuery.name);
      if (searchQuery.gender) params.set('gender', searchQuery.gender);
      if (searchQuery.married !== undefined) params.set('married', searchQuery.married ? '1' : '0');
    }
    setSearchParams(params, { replace: true });
  }, [viewMode, rootId, searchQuery, setSearchParams]);

  // Load data based on view mode
  useEffect(() => {
    if (viewMode === 'tree') {
      loadTree();
    } else if (viewMode === 'list') {
      loadListView();
    }
  }, [rootId, viewMode, searchQuery]);

  const loadTree = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await treeApi.getTree({ root: rootId, style: 'tree' });
      setTreeData(data);
    } catch (error) {

      setError(t('apiErrors.failedToLoadTree'));
    } finally {
      setLoading(false);
    }
  };

  const loadListView = async (loadMore: boolean = false) => {
    if (loadMore) {
      setLoadingMore(true);
    } else {
      setLoading(true);
      setListMembers([]); // Clear previous data on fresh load
      setNextCursor(null);
      setHasMore(false);
    }
    setError(null);
    try {
      // Use the members search API with cursor-based pagination
      const params: MemberSearchQuery = {
        ...searchQuery,
        limit: 10
      };
      if (loadMore && nextCursor) {
        params.cursor = nextCursor;
      }
      const response = await membersApi.searchMembers(params);

      if (loadMore) {
        setListMembers(prev => [...prev, ...(response.members || [])]);
      } else {
        setListMembers(response.members || []);
      }

      // Only set cursor if it exists and is not empty
      const validCursor = response.next_cursor && response.next_cursor.trim() !== '' ? response.next_cursor : null;
      setNextCursor(validCursor);
      setHasMore(!!validCursor);

    } catch (error) {

      setError(t('apiErrors.failedToLoadMemberList'));
      if (!loadMore) {
        setListMembers([]);
      }
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  };

  const handleClearFilters = () => {
    setSearchQuery({});
  };

  const handleClearFilter = (filterKey: keyof MemberSearchQuery) => {
    const newQuery = { ...searchQuery };
    delete newQuery[filterKey];
    setSearchQuery(newQuery);
  };

  const handleLoadMore = () => {
    loadListView(true);
  };

  // Infinite scroll observer for list view
  useEffect(() => {
    if (viewMode !== 'list' || !loadMoreRef.current || !hasMore || loadingMore) return;

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
  }, [viewMode, hasMore, loadingMore, nextCursor]);

  const handleFindRelation = async (member1Id: number, member2Id: number) => {
    setRelationLoading(true);
    setError(null);
    try {
      const data = await treeApi.getRelation({ member1: member1Id, member2: member2Id });
      setRelationTree(data);
      setViewMode('relation');
    } catch (error) {

      setError(t('apiErrors.noRelationFound'));
    } finally {
      setRelationLoading(false);
    }
  };

  const handleMemberClick = async (member: Member | MemberListItem) => {
    // For admins, open the member management dialog
    if (hasRole(Roles.ADMIN)) {
      await handleOpenMemberDialog(member);
    } else {
      // For non-admins, open the view-only drawer
      try {
        const fullMember = await membersApi.getMember(member.member_id);
        setSelectedMember(fullMember);
        setDrawerOpen(true);
      } catch (error) {
        console.error('Failed to load member:', error);
      }
    }
  };

  const handleSetRoot = (memberId: number) => {
    if (memberId === -1) {
      // Special value to reset root
      setRootId(undefined);
    } else {
      setRootId(memberId);
    }
    setViewMode('tree');
  };

  const handleResetRoot = () => {
    setRootId(undefined);
    if (viewMode === 'relation') {
      setViewMode('tree');
    }
  };

  const handleViewModeChange = (_: React.MouseEvent<HTMLElement>, newMode: ViewMode | null) => {
    if (newMode) {
      setViewMode(newMode);
      if (newMode === 'relation') {
        setRelationTree(null); // Clear previous relation
      }
    }
  };

  // Member management handlers
  const handleOpenMemberDialog = async (memberIdOrMember?: number | MemberListItem) => {
    setDialogTab(0); // Reset to first tab
    setMemberHistory([]); // Clear previous history
    setDisplayedHistoryCount(10); // Reset pagination

    if (memberIdOrMember) {
      try {
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
        if (isSuperAdmin) {
          try {
            const historyResponse = await membersApi.getMemberHistory(memberId);
            setMemberHistory(historyResponse.history || []);
          } catch (error) {
            console.error('Failed to load member history:', error);
          }
        }
      } catch (error) {
        console.error('Failed to load member details:', error);
        enqueueSnackbar(t('apiErrors.failedToLoadMemberDetails'), { variant: 'error' });
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

  const handleCloseMemberDialog = () => {
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
    handleCloseMemberDialog(); // Close current dialog first
    setTimeout(() => handleOpenMemberDialog(memberId), 100); // Open new dialog with slight delay
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

  const handleSubmitMember = async () => {
    // Check if there are any changes for update operations
    if (editingMember && !hasChanges()) {
      handleCloseMemberDialog();
      return;
    }

    // Validate that all active languages have names
    const activeLanguages = Array.isArray(languages) ? languages.filter((lang: Language) => lang.is_active) : [];
    const missingLanguages = activeLanguages.filter(
      (lang: Language) => !formData.names[lang.language_code] || formData.names[lang.language_code].trim() === ''
    );

    if (missingLanguages.length > 0) {
      const missingNames = missingLanguages.map((lang: Language) => lang.language_name).join(', ');
      enqueueSnackbar(t('validation.provideNamesForLanguages', { missing: missingNames }), { variant: 'warning' });
      return;
    }

    try {
      if (editingMember) {
        const updateData: UpdateMemberRequest = {
          ...formData,
          version: editingMember.version,
        };
        await membersApi.updateMember(editingMember.member_id, updateData);

        // Refetch member data after update to get latest computed fields
        const updatedMember = await membersApi.getMember(editingMember.member_id);
        setEditingMember(updatedMember);
      } else {
        await membersApi.createMember(formData);
      }
      handleCloseMemberDialog();

      // Refresh list view if in list mode
      if (viewMode === 'list') {
        loadListView();
      } else if (viewMode === 'tree') {
        loadTree();
      }
    } catch (error: any) {
      const errorMsg = error?.response?.data?.error || 'Failed to save member';
      enqueueSnackbar(errorMsg, { variant: 'error' });
      console.error('Failed to save member:', error);
    }
  };

  const handleDeleteClick = (memberId: number) => {
    setMemberToDelete(memberId);
    setOpenDeleteDialog(true);
  };

  const handleDeleteConfirm = async () => {
    if (!memberToDelete) return;

    setDeleting(true);
    try {
      await membersApi.deleteMember(memberToDelete);
      setOpenDeleteDialog(false);
      setMemberToDelete(null);
      handleCloseMemberDialog(); // Close dialog after delete

      // Refresh list view if in list mode
      if (viewMode === 'list') {
        loadListView();
      } else if (viewMode === 'tree') {
        loadTree();
      }
    } catch (error: any) {
      console.error('Failed to delete member:', error);
      const errorMessage = error?.response?.data?.error || t('member.failedToDeleteMember');
      enqueueSnackbar(errorMessage, { variant: 'error' });
    } finally {
      setDeleting(false);
    }
  };

  const handleDeleteCancel = () => {
    setOpenDeleteDialog(false);
    setMemberToDelete(null);
  };

  const handlePhotoChange = async (memberId: number, pictureUrl: string | null) => {
    // Update the member picture in the list
    setListMembers(prevMembers =>
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
        console.error('Failed to refresh member after photo change:', error);
      }
    }
  };

  return (
    <Layout>
      <Box sx={{ width: '100%' }}>
        {/* Page Header */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3, width: '100%' }}>
          <Typography variant="h4" gutterBottom>
            {t('tree.title')}
          </Typography>
          {rootId && (
            <Button startIcon={<Refresh />} onClick={handleResetRoot} variant="outlined" sx={{ flexShrink: 0 }}>
              {t('tree.resetToDefaultRoot')}
            </Button>
          )}
        </Box>

        <Divider sx={{ mb: 3 }} />

        {/* Section 1: View Mode Selection */}
        <Paper sx={{ p: 2, mb: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 2, mb: 2 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, flexWrap: 'wrap' }}>
              <Typography variant="h6">
                {t('tree.viewMode')}
              </Typography>
              {hasRole(Roles.ADMIN) && (
                <Button
                  variant="contained"
                  color="primary"
                  startIcon={<Add />}
                  onClick={() => handleOpenMemberDialog()}
                  sx={{
                    fontWeight: 'bold',
                    boxShadow: 2,
                    '&:hover': {
                      boxShadow: 4,
                    }
                  }}
                >
                  {t('member.addMember')}
                </Button>
              )}
            </Box>
          </Box>
          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'center' }}>
            <ToggleButtonGroup
              value={viewMode}
              exclusive
              onChange={handleViewModeChange}
              aria-label="view mode"
              sx={{
                '& .MuiToggleButton-root': {
                  border: '1px solid rgba(0, 0, 0, 0.12)',
                  '&:first-of-type': {
                    borderStartStartRadius: 4,
                    borderEndStartRadius: 4,
                    borderStartEndRadius: 0,
                    borderEndEndRadius: 0,
                  },
                  '&:last-of-type': {
                    borderStartStartRadius: 0,
                    borderEndStartRadius: 0,
                    borderStartEndRadius: 4,
                    borderEndEndRadius: 4,
                  },
                  '&:not(:first-of-type):not(:last-of-type)': {
                    borderRadius: 0,
                  },
                  '&:not(:first-of-type)': {
                    marginInlineStart: '-1px',
                  }
                }
              }}
            >
              <ToggleButton value="tree" aria-label="tree view">
                <AccountTree sx={{ marginInlineEnd: 1 }} />
                {t('tree.treeDiagram')}
              </ToggleButton>
              <ToggleButton value="list" aria-label="list view">
                <TableChart sx={{ marginInlineEnd: 1 }} />
                {t('tree.tableView')}
              </ToggleButton>
              <ToggleButton value="relation" aria-label="relation view">
                <AccountTree sx={{ marginInlineEnd: 1 }} />
                {t('tree.findRelation')}
              </ToggleButton>
            </ToggleButtonGroup>
          </Box>
        </Paper>

        {/* Section 2: Relation Finder (visible when in relation mode) */}
        {viewMode === 'relation' && (
          <Box sx={{ mb: 3 }}>
            <RelationFinder
              onFindRelation={handleFindRelation}
              loading={relationLoading}
            />
          </Box>
        )}

        {/* Error Display */}
        {error && (
          <Alert
            severity="error"
            sx={{
              mb: 3,
              textAlign: isRTL ? 'right' : 'left',
              '& .MuiAlert-icon': {
                marginInlineEnd: 1.5,
                marginInlineStart: 0,
              }
            }}
            onClose={() => setError(null)}
          >
            {error}
          </Alert>
        )}

        {/* Section 4: Main Content Area */}
      {loading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', py: 10 }}>
            <CircularProgress />
          </Box>
        ) : (
          <>
            {/* Tree View */}
            {viewMode === 'tree' && treeData && (
                <Box>
                  <Typography variant="h6" gutterBottom>
                    {t('tree.hierarchicalTreeView')}
                  </Typography>
                  <TreeVisualization
                    data={treeData}
                    onNodeClick={handleMemberClick}
                    onSetRoot={handleSetRoot}
                    currentRootId={rootId}
                  />
                </Box>
            )}

            {/* List View */}
            {viewMode === 'list' && (
              <Box>
                {/* Search Filters */}
                <Paper sx={{ p: 2, mb: 3 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', mb: 2, flexWrap: 'wrap', gap: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', flexGrow: 1 }}>
                      <FilterAlt sx={{ marginInlineEnd: 1, color: 'text.secondary' }} />
                      <Typography variant="h6">
                        {t('tree.searchFilters')} {!searchQuery.name && !searchQuery.gender && searchQuery.married === undefined && t('tree.showingAllMembers')}
                      </Typography>
                    </Box>
                    {(searchQuery.name || searchQuery.gender || searchQuery.married !== undefined) && (
                      <Button
                        size="small"
                        startIcon={<Clear />}
                        onClick={handleClearFilters}
                        color="secondary"
                      >
                        {t('tree.clearFilters')}
                      </Button>
                    )}
                  </Box>
                  <Grid container spacing={2}>
                    <Grid item xs={12} sm={6} md={4}>
                      <TextField
                        fullWidth
                        label={t('member.name')}
                        placeholder={t('member.searchPlaceholder')}
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
                        <InputLabel>{t('member.gender')}</InputLabel>
                        <Select
                          value={searchQuery.gender || ''}
                          label={t('member.gender')}
                          onChange={(e) => {
                            const value = e.target.value;
                            setSearchQuery({
                              ...searchQuery,
                              gender: value === 'M' || value === 'F' ? value : undefined
                            });
                          }}
                          endAdornment={
                            searchQuery.gender && (
                              <InputAdornment position="end" sx={{ marginInlineEnd: 3 }}>
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
                          <MenuItem value="">{t('common.all')}</MenuItem>
                          <MenuItem value="M">{t('member.male')}</MenuItem>
                          <MenuItem value="F">{t('member.female')}</MenuItem>
                        </Select>
                      </FormControl>
                    </Grid>
                    <Grid item xs={12} sm={6} md={4}>
                      <FormControl fullWidth>
                        <InputLabel>{t('member.married')}</InputLabel>
                        <Select
                          value={searchQuery.married === undefined ? '' : searchQuery.married ? 1 : 0}
                          label={t('member.married')}
                          onChange={(e) =>
                            setSearchQuery({
                              ...searchQuery,
                              married: e.target.value === '' ? undefined : e.target.value === 1,
                            })
                          }
                          endAdornment={
                            searchQuery.married !== undefined && (
                              <InputAdornment position="end" sx={{ marginInlineEnd: 3 }}>
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
                          <MenuItem value="">{t('common.all')}</MenuItem>
                          <MenuItem value={1}>{t('common.yes')}</MenuItem>
                          <MenuItem value={0}>{t('common.no')}</MenuItem>
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
                  sx={{
                    position: 'relative',
                    minHeight: '400px',
                    transition: 'opacity 0.3s ease-in-out',
                    opacity: loading && listMembers.length === 0 ? 0.6 : 1,
                  }}
                >
                  {loading && listMembers.length > 0 && (
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
                        <TableCell className="table-header-cell">{t('member.avatar')}</TableCell>
                        <TableCell className="table-header-cell">{t('member.name')}</TableCell>
                        <TableCell className="table-header-cell">{t('member.gender')}</TableCell>
                        <TableCell className="table-header-cell numeric-cell">{t('member.dateOfBirth')}</TableCell>
                        <TableCell className="table-header-cell">{t('member.married')}</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {(!listMembers || listMembers.length === 0) && !loading && (
                        <TableRow>
                          <TableCell colSpan={5} align="center" sx={{ py: 4, color: 'text.secondary' }}>
                            {searchQuery.name || searchQuery.gender || searchQuery.married !== undefined
                              ? t('tree.noMembersMatchingFilters')
                              : t('member.noMembers')}
                          </TableCell>
                        </TableRow>
                      )}
                      {loading && listMembers.length === 0 && (
                        <TableRow>
                          <TableCell colSpan={5} align="center" sx={{ py: 8 }}>
                            <CircularProgress />
                            <Typography variant="body2" color="text.secondary" sx={{ mt: 2 }}>
                              {t('tree.loadingMembers')}
                            </Typography>
                          </TableCell>
                        </TableRow>
                      )}
                      {listMembers && listMembers.map((member) => (
                        <TableRow
                          key={member.member_id}
                          hover
                          className={member.gender === 'M' ? 'table-row-male' : 'table-row-female'}
                          sx={{ cursor: 'pointer' }}
                          onClick={() => handleMemberClick(member)}
                        >
                          <TableCell>
                            <Avatar
                              src={getMemberPictureUrl(member.member_id, member.picture) || undefined}
                              className={member.gender === 'M' ? 'avatar-ring-male' : 'avatar-ring-female'}
                              sx={{
                                width: 50,
                                height: 50,
                                bgcolor: member.gender === 'M' ? '#4299e1' : member.gender === 'F' ? '#ed64a6' : '#9E9E9E'
                              }}
                            >
                              {member.name?.[0] || '?'}
                            </Avatar>
                          </TableCell>
                          <TableCell className="mixed-content-cell">{member.name}</TableCell>
                          <TableCell>
                            <Chip
                              label={member.gender === 'M' ? t('member.male') : t('member.female')}
                              size="small"
                              className={member.gender === 'M' ? 'gender-male-bg' : 'gender-female-bg'}
                            />
                          </TableCell>
                          <TableCell className="numeric-cell">{formatDateOfBirth(member.date_of_birth, isSuperAdmin)}</TableCell>
                          <TableCell>
                            {member.is_married ? (
                              <Chip label={t('common.yes')} size="small" color="secondary" variant="filled" className="enhanced-chip" />
                            ) : (
                              <Chip label={t('common.no')} size="small" variant="outlined" className="enhanced-chip" />
                            )}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>

                {/* Load More Sentinel */}
                {hasMore && listMembers && listMembers.length > 0 && (
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
                          {t('tree.loadingMoreMembers')}
                        </Typography>
                      </Box>
                    ) : (
                      <Button
                        variant="outlined"
                        onClick={handleLoadMore}
                        size="large"
                      >
                        {t('tree.loadMoreMembers')}
                      </Button>
                    )}
                  </Box>
                )}
              </Box>
            )}

            {/* Relation View */}
            {viewMode === 'relation' && relationTree && (
                <Box>
                  <Typography variant="h6" gutterBottom>
                    {t('tree.relationPath')} {t('tree.relationPathDescription')}
                  </Typography>
                  <TreeVisualization
                    data={relationTree}
                    onNodeClick={handleMemberClick}
                    onSetRoot={handleSetRoot}
                    currentRootId={rootId}
                  />
                </Box>
            )}

            {/* Empty State */}
            {viewMode === 'relation' && !relationTree && !relationLoading && (
              <Paper sx={{ p: 5, textAlign: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  {t('tree.selectTwoMembers')}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {t('tree.relationFinderDescription')}
                </Typography>
              </Paper>
            )}
          </>
        )}
      </Box>

      {/* Member Details Drawer */}
      <Drawer
        anchor={isRTL ? 'left' : 'right'}
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        SlideProps={{
          direction: isRTL ? 'right' : 'left',
        }}
      >
        <Box sx={{ width: { xs: '100vw', sm: 450 }, maxWidth: 450, p: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
            <Typography variant="h6">{t('member.memberDetails')}</Typography>
            <Button onClick={() => setDrawerOpen(false)} startIcon={<Close />}>
              {t('common.close')}
            </Button>
          </Box>

          {selectedMember && (
            <Box>
              <Box
                sx={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                  textAlign: 'center',
                  mb: 2
                }}
              >
                <Avatar
                  src={
                    getMemberPictureUrl(selectedMember.member_id, selectedMember.picture) || undefined
                  }
                  sx={{
                    width: 120,
                    height: 120,
                    mb: 2,
                    bgcolor: getGenderColor(selectedMember.gender),
                  }}
                >
                  {getPreferredName(selectedMember)[0] || '?'}
                </Avatar>

                <Typography variant="h6" gutterBottom>
                  {getPreferredName(selectedMember)}
                </Typography>
                {/* Display full names if different from preferred name */}
                {selectedMember.full_names && Object.keys(selectedMember.full_names).length > 0 && (
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    {getAllNamesFormatted({ names: selectedMember.full_names })}
                  </Typography>
                )}
              </Box>

              <Divider sx={{ my: 2 }} />

              {/* Details Grid */}
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                <Box>
                <Typography variant="subtitle2" color="text.secondary">
                  {t('member.gender')}
                </Typography>
                  <Chip
                    label={
                      selectedMember.gender === 'M'
                        ? t('member.male')
                        : selectedMember.gender === 'F'
                        ? t('member.female')
                        : t('general.other')
                    }
                    size="small"
                    sx={{ bgcolor: getGenderColor(selectedMember.gender), color: 'white' }}
                  />
                </Box>

                {selectedMember.date_of_birth && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      {t('member.dateOfBirth')}
                    </Typography>
                    <Typography variant="body1">
                      {formatDateOfBirth(selectedMember.date_of_birth, isSuperAdmin)}
                    </Typography>
                  </Box>
                )}

                {selectedMember.date_of_death && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      {t('member.dateOfDeath')}
                    </Typography>
                    <Typography variant="body1">{formatDate(selectedMember.date_of_death)}</Typography>
                  </Box>
                )}

                {selectedMember.age && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      {t('member.age')}
                    </Typography>
                    <Typography variant="body1">{selectedMember.age} {t('member.years')}</Typography>
                  </Box>
                )}

                {selectedMember.generation_level !== undefined && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      {t('member.generationLevel')}
                    </Typography>
                    <Typography variant="body1">{selectedMember.generation_level}</Typography>
                  </Box>
                )}

                {selectedMember.profession && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      {t('member.profession')}
                    </Typography>
                    <Typography variant="body1">{selectedMember.profession}</Typography>
                  </Box>
                )}

                {selectedMember.nicknames && selectedMember.nicknames.length > 0 && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      {t('member.nicknames')}
                    </Typography>
                    <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                      {selectedMember.nicknames.map((nickname, idx) => (
                        <Chip key={idx} label={nickname} size="small" />
                      ))}
                    </Box>
                  </Box>
                )}

                {selectedMember.spouses && selectedMember.spouses.length > 0 && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      {t('member.spouses')}
                    </Typography>
                    {selectedMember.spouses.map((spouse) => (
                      <Box
                        key={spouse.spouse_id}
                        sx={{
                          padding: 1.5,
                          bgcolor: 'background.default',
                          borderRadius: 1,
                          marginBlockEnd: 1,
                        }}
                      >
                        <Typography variant="body2" fontWeight="medium">
                          {spouse.name}
                        </Typography>
                        {spouse.marriage_date && (
                          <Typography variant="caption" color="text.secondary">
                            {t('spouse.married')}: {formatDate(spouse.marriage_date)}
                          </Typography>
                        )}
                        {spouse.divorce_date && (
                          <Typography variant="caption" color="text.secondary" display="block">
                            {t('spouse.divorced')}: {formatDate(spouse.divorce_date)}
                          </Typography>
                        )}
                      </Box>
                    ))}
                  </Box>
                )}
              </Box>

              <Divider sx={{ my: 2 }} />

              {/* Actions */}
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                {(viewMode === 'tree' || viewMode === 'relation') && (
                  <Button
                    fullWidth
                    variant="outlined"
                    startIcon={<AccountTree />}
                    onClick={() => {
                      handleSetRoot(selectedMember.member_id);
                      setDrawerOpen(false);
                    }}
                  >
                    {t('tree.setAsTreeRoot')}
                  </Button>
                )}
              </Box>
            </Box>
          )}
        </Box>
      </Drawer>

      {/* Member Management Dialog */}
      {hasRole(Roles.ADMIN) && (
        <Dialog open={openDialog} onClose={handleCloseMemberDialog} maxWidth="lg" fullWidth>
          <DialogTitle>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Typography variant="h6">
                {editingMember ? t('member.editMember') : t('member.addNewMember')}
              </Typography>
              <IconButton onClick={handleCloseMemberDialog} size="small">
                <Close />
              </IconButton>
            </Box>
          </DialogTitle>
          {editingMember && isSuperAdmin && (
            <Box sx={{ borderBottom: 1, borderColor: 'divider', px: 3 }}>
              <Tabs value={dialogTab} onChange={(_, v) => setDialogTab(v)}>
                <Tab label={t('member.details')} />
                <Tab label={`${t('member.changeHistory')} (${memberHistory.length})`} />
              </Tabs>
            </Box>
          )}
          <DialogContent>
            {/* Details Tab */}
            {(!editingMember || dialogTab === 0) && (
              <Grid container spacing={2} sx={{ mt: 1 }}>
              {editingMember && (
                <Grid item xs={12}>
                  <Box
                    sx={{
                      display: 'flex',
                      flexDirection: { xs: 'column', sm: 'row' },
                      alignItems: { xs: 'center', sm: 'flex-start' },
                      justifyContent: 'space-between',
                      py: 3,
                      px: 2,
                      gap: 3,
                      bgcolor: 'action.hover',
                      borderRadius: 2,
                    }}
                  >
                    {/* Avatar */}
                    <Box sx={{ flexShrink: 0 }}>
                      <MemberPhotoUpload
                        memberId={editingMember.member_id}
                        currentPhoto={editingMember.picture}
                        memberName={editingMember.name}
                        gender={editingMember.gender}
                        version={editingMember.version}
                        onPhotoChange={handlePhotoChange}
                        size={120}
                        showName={false}
                      />
                    </Box>

                    {/* Member Info */}
                    <Box
                      sx={{
                        flex: 1,
                        display: 'flex',
                        flexDirection: 'column',
                        gap: 1.5,
                        textAlign: { xs: 'center', sm: 'start' },
                        alignItems: { xs: 'center', sm: 'flex-start' },
                      }}
                    >
                      {/* Member Name */}
                      <Typography variant="h5" fontWeight="bold">
                        {editingMember.name}
                      </Typography>

                      {/* Full Name - only if exists and different from name */}
                      {editingMember.full_name && editingMember.full_name !== editingMember.name && (
                        <Box>
                          <Typography variant="caption" color="text.secondary" display="block">
                            {t('member.fullName')}
                          </Typography>
                          <Typography variant="body1" fontWeight="medium" color="text.primary">
                            {editingMember.full_name}
                          </Typography>
                        </Box>
                      )}

                      {/* Age - only if exists */}
                      {editingMember.age !== undefined && editingMember.age !== null && (
                        <Typography variant="body1" color="text.secondary">
                          {t('member.age')}: <strong>{editingMember.age} {t('member.years')}</strong>
                        </Typography>
                      )}
                    </Box>
                  </Box>
                </Grid>
              )}
              {/* Multi-language name inputs */}
              {Array.isArray(languages) && languages.map((lang: Language) => (
                <Grid item xs={12} sm={6} key={lang.language_code}>
                  <TextField
                    fullWidth
                    label={`${t('member.name')} (${getLocalizedLanguageName(lang.language_code)})`}
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
                    helperText={`${t('common.add')} ${t('member.name')} (${getLocalizedLanguageName(lang.language_code)})`}
                  />
                </Grid>
              ))}
              <Grid item xs={12} sm={6}>
                <FormControl fullWidth required>
                  <InputLabel>{t('member.gender')}</InputLabel>
                  <Select
                    value={formData.gender}
                    label={t('member.gender')}
                    onChange={(e) =>
                      setFormData({ ...formData, gender: e.target.value as 'M' | 'F' })
                    }
                  >
                    <MenuItem value="M">{t('member.male')}</MenuItem>
                    <MenuItem value="F">{t('member.female')}</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label={t('member.dateOfBirth')}
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
                  label={t('member.dateOfDeath')}
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
                  label={t('member.profession')}
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
                      label={t('member.nicknames')}
                      placeholder={t('member.nicknames')}
                    />
                  )}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <ParentAutocomplete
                  label={t('member.father')}
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
                  label={t('member.mother')}
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
                    {t('member.parents')}
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
                      {t('member.spouses')} {editingMember.spouses && editingMember.spouses.length > 0 && `(${editingMember.spouses.length})`}
                    </Typography>
                    <DirectionalButton
                      variant="outlined"
                      size="small"
                      icon={<Add />}
                      onClick={() => setOpenAddSpouseDialog(true)}
                    >
                      {t('spouse.addSpouse')}
                    </DirectionalButton>
                  </Box>
                  <Grid container spacing={2}>
                    {editingMember.spouses && editingMember.spouses.length > 0 ? (
                      editingMember.spouses.map((spouse) => (
                        <Grid item xs={12} md={6} key={spouse.member_id}>
                          <SpouseCard
                            spouse={spouse}
                            currentMemberId={editingMember.member_id}
                            onUpdate={async () => {
                              if (viewMode === 'list') {
                                loadListView();
                              } else if (viewMode === 'tree') {
                                loadTree();
                              }
                              // Refetch member data after spouse update/delete
                              try {
                                const updated = await membersApi.getMember(editingMember.member_id);
                                setEditingMember(updated);
                              } catch (error) {
                                console.error('Failed to refresh member:', error);
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
                            {t('member.noSpousesAdded')}
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
                    {t('member.children')} ({editingMember.children.length})
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
                    {t('member.siblings')} ({editingMember.siblings.length})
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
            {editingMember && isSuperAdmin && dialogTab === 1 && (
              <Box sx={{ mt: 2 }}>
                {memberHistory.length === 0 ? (
                  <Box sx={{ textAlign: 'center', py: 4 }}>
                    <Typography variant="body2" color="text.secondary">
                      {t('userProfile.noRecentChanges')}
                    </Typography>
                  </Box>
                ) : (
                  <>
                    <TableContainer>
                      <Table>
                        <TableHead>
                          <TableRow>
                            <TableCell>{t('userProfile.changeType')}</TableCell>
                            <TableCell>{t('leaderboard.user')}</TableCell>
                            <TableCell>{t('userProfile.date')}</TableCell>
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
                                    {formatRelativeTime(change.changed_at, t)}
                                  </Typography>
                                </Box>
                              </TableCell>
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
                          {t('userProfile.loadMore')} ({memberHistory.length - displayedHistoryCount} {t('userProfile.remaining')})
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
                  <Tooltip title={t('member.deleteMemberTooltip')}>
                    <IconButton
                      onClick={() => handleDeleteClick(editingMember.member_id)}
                      color="error"
                      size="small"
                    >
                      <Delete />
                    </IconButton>
                  </Tooltip>
                )}
              </Box>
              <Box sx={{ display: 'flex', gap: 1 }}>
                <Button onClick={handleCloseMemberDialog}>{t('member.cancel')}</Button>
                {(!editingMember || dialogTab === 0) && (
                  <Button onClick={handleSubmitMember} variant="contained">
                    {editingMember ? t('member.update') : t('member.create')}
                  </Button>
                )}
              </Box>
            </Box>
          </DialogActions>
        </Dialog>
      )}

      {/* Add Spouse Dialog */}
      {editingMember && (
        <AddSpouseDialog
          open={openAddSpouseDialog}
          onClose={() => setOpenAddSpouseDialog(false)}
          memberId={editingMember.member_id}
          memberName={editingMember.name}
          memberGender={editingMember.gender}
          onSuccess={() => {
            if (viewMode === 'list') {
              loadListView();
            } else if (viewMode === 'tree') {
              loadTree();
            }
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

      {/* History Diff Dialog */}
      <HistoryDiffDialog
        open={diffDialogOpen}
        onClose={handleCloseDiff}
        history={selectedHistory}
      />

      {/* Delete Confirmation Dialog */}
      <Dialog
        open={openDeleteDialog}
        onClose={handleDeleteCancel}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle sx={{ color: 'error.main' }}>
          {t('member.deleteWarningTitle')}
        </DialogTitle>
        <DialogContent>
          <Typography variant="body1" gutterBottom>
            {t('member.deleteWarningMessage')}
          </Typography>

          <Typography variant="subtitle2" sx={{ mt: 2, mb: 1, fontWeight: 'bold' }}>
            {t('member.deleteWillRemove')}
          </Typography>

          <Box component="ul" sx={{ pl: 2, my: 1 }}>
            <Typography component="li" variant="body2" sx={{ mb: 0.5 }}>
              {t('member.deleteItemNames')}
            </Typography>
            <Typography component="li" variant="body2" sx={{ mb: 0.5 }}>
              {t('member.deleteItemSpouses')}
            </Typography>
            <Typography component="li" variant="body2" sx={{ mb: 0.5 }}>
              {t('member.deleteItemPicture')}
            </Typography>
            <Typography component="li" variant="body2" sx={{ mb: 0.5 }}>
              {t('member.deleteItemSoftDelete')}
            </Typography>
          </Box>

          <Typography variant="body2" color="text.secondary" sx={{ mt: 2, mb: 2 }}>
            {t('member.deleteHistoryNote')}
          </Typography>

          <Typography variant="body1" sx={{ mt: 2, fontWeight: 'bold', color: 'error.main' }}>
            {t('member.deleteConfirmation')}
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button
            onClick={handleDeleteCancel}
            disabled={deleting}
          >
            {t('common.cancel')}
          </Button>
          <Button
            onClick={handleDeleteConfirm}
            variant="contained"
            color="error"
            disabled={deleting}
          >
            {deleting ? t('member.deleting') : t('member.deleteMember')}
          </Button>
        </DialogActions>
      </Dialog>
    </Layout>
  );
};

export default TreePage;
