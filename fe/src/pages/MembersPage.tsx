import React, { useState } from 'react';
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
} from '@mui/material';
import { Add, Edit, Delete, Search } from '@mui/icons-material';
import { membersApi } from '../api';
import { Member, MemberSearchQuery, CreateMemberRequest, UpdateMemberRequest } from '../types';
import { formatDate } from '../utils/helpers';
import Layout from '../components/Layout/Layout';
import MemberPhotoUpload from '../components/MemberPhotoUpload';

const MembersPage: React.FC = () => {
  const [members, setMembers] = useState<Member[]>([]);
  const [searchQuery, setSearchQuery] = useState<MemberSearchQuery>({});
  const [openDialog, setOpenDialog] = useState(false);
  const [editingMember, setEditingMember] = useState<Member | null>(null);
  const [formData, setFormData] = useState<CreateMemberRequest>({
    arabic_name: '',
    english_name: '',
    gender: 'M',
  });

  const handleSearch = async () => {
    try {
      const response = await membersApi.searchMembers(searchQuery);
      setMembers(response.members);
    } catch (error) {
      console.error('Search failed:', error);
    }
  };

  const handleOpenDialog = (member?: Member) => {
    if (member) {
      setEditingMember(member);
      setFormData({
        arabic_name: member.arabic_name,
        english_name: member.english_name,
        gender: member.gender,
        date_of_birth: member.date_of_birth || undefined,
        date_of_death: member.date_of_death || undefined,
        father_id: member.father_id || undefined,
        mother_id: member.mother_id || undefined,
        nicknames: member.nicknames || [],
        profession: member.profession || undefined,
      });
    } else {
      setEditingMember(null);
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
  };

  const handleSubmit = async () => {
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
      handleSearch(); // Refresh list
    } catch (error) {
      console.error('Failed to save member:', error);
    }
  };

  const handleDelete = async (memberId: number) => {
    if (confirm('Are you sure you want to delete this member?')) {
      try {
        await membersApi.deleteMember(memberId);
        handleSearch(); // Refresh list
      } catch (error) {
        console.error('Failed to delete member:', error);
      }
    }
  };

  const handlePhotoChange = () => {
    handleSearch(); // Refresh list after photo change
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
        </Paper>

        {/* Members Table */}
        <TableContainer component={Paper}>
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
              {members.map((member) => (
                <TableRow key={member.member_id}>
                  <TableCell>
                    <MemberPhotoUpload
                      memberId={member.member_id}
                      currentPhoto={member.picture}
                      memberName={member.english_name}
                      gender={member.gender}
                      onPhotoChange={handlePhotoChange}
                      size={50}
                      compact
                    />
                  </TableCell>
                  <TableCell>{member.arabic_name}</TableCell>
                  <TableCell>{member.english_name}</TableCell>
                  <TableCell>
                    {member.gender === 'M' ? 'Male' : member.gender === 'F' ? 'Female' : 'Other'}
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

        {/* Create/Edit Dialog */}
        <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="md" fullWidth>
          <DialogTitle>
            {editingMember ? 'Edit Member' : 'Add New Member'}
          </DialogTitle>
          <DialogContent>
            <Grid container spacing={2} sx={{ mt: 1 }}>
              {editingMember && (
                <Grid item xs={12}>
                  <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}>
                    <MemberPhotoUpload
                      memberId={editingMember.member_id}
                      currentPhoto={editingMember.picture}
                      memberName={editingMember.english_name}
                      gender={editingMember.gender}
                      onPhotoChange={handlePhotoChange}
                      size={120}
                      showName
                    />
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
                      setFormData({ ...formData, gender: e.target.value as 'M' | 'F' | 'N' })
                    }
                  >
                    <MenuItem value="M">Male</MenuItem>
                    <MenuItem value="F">Female</MenuItem>
                    <MenuItem value="N">Other</MenuItem>
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
                <TextField
                  fullWidth
                  label="Father ID"
                  type="number"
                  value={formData.father_id || ''}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      father_id: e.target.value ? Number(e.target.value) : undefined,
                    })
                  }
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label="Mother ID"
                  type="number"
                  value={formData.mother_id || ''}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      mother_id: e.target.value ? Number(e.target.value) : undefined,
                    })
                  }
                />
              </Grid>
            </Grid>
          </DialogContent>
          <DialogActions>
            <Button onClick={handleCloseDialog}>Cancel</Button>
            <Button onClick={handleSubmit} variant="contained">
              {editingMember ? 'Update' : 'Create'}
            </Button>
          </DialogActions>
        </Dialog>
      </Box>
    </Layout>
  );
};

export default MembersPage;
