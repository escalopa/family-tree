import React, { useEffect, useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  ToggleButtonGroup,
  ToggleButton,
  TextField,
  Grid,
  Card,
  CardContent,
  Avatar,
  Chip,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Drawer,
} from '@mui/material';
import { AccountTree, List as ListIcon, Search, Close } from '@mui/icons-material';
import { treeApi, membersApi } from '../api';
import { TreeNode, Member, MemberSearchQuery } from '../types';
import { getGenderColor, formatDate, calculateAge } from '../utils/helpers';
import Layout from '../components/Layout/Layout';

const TreePage: React.FC = () => {
  const [viewStyle, setViewStyle] = useState<'tree' | 'list'>('tree');
  const [rootId, setRootId] = useState<number | undefined>(undefined);
  const [treeData, setTreeData] = useState<TreeNode | null>(null);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState<MemberSearchQuery>({});
  const [searchResults, setSearchResults] = useState<Member[]>([]);
  const [selectedMember, setSelectedMember] = useState<Member | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);

  useEffect(() => {
    loadTree();
  }, [rootId, viewStyle]);

  const loadTree = async () => {
    setLoading(true);
    try {
      const data = await treeApi.getTree({ root: rootId, style: viewStyle });
      setTreeData(data);
    } catch (error) {
      console.error('Failed to load tree:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async () => {
    try {
      const response = await membersApi.searchMembers(searchQuery);
      setSearchResults(response.members);
    } catch (error) {
      console.error('Search failed:', error);
    }
  };

  const handleMemberClick = async (memberId: number) => {
    try {
      const member = await membersApi.getMember(memberId);
      setSelectedMember(member);
      setDrawerOpen(true);
    } catch (error) {
      console.error('Failed to load member:', error);
    }
  };

  const renderTreeNode = (node: TreeNode, level: number = 0) => {
    const { member } = node;
    return (
      <Box key={member.member_id} sx={{ ml: level * 4, my: 1 }}>
        <Card
          sx={{
            cursor: 'pointer',
            borderLeft: `4px solid ${getGenderColor(member.gender)}`,
            '&:hover': { boxShadow: 4 },
          }}
          onClick={() => handleMemberClick(member.member_id)}
        >
          <CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2, py: 2 }}>
            <Avatar
              src={member.picture || undefined}
              sx={{ bgcolor: getGenderColor(member.gender) }}
            >
              {member.english_name[0]}
            </Avatar>
            <Box sx={{ flexGrow: 1 }}>
              <Typography variant="subtitle1">{member.arabic_name}</Typography>
              <Typography variant="body2" color="text.secondary">
                {member.english_name}
              </Typography>
            </Box>
            {member.is_married && <Chip label="Married" size="small" />}
            {member.age && (
              <Typography variant="body2" color="text.secondary">
                Age: {member.age}
              </Typography>
            )}
          </CardContent>
        </Card>
        {node.children && node.children.map((child) => renderTreeNode(child, level + 1))}
      </Box>
    );
  };

  const renderListView = (node: TreeNode) => {
    const allMembers: Member[] = [];
    const collectMembers = (n: TreeNode) => {
      allMembers.push(n.member);
      n.children?.forEach(collectMembers);
    };
    collectMembers(node);

    return (
      <Grid container spacing={2}>
        {allMembers.map((member) => (
          <Grid item xs={12} sm={6} md={4} key={member.member_id}>
            <Card
              sx={{
                cursor: 'pointer',
                borderLeft: `4px solid ${getGenderColor(member.gender)}`,
                '&:hover': { boxShadow: 4 },
              }}
              onClick={() => handleMemberClick(member.member_id)}
            >
              <CardContent sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                <Avatar
                  src={member.picture || undefined}
                  sx={{ bgcolor: getGenderColor(member.gender) }}
                >
                  {member.english_name[0]}
                </Avatar>
                <Box sx={{ flexGrow: 1 }}>
                  <Typography variant="subtitle1">{member.arabic_name}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    {member.english_name}
                  </Typography>
                  {member.age && (
                    <Typography variant="caption" color="text.secondary">
                      Age: {member.age}
                    </Typography>
                  )}
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    );
  };

  return (
    <Layout>
      <Box sx={{ mb: 3 }}>
        <Typography variant="h4" gutterBottom>
          Family Tree
        </Typography>

        {/* View Toggle */}
        <Box sx={{ display: 'flex', gap: 2, mb: 3, flexWrap: 'wrap' }}>
          <ToggleButtonGroup
            value={viewStyle}
            exclusive
            onChange={(_, value) => value && setViewStyle(value)}
          >
            <ToggleButton value="tree">
              <AccountTree sx={{ mr: 1 }} />
              Tree View
            </ToggleButton>
            <ToggleButton value="list">
              <ListIcon sx={{ mr: 1 }} />
              List View
            </ToggleButton>
          </ToggleButtonGroup>
        </Box>

        {/* Search Filters */}
        <Paper sx={{ p: 2, mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            Search Members
          </Typography>
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} md={3}>
              <TextField
                fullWidth
                label="Arabic Name"
                value={searchQuery.arabic_name || ''}
                onChange={(e) =>
                  setSearchQuery({ ...searchQuery, arabic_name: e.target.value })
                }
              />
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <TextField
                fullWidth
                label="English Name"
                value={searchQuery.english_name || ''}
                onChange={(e) =>
                  setSearchQuery({ ...searchQuery, english_name: e.target.value })
                }
              />
            </Grid>
            <Grid item xs={12} sm={6} md={2}>
              <FormControl fullWidth>
                <InputLabel>Gender</InputLabel>
                <Select
                  value={searchQuery.gender || ''}
                  label="Gender"
                  onChange={(e) =>
                    setSearchQuery({ ...searchQuery, gender: e.target.value })
                  }
                >
                  <MenuItem value="">All</MenuItem>
                  <MenuItem value="M">Male</MenuItem>
                  <MenuItem value="F">Female</MenuItem>
                  <MenuItem value="N">Other</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6} md={2}>
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
                >
                  <MenuItem value="">All</MenuItem>
                  <MenuItem value={1}>Yes</MenuItem>
                  <MenuItem value={0}>No</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6} md={2}>
              <Button
                fullWidth
                variant="contained"
                startIcon={<Search />}
                onClick={handleSearch}
                sx={{ height: '56px' }}
              >
                Search
              </Button>
            </Grid>
          </Grid>

          {/* Search Results */}
          {searchResults.length > 0 && (
            <Box sx={{ mt: 2 }}>
              <Typography variant="subtitle2" gutterBottom>
                Search Results ({searchResults.length})
              </Typography>
              <Grid container spacing={1}>
                {searchResults.map((member) => (
                  <Grid item key={member.member_id}>
                    <Chip
                      label={`${member.arabic_name} (${member.english_name})`}
                      onClick={() => handleMemberClick(member.member_id)}
                      onDelete={() => setRootId(member.member_id)}
                      deleteIcon={<AccountTree />}
                    />
                  </Grid>
                ))}
              </Grid>
            </Box>
          )}
        </Paper>
      </Box>

      {/* Tree/List Display */}
      {loading ? (
        <Typography>Loading...</Typography>
      ) : treeData ? (
        viewStyle === 'tree' ? (
          renderTreeNode(treeData)
        ) : (
          renderListView(treeData)
        )
      ) : (
        <Typography>No data available</Typography>
      )}

      {/* Member Details Drawer */}
      <Drawer anchor="right" open={drawerOpen} onClose={() => setDrawerOpen(false)}>
        <Box sx={{ width: 400, p: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
            <Typography variant="h6">Member Details</Typography>
            <Button onClick={() => setDrawerOpen(false)}>
              <Close />
            </Button>
          </Box>
          {selectedMember && (
            <Box>
              <Avatar
                src={selectedMember.picture || undefined}
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

              <Box sx={{ mt: 3 }}>
                <Typography variant="subtitle2" color="text.secondary">
                  Gender
                </Typography>
                <Typography variant="body1" gutterBottom>
                  {selectedMember.gender === 'M' ? 'Male' : selectedMember.gender === 'F' ? 'Female' : 'Other'}
                </Typography>

                {selectedMember.date_of_birth && (
                  <>
                    <Typography variant="subtitle2" color="text.secondary">
                      Date of Birth
                    </Typography>
                    <Typography variant="body1" gutterBottom>
                      {formatDate(selectedMember.date_of_birth)}
                    </Typography>
                  </>
                )}

                {selectedMember.date_of_death && (
                  <>
                    <Typography variant="subtitle2" color="text.secondary">
                      Date of Death
                    </Typography>
                    <Typography variant="body1" gutterBottom>
                      {formatDate(selectedMember.date_of_death)}
                    </Typography>
                  </>
                )}

                {selectedMember.age && (
                  <>
                    <Typography variant="subtitle2" color="text.secondary">
                      Age
                    </Typography>
                    <Typography variant="body1" gutterBottom>
                      {selectedMember.age} years
                    </Typography>
                  </>
                )}

                {selectedMember.profession && (
                  <>
                    <Typography variant="subtitle2" color="text.secondary">
                      Profession
                    </Typography>
                    <Typography variant="body1" gutterBottom>
                      {selectedMember.profession}
                    </Typography>
                  </>
                )}

                {selectedMember.nicknames && selectedMember.nicknames.length > 0 && (
                  <>
                    <Typography variant="subtitle2" color="text.secondary">
                      Nicknames
                    </Typography>
                    <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap', mb: 2 }}>
                      {selectedMember.nicknames.map((nickname, idx) => (
                        <Chip key={idx} label={nickname} size="small" />
                      ))}
                    </Box>
                  </>
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
