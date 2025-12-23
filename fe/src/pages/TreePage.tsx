import React, { useEffect, useState } from 'react';
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
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material';
import {
  AccountTree,
  TableChart,
  Close,
  ExpandMore,
  Refresh,
} from '@mui/icons-material';
import { treeApi, membersApi } from '../api';
import { TreeNode, Member } from '../types';
import { getGenderColor, formatDate, formatDateOfBirth, getMemberPictureUrl } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import TreeVisualization from '../components/TreeVisualization';
import MemberTableView from '../components/MemberTableView';
import RelationFinder from '../components/RelationFinder';
import { useAuth } from '../contexts/AuthContext';
import { Roles } from '../types';

type ViewMode = 'tree' | 'list' | 'relation';

const TreePage: React.FC = () => {
  const { hasRole } = useAuth();
  const isSuperAdmin = hasRole(Roles.SUPER_ADMIN);

  // View state
  const [viewMode, setViewMode] = useState<ViewMode>('tree');
  const [rootId, setRootId] = useState<number | undefined>(undefined);

  // Data state
  const [treeData, setTreeData] = useState<TreeNode | null>(null);
  const [listMembers, setListMembers] = useState<Member[]>([]);
  const [relationTree, setRelationTree] = useState<TreeNode | null>(null);
  const [allMembers, setAllMembers] = useState<Member[]>([]);

  // UI state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedMember, setSelectedMember] = useState<Member | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [relationLoading, setRelationLoading] = useState(false);

  // Load all members for search/autocomplete
  useEffect(() => {
    loadAllMembers();
  }, []);

  // Load data based on view mode
  useEffect(() => {
    if (viewMode === 'tree') {
      loadTree();
    } else if (viewMode === 'list') {
      loadListView();
    }
  }, [rootId, viewMode]);

  const loadAllMembers = async () => {
    try {
      const data = await treeApi.getListView();
      setAllMembers(data);
    } catch (error) {
      console.error('Failed to load members:', error);
    }
  };

  const loadTree = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await treeApi.getTree({ root: rootId, style: 'tree' });
      setTreeData(data);
    } catch (error) {
      console.error('Failed to load tree:', error);
      setError('Failed to load family tree. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const loadListView = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await treeApi.getListView();
      setListMembers(data);
    } catch (error) {
      console.error('Failed to load list view:', error);
      setError('Failed to load member list. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleFindRelation = async (member1Id: number, member2Id: number) => {
    setRelationLoading(true);
    setError(null);
    try {
      const data = await treeApi.getRelation({ member1: member1Id, member2: member2Id });
      setRelationTree(data);
      setViewMode('relation');
    } catch (error) {
      console.error('Failed to find relation:', error);
      setError('No relation found between the selected members.');
    } finally {
      setRelationLoading(false);
    }
  };

  const handleMemberClick = async (member: Member) => {
    try {
      const fullMember = await membersApi.getMember(member.member_id);
      setSelectedMember(fullMember);
      setDrawerOpen(true);
    } catch (error) {
      console.error('Failed to load member details:', error);
    }
  };

  const handleSetRoot = (memberId: number) => {
    setRootId(memberId);
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
      <Box sx={{ mb: 3 }}>
        {/* Page Header */}
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Box>
            <Typography variant="h4" gutterBottom>
              Family Tree
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Explore your family connections through interactive visualizations
            </Typography>
          </Box>
          {rootId && (
            <Button startIcon={<Refresh />} onClick={handleResetRoot} variant="outlined">
              Reset to Default Root
            </Button>
          )}
        </Box>

        <Divider sx={{ mb: 3 }} />

        {/* Section 1: View Mode Selection */}
        <Paper sx={{ p: 2, mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            View Mode
          </Typography>
          <ToggleButtonGroup
            value={viewMode}
            exclusive
            onChange={handleViewModeChange}
            aria-label="view mode"
          >
            <ToggleButton value="tree" aria-label="tree view">
              <AccountTree sx={{ mr: 1 }} />
              Tree Diagram
            </ToggleButton>
            <ToggleButton value="list" aria-label="list view">
              <TableChart sx={{ mr: 1 }} />
              Table View
            </ToggleButton>
            <ToggleButton value="relation" aria-label="relation view">
              <AccountTree sx={{ mr: 1 }} />
              Find Relation
            </ToggleButton>
          </ToggleButtonGroup>
        </Paper>

        {/* Section 2: Relation Finder (visible when in relation mode) */}
        {viewMode === 'relation' && (
          <Box sx={{ mb: 3 }}>
            <RelationFinder
              members={allMembers}
              onFindRelation={handleFindRelation}
              loading={relationLoading}
            />
          </Box>
        )}

        {/* Section 3: Tips & Instructions */}
        <Accordion sx={{ mb: 3 }}>
          <AccordionSummary expandIcon={<ExpandMore />}>
            <Typography variant="h6">Tips & Instructions</Typography>
          </AccordionSummary>
          <AccordionDetails>
            <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '1fr 1fr' }, gap: 2 }}>
              <Box>
                <Typography variant="subtitle2" gutterBottom>
                  Tree Diagram View:
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  • Click on any node to view detailed member information
                  <br />
                  • Right-click a node to set it as the tree root
                  <br />
                  • Drag to pan, scroll to zoom
                  <br />
                  • Pink lines connect spouses
                  <br />
                  • Black lines connect parents to children
                  <br />• Orange highlights show the relation path
                </Typography>
              </Box>
              <Box>
                <Typography variant="subtitle2" gutterBottom>
                  Table View:
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  • Search members by name, nickname, or profession
                  <br />
                  • Click column headers to sort
                  <br />
                  • Click on any row to view details
                  <br />• Use the tree icon to set member as root
                </Typography>
              </Box>
            </Box>
          </AccordionDetails>
        </Accordion>

        {/* Error Display */}
        {error && (
          <Alert severity="error" sx={{ mb: 3 }} onClose={() => setError(null)}>
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
                  Tree Visualization
                </Typography>
                <TreeVisualization
                  data={treeData}
                  onNodeClick={handleMemberClick}
                  onSetRoot={handleSetRoot}
                />
              </Box>
            )}

            {/* List View */}
            {viewMode === 'list' && listMembers.length > 0 && (
              <Box>
                <Typography variant="h6" gutterBottom>
                  All Family Members ({listMembers.length})
                </Typography>
                <MemberTableView
                  members={listMembers}
                  onViewMember={handleMemberClick}
                  onSetRoot={handleSetRoot}
                />
              </Box>
            )}

            {/* Relation View */}
            {viewMode === 'relation' && relationTree && (
              <Box>
                <Typography variant="h6" gutterBottom>
                  Relation Path (Orange highlights show the connection)
                </Typography>
                <TreeVisualization
                  data={relationTree}
                  onNodeClick={handleMemberClick}
                  onSetRoot={handleSetRoot}
                />
              </Box>
            )}

            {/* Empty State */}
            {viewMode === 'relation' && !relationTree && !relationLoading && (
              <Paper sx={{ p: 5, textAlign: 'center' }}>
                <Typography variant="h6" color="text.secondary" gutterBottom>
                  Select Two Members to Find Their Relation
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Use the relation finder above to explore family connections
                </Typography>
              </Paper>
            )}
          </>
        )}
      </Box>

      {/* Member Details Drawer */}
      <Drawer anchor="right" open={drawerOpen} onClose={() => setDrawerOpen(false)}>
        <Box sx={{ width: 450, p: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
            <Typography variant="h6">Member Details</Typography>
            <Button onClick={() => setDrawerOpen(false)} startIcon={<Close />}>
              Close
            </Button>
          </Box>

          {selectedMember && (
            <Box>
              <Avatar
                src={
                  getMemberPictureUrl(selectedMember.member_id, selectedMember.picture) || undefined
                }
                sx={{
                  width: 120,
                  height: 120,
                  mx: 'auto',
                  mb: 2,
                  bgcolor: getGenderColor(selectedMember.gender),
                }}
              >
                {selectedMember.english_name[0]}
              </Avatar>

              <Typography variant="h6" align="center" gutterBottom>
                {selectedMember.arabic_name}
              </Typography>
              <Typography variant="body1" align="center" color="text.secondary" gutterBottom>
                {selectedMember.english_name}
              </Typography>

              {/* Full Name */}
              {(selectedMember.english_full_name || selectedMember.arabic_full_name) && (
                <Box sx={{ mt: 2, p: 2, bgcolor: 'background.default', borderRadius: 1 }}>
                  <Typography variant="caption" color="text.secondary" gutterBottom display="block">
                    Full Name
                  </Typography>
                  {selectedMember.english_full_name && (
                    <Typography variant="body2" fontWeight="medium" gutterBottom>
                      {selectedMember.english_full_name}
                    </Typography>
                  )}
                  {selectedMember.arabic_full_name && (
                    <Typography variant="body2" fontWeight="medium" dir="rtl">
                      {selectedMember.arabic_full_name}
                    </Typography>
                  )}
                </Box>
              )}

              <Divider sx={{ my: 2 }} />

              {/* Details Grid */}
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                <Box>
                  <Typography variant="subtitle2" color="text.secondary">
                    Gender
                  </Typography>
                  <Chip
                    label={
                      selectedMember.gender === 'M'
                        ? 'Male'
                        : selectedMember.gender === 'F'
                        ? 'Female'
                        : 'Other'
                    }
                    size="small"
                    sx={{ bgcolor: getGenderColor(selectedMember.gender), color: 'white' }}
                  />
                </Box>

                {selectedMember.date_of_birth && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      Date of Birth
                    </Typography>
                    <Typography variant="body1">
                      {formatDateOfBirth(selectedMember.date_of_birth, isSuperAdmin)}
                    </Typography>
                  </Box>
                )}

                {selectedMember.date_of_death && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      Date of Death
                    </Typography>
                    <Typography variant="body1">{formatDate(selectedMember.date_of_death)}</Typography>
                  </Box>
                )}

                {selectedMember.age && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      Age
                    </Typography>
                    <Typography variant="body1">{selectedMember.age} years</Typography>
                  </Box>
                )}

                {selectedMember.generation_level !== undefined && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      Generation Level
                    </Typography>
                    <Typography variant="body1">{selectedMember.generation_level}</Typography>
                  </Box>
                )}

                {selectedMember.profession && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary">
                      Profession
                    </Typography>
                    <Typography variant="body1">{selectedMember.profession}</Typography>
                  </Box>
                )}

                {selectedMember.nicknames && selectedMember.nicknames.length > 0 && (
                  <Box>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      Nicknames
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
                      Spouses
                    </Typography>
                    {selectedMember.spouses.map((spouse) => (
                      <Box
                        key={spouse.spouse_id}
                        sx={{
                          p: 1.5,
                          bgcolor: 'background.default',
                          borderRadius: 1,
                          mb: 1,
                        }}
                      >
                        <Typography variant="body2" fontWeight="medium">
                          {spouse.arabic_name} ({spouse.english_name})
                        </Typography>
                        {spouse.marriage_date && (
                          <Typography variant="caption" color="text.secondary">
                            Married: {formatDate(spouse.marriage_date)}
                          </Typography>
                        )}
                        {spouse.divorce_date && (
                          <Typography variant="caption" color="text.secondary" display="block">
                            Divorced: {formatDate(spouse.divorce_date)}
                          </Typography>
                        )}
                      </Box>
                    ))}
                  </Box>
                )}
              </Box>

              <Divider sx={{ my: 2 }} />

              {/* Actions */}
              <Box sx={{ display: 'flex', gap: 1 }}>
                <Button
                  fullWidth
                  variant="outlined"
                  startIcon={<AccountTree />}
                  onClick={() => {
                    handleSetRoot(selectedMember.member_id);
                    setDrawerOpen(false);
                  }}
                >
                  Set as Root
                </Button>
              </Box>
            </Box>
          )}
        </Box>
      </Drawer>
    </Layout>
  );
};

export default TreePage;
