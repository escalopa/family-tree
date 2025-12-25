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
} from '@mui/material';
import {
  AccountTree,
  TableChart,
  Close,
  Refresh,
  AccountTreeOutlined,
  BubbleChart,
  FilterAlt,
  Clear,
} from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useTranslation } from 'react-i18next';
import { treeApi, membersApi } from '../api';
import { MemberListItem, MemberSearchQuery } from '../types';
import { TreeNode, Member } from '../types';
import { getGenderColor, formatDate, formatDateOfBirth, getMemberPictureUrl } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import ForceDirectedTree from '../components/ForceDirectedTree';
import TreeVisualization from '../components/TreeVisualization';
import RelationFinder from '../components/RelationFinder';
import { useAuth } from '../contexts/AuthContext';
import { useLanguage } from '../contexts/LanguageContext';
import { Roles } from '../types';

type ViewMode = 'tree' | 'list' | 'relation';
type TreeLayout = 'hierarchical' | 'force';

const TreePage: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { hasRole } = useAuth();
  const { getPreferredName, getAllNamesFormatted } = useLanguage();
  const isSuperAdmin = hasRole(Roles.SUPER_ADMIN);
  const isRTL = i18n.dir() === 'rtl';
  const [searchParams, setSearchParams] = useSearchParams();

  // Initialize from URL params
  const initialViewMode = (searchParams.get('view') as ViewMode) || 'tree';
  const initialLayout = (searchParams.get('layout') as TreeLayout) || 'force';
  const initialRootId = searchParams.get('root') ? parseInt(searchParams.get('root')!) : undefined;

  // View state
  const [viewMode, setViewMode] = useState<ViewMode>(initialViewMode);
  const [treeLayout, setTreeLayout] = useState<TreeLayout>(initialLayout);
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

  // Ref for intersection observer
  const loadMoreRef = useRef<HTMLDivElement>(null);

  // Update URL params when state changes
  useEffect(() => {
    const params = new URLSearchParams();
    params.set('view', viewMode);
    if (viewMode === 'tree' || viewMode === 'relation') {
      params.set('layout', treeLayout);
    }
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
  }, [viewMode, treeLayout, rootId, searchQuery, setSearchParams]);

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
    try {
      const fullMember = await membersApi.getMember(member.member_id);
      setSelectedMember(fullMember);
      setDrawerOpen(true);
    } catch (error) {

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

  return (
    <Layout>
      <Box sx={{ width: '100%' }}>
        {/* Page Header */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
        >
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
        </motion.div>

        {/* Section 1: View Mode Selection */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, delay: 0.1 }}
        >
        <Paper sx={{ p: 2, mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            {t('tree.viewMode')}
          </Typography>
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

            {/* Layout Toggle (only for tree/relation views) */}
            {(viewMode === 'tree' || viewMode === 'relation') && (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <Divider orientation="vertical" flexItem />
                <Typography variant="body2" color="text.secondary">
                  {t('tree.layout')}:
                </Typography>
                <ToggleButtonGroup
                  value={treeLayout}
                  exclusive
                  onChange={(_, value) => value && setTreeLayout(value)}
                  size="small"
                  aria-label="tree layout"
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
                      '&:not(:first-of-type)': {
                        marginInlineStart: '-1px',
                      }
                    }
                  }}
                >
                  <ToggleButton value="hierarchical" aria-label="hierarchical layout">
                    <AccountTreeOutlined sx={{ marginInlineEnd: 0.5 }} fontSize="small" />
                    {t('tree.hierarchical')}
                  </ToggleButton>
                  <ToggleButton value="force" aria-label="force directed layout">
                    <BubbleChart sx={{ marginInlineEnd: 0.5 }} fontSize="small" />
                    {t('tree.forceDirected')}
                  </ToggleButton>
                </ToggleButtonGroup>
            </Box>
          )}
          </Box>
        </Paper>
        </motion.div>

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
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.4, delay: 0.2 }}
              >
                <Box>
                  <Typography variant="h6" gutterBottom>
                    {treeLayout === 'force' ? t('tree.interactiveFamilyGraph') : t('tree.hierarchicalTreeView')}
                  </Typography>
                  {treeLayout === 'force' ? (
                    <ForceDirectedTree
                      data={treeData}
                      onNodeClick={handleMemberClick}
                      onSetRoot={handleSetRoot}
                      currentRootId={rootId}
                    />
                  ) : (
                    <TreeVisualization
                      data={treeData}
                      onNodeClick={handleMemberClick}
                      onSetRoot={handleSetRoot}
                      currentRootId={rootId}
                    />
                  )}
                </Box>
              </motion.div>
            )}

            {/* List View */}
            {viewMode === 'list' && (
              <Box>
                {/* Search Filters */}
                <Paper sx={{ p: 2, mb: 3 }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                    <FilterAlt sx={{ marginInlineEnd: 1, color: 'text.secondary' }} />
                    <Typography variant="h6" sx={{ flexGrow: 1 }}>
                      {t('tree.searchFilters')} {!searchQuery.name && !searchQuery.gender && searchQuery.married === undefined && t('tree.showingAllMembers')}
                    </Typography>
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
                        <TableCell>{t('general.id')}</TableCell>
                        <TableCell>{t('member.avatar')}</TableCell>
                        <TableCell>{t('member.name')}</TableCell>
                        <TableCell>{t('member.gender')}</TableCell>
                        <TableCell>{t('member.dateOfBirth')}</TableCell>
                        <TableCell>{t('member.married')}</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {(!listMembers || listMembers.length === 0) && !loading && (
                        <TableRow>
                          <TableCell colSpan={6} align="center" sx={{ py: 4, color: 'text.secondary' }}>
                            {searchQuery.name || searchQuery.gender || searchQuery.married !== undefined
                              ? t('tree.noMembersMatchingFilters')
                              : t('member.noMembers')}
                          </TableCell>
                        </TableRow>
                      )}
                      {loading && listMembers.length === 0 && (
                        <TableRow>
                          <TableCell colSpan={6} align="center" sx={{ py: 8 }}>
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
                          sx={{ cursor: 'pointer' }}
                          onClick={() => handleMemberClick(member)}
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
                            {member.gender === 'M' ? t('member.male') : t('member.female')}
                          </TableCell>
                          <TableCell>{formatDateOfBirth(member.date_of_birth, isSuperAdmin)}</TableCell>
                          <TableCell>
                            {member.is_married ? (
                              <Chip label={t('common.yes')} color="primary" size="small" />
                            ) : (
                              <Chip label={t('common.no')} size="small" />
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
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.4, delay: 0.2 }}
              >
                <Box>
                  <Typography variant="h6" gutterBottom>
                    {t('tree.relationPath')} {t('tree.relationPathDescription')}
                  </Typography>
                  {treeLayout === 'force' ? (
                    <ForceDirectedTree
                      data={relationTree}
                      onNodeClick={handleMemberClick}
                      onSetRoot={handleSetRoot}
                      currentRootId={rootId}
                    />
                  ) : (
                    <TreeVisualization
                      data={relationTree}
                      onNodeClick={handleMemberClick}
                      onSetRoot={handleSetRoot}
                      currentRootId={rootId}
                    />
                  )}
                </Box>
              </motion.div>
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
    </Layout>
  );
};

export default TreePage;
