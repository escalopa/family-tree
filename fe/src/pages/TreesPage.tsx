import React, { useEffect, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  Grid,
  IconButton,
  Paper,
  Stack,
  TextField,
  Typography,
} from '@mui/material';
import { AccountTree, Add, Link, Login, PersonAdd, Refresh } from '@mui/icons-material';
import { enqueueSnackbar } from 'notistack';
import { useNavigate } from 'react-router-dom';
import { familyTreesApi, setActiveTreeId } from '../api';
import { FamilyTree, FamilyTreeInvitation, FamilyTreeShareLink } from '../types';
import Layout from '../components/Layout/Layout';

const TreesPage: React.FC = () => {
  const navigate = useNavigate();
  const [trees, setTrees] = useState<FamilyTree[]>([]);
  const [invitations, setInvitations] = useState<FamilyTreeInvitation[]>([]);
  const [shareLinks, setShareLinks] = useState<Record<number, FamilyTreeShareLink[]>>({});
  const [inviteEmails, setInviteEmails] = useState<Record<number, string>>({});
  const [shareMaxVisits, setShareMaxVisits] = useState<Record<number, string>>({});
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [newTreeName, setNewTreeName] = useState('');
  const [newTreeDescription, setNewTreeDescription] = useState('');

  const load = async () => {
    setLoading(true);
    try {
      const [treeList, invitationList] = await Promise.all([
        familyTreesApi.list(),
        familyTreesApi.listMyInvitations(),
      ]);
      setTrees(treeList);
      setInvitations(invitationList);
      const links: Record<number, FamilyTreeShareLink[]> = {};
      await Promise.all(treeList.map(async (tree) => {
        links[tree.tree_id] = await familyTreesApi.listShareLinks(tree.tree_id);
      }));
      setShareLinks(links);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const openTree = (treeId: number) => {
    setActiveTreeId(treeId);
    navigate(`/trees/${treeId}/tree`);
  };

  const createTree = async () => {
    const tree = await familyTreesApi.create({
      name: newTreeName.trim(),
      description: newTreeDescription.trim() || undefined,
    });
    setCreateOpen(false);
    setNewTreeName('');
    setNewTreeDescription('');
    setActiveTreeId(tree.tree_id);
    navigate(`/trees/${tree.tree_id}/tree`);
  };

  const respondToInvitation = async (invitationId: number, accept: boolean) => {
    if (accept) {
      await familyTreesApi.acceptInvitation(invitationId);
    } else {
      await familyTreesApi.declineInvitation(invitationId);
    }
    await load();
  };

  const inviteUser = async (treeId: number) => {
    const email = inviteEmails[treeId]?.trim();
    if (!email) return;
    await familyTreesApi.invite(treeId, { email });
    setInviteEmails((prev) => ({ ...prev, [treeId]: '' }));
    enqueueSnackbar('Invitation created', { variant: 'success' });
  };

  const createShareLink = async (treeId: number) => {
    const maxVisitsValue = shareMaxVisits[treeId]?.trim();
    const link = await familyTreesApi.createShareLink(treeId, {
      max_visits: maxVisitsValue ? Number(maxVisitsValue) : undefined,
    });
    setShareMaxVisits((prev) => ({ ...prev, [treeId]: '' }));
    setShareLinks((prev) => ({ ...prev, [treeId]: [link, ...(prev[treeId] || [])] }));
    await navigator.clipboard?.writeText(link.url).catch(() => undefined);
    enqueueSnackbar('Share link created', { variant: 'success' });
  };

  return (
    <Layout>
      <Box sx={{ width: '100%' }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 3 }}>
          <Typography variant="h4">Family Trees</Typography>
          <Stack direction="row" spacing={1}>
            <IconButton onClick={load} aria-label="Refresh" disabled={loading}>
              <Refresh />
            </IconButton>
            <Button startIcon={<Add />} variant="contained" onClick={() => setCreateOpen(true)}>
              New Tree
            </Button>
          </Stack>
        </Stack>

        {invitations.length > 0 && (
          <Paper sx={{ p: 2, mb: 3 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>Pending Invitations</Typography>
            <Stack spacing={1.5}>
              {invitations.map((invitation) => (
                <Stack key={invitation.invitation_id} direction={{ xs: 'column', sm: 'row' }} alignItems={{ sm: 'center' }} justifyContent="space-between" spacing={1}>
                  <Box>
                    <Typography fontWeight={700}>{invitation.tree_name}</Typography>
                    <Typography variant="body2" color="text.secondary">
                      Invited by {invitation.inviter_name || 'another user'}
                    </Typography>
                  </Box>
                  <Stack direction="row" spacing={1}>
                    <Button variant="contained" size="small" onClick={() => respondToInvitation(invitation.invitation_id, true)}>
                      Join
                    </Button>
                    <Button variant="outlined" size="small" onClick={() => respondToInvitation(invitation.invitation_id, false)}>
                      Decline
                    </Button>
                  </Stack>
                </Stack>
              ))}
            </Stack>
          </Paper>
        )}

        {trees.length === 0 && !loading && (
          <Alert severity="info" sx={{ mb: 3 }}>
            Create your first tree or accept an invitation to start collaborating.
          </Alert>
        )}

        <Grid container spacing={2}>
          {trees.map((tree) => (
            <Grid item xs={12} md={6} key={tree.tree_id}>
              <Paper sx={{ p: 2, height: '100%' }}>
                <Stack spacing={2}>
                  <Stack direction="row" justifyContent="space-between" alignItems="flex-start" spacing={2}>
                    <Box>
                      <Typography variant="h6">{tree.name}</Typography>
                      {tree.description && (
                        <Typography variant="body2" color="text.secondary">{tree.description}</Typography>
                      )}
                    </Box>
                    <Chip icon={<AccountTree />} label={`${tree.member_count} members`} size="small" />
                  </Stack>

                  <Stack direction="row" spacing={1} flexWrap="wrap">
                    <Button startIcon={<Login />} variant="contained" onClick={() => openTree(tree.tree_id)}>
                      Open
                    </Button>
                    <Chip label={tree.user_role || 'member'} size="small" />
                  </Stack>

                  <Divider />

                  <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
                    <TextField
                      size="small"
                      label="Invite user email"
                      value={inviteEmails[tree.tree_id] || ''}
                      onChange={(event) => setInviteEmails((prev) => ({ ...prev, [tree.tree_id]: event.target.value }))}
                      fullWidth
                    />
                    <Button startIcon={<PersonAdd />} variant="outlined" onClick={() => inviteUser(tree.tree_id)}>
                      Invite
                    </Button>
                  </Stack>

                  <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1}>
                    <TextField
                      size="small"
                      label="Max visits"
                      type="number"
                      value={shareMaxVisits[tree.tree_id] || ''}
                      onChange={(event) => setShareMaxVisits((prev) => ({ ...prev, [tree.tree_id]: event.target.value }))}
                      fullWidth
                    />
                    <Button startIcon={<Link />} variant="outlined" onClick={() => createShareLink(tree.tree_id)}>
                      Share
                    </Button>
                  </Stack>

                  {(shareLinks[tree.tree_id] || []).slice(0, 3).map((link) => (
                    <Box key={link.share_id} sx={{ p: 1, border: '1px solid', borderColor: 'divider', borderRadius: 1 }}>
                      <Typography variant="body2" sx={{ overflowWrap: 'anywhere' }}>{link.url}</Typography>
                      <Typography variant="caption" color="text.secondary">
                        {link.visit_count}{link.max_visits ? ` / ${link.max_visits}` : ''} visits
                        {link.expires_at ? `, expires ${link.expires_at}` : ''}
                      </Typography>
                    </Box>
                  ))}
                </Stack>
              </Paper>
            </Grid>
          ))}
        </Grid>
      </Box>

      <Dialog open={createOpen} onClose={() => setCreateOpen(false)} fullWidth maxWidth="sm">
        <DialogTitle>Create Family Tree</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField label="Tree name" value={newTreeName} onChange={(event) => setNewTreeName(event.target.value)} autoFocus />
            <TextField label="Description" value={newTreeDescription} onChange={(event) => setNewTreeDescription(event.target.value)} multiline rows={3} />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateOpen(false)}>Cancel</Button>
          <Button variant="contained" onClick={createTree} disabled={!newTreeName.trim()}>Create</Button>
        </DialogActions>
      </Dialog>
    </Layout>
  );
};

export default TreesPage;
