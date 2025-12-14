import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Container,
  Typography,
  Box,
  Button,
  TextField,
  Grid,
  Card,
  CardContent,
  Avatar,
  Chip,
  Paper,
  CircularProgress,
} from '@mui/material';
import { Layout } from '../components/common/Layout';
import { LoadingSpinner } from '../components/common/LoadingSpinner';
import { membersApi } from '../api/members';
import { MemberSearchParams, Member } from '../types/member';
import AddIcon from '@mui/icons-material/Add';
import { GENDER_COLORS } from '../utils/constants';

export const MembersPage: React.FC = () => {
  const [searchParams, setSearchParams] = useState<MemberSearchParams>({
    arabic_name: '',
    limit: 20,
  });
  const [allMembers, setAllMembers] = useState<Member[]>([]);
  const [nextCursor, setNextCursor] = useState<string | undefined>();
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['members', searchParams],
    queryFn: async () => {
      const result = await membersApi.searchMembers(searchParams);
      setAllMembers(result.members);
      setNextCursor(result.next_cursor);
      return result;
    },
    enabled: Object.values(searchParams).some((v) => v !== '' && v !== undefined && v !== 'cursor'),
  });

  const handleLoadMore = async () => {
    if (!nextCursor || isLoadingMore) return;

    setIsLoadingMore(true);
    try {
      const result = await membersApi.searchMembers({
        ...searchParams,
        cursor: nextCursor,
      });
      setAllMembers([...allMembers, ...result.members]);
      setNextCursor(result.next_cursor);
    } catch (error) {
      console.error('Failed to load more members:', error);
    } finally {
      setIsLoadingMore(false);
    }
  };

  return (
    <Layout>
      <Container maxWidth="lg">
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h4">Manage Members</Typography>
          <Button variant="contained" startIcon={<AddIcon />}>
            Add Member
          </Button>
        </Box>

        <Paper sx={{ p: 3, mb: 3 }}>
          <Grid container spacing={2}>
            <Grid item xs={12} md={4}>
              <TextField
                fullWidth
                label="Arabic Name"
                value={searchParams.arabic_name || ''}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, arabic_name: e.target.value || undefined })
                }
              />
            </Grid>
            <Grid item xs={12} md={4}>
              <TextField
                fullWidth
                label="English Name"
                value={searchParams.english_name || ''}
                onChange={(e) =>
                  setSearchParams({ ...searchParams, english_name: e.target.value || undefined })
                }
              />
            </Grid>
            <Grid item xs={12} md={4}>
              <TextField
                fullWidth
                select
                label="Gender"
                value={searchParams.gender || ''}
                onChange={(e) =>
                  setSearchParams({
                    ...searchParams,
                    gender: (e.target.value as 'M' | 'F' | 'N') || undefined,
                  })
                }
                SelectProps={{ native: true }}
              >
                <option value="">All</option>
                <option value="M">Male</option>
                <option value="F">Female</option>
                <option value="N">Other</option>
              </TextField>
            </Grid>
          </Grid>
        </Paper>

        {isLoading ? (
          <LoadingSpinner />
        ) : (
          <>
            <Grid container spacing={2}>
              {allMembers?.map((member) => (
                <Grid item xs={12} md={6} key={member.member_id}>
                  <Card>
                    <CardContent>
                      <Box sx={{ display: 'flex', gap: 2 }}>
                        <Avatar
                          src={member.picture || undefined}
                          sx={{ bgcolor: GENDER_COLORS[member.gender], width: 56, height: 56 }}
                        >
                          {member.arabic_name[0]}
                        </Avatar>
                        <Box sx={{ flexGrow: 1 }}>
                          <Typography variant="h6">{member.arabic_name}</Typography>
                          <Typography variant="body2" color="textSecondary">
                            {member.english_name}
                          </Typography>
                          <Box sx={{ mt: 1 }}>
                            <Chip
                              label={member.gender === 'M' ? 'Male' : 'Female'}
                              size="small"
                              sx={{ mr: 1 }}
                            />
                            {member.is_married && <Chip label="Married" size="small" />}
                          </Box>
                        </Box>
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
              ))}
            </Grid>

            {nextCursor && (
              <Box sx={{ display: 'flex', justifyContent: 'center', mt: 3 }}>
                <Button
                  variant="outlined"
                  onClick={handleLoadMore}
                  disabled={isLoadingMore}
                  startIcon={isLoadingMore ? <CircularProgress size={20} /> : null}
                >
                  {isLoadingMore ? 'Loading...' : 'Load More'}
                </Button>
              </Box>
            )}
          </>
        )}
      </Container>
    </Layout>
  );
};


