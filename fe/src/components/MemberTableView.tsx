import React, { useState, useMemo, useEffect, useRef } from 'react';
import {
  Box,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TableSortLabel,
  Avatar,
  Chip,
  TextField,
  InputAdornment,
  TablePagination,
  CircularProgress,
  Typography,
} from '@mui/material';
import { Search } from '@mui/icons-material';
import { MemberListItem } from '../types';
import { getGenderColor, formatDate, getMemberPictureUrl } from '../utils/helpers';

interface MemberTableViewProps {
  members: MemberListItem[];
  onViewMember: (member: MemberListItem) => void;
  loading?: boolean;
  loadingMore?: boolean;
  hasMore?: boolean;
  onLoadMore?: () => void;
}

type SortField = 'arabic_name' | 'english_name' | 'gender' | 'date_of_birth';
type SortOrder = 'asc' | 'desc';

const MemberTableView: React.FC<MemberTableViewProps> = ({
  members,
  onViewMember,
  loading = false,
  loadingMore = false,
  hasMore = false,
  onLoadMore
}) => {
  const [sortField, setSortField] = useState<SortField>('date_of_birth');
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc');
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(25);
  const loadMoreRef = useRef<HTMLDivElement>(null);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('asc');
    }
  };

  const filteredAndSortedMembers = useMemo(() => {
    // Filter by search query
    let filtered = members;
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = members.filter(
        (member) =>
          member.arabic_name.toLowerCase().includes(query) ||
          member.english_name.toLowerCase().includes(query)
      );
    }

    // Sort
    const sorted = [...filtered].sort((a, b) => {
      let aValue: any;
      let bValue: any;

      switch (sortField) {
        case 'arabic_name':
          aValue = a.arabic_name;
          bValue = b.arabic_name;
          break;
        case 'english_name':
          aValue = a.english_name;
          bValue = b.english_name;
          break;
        case 'gender':
          aValue = a.gender;
          bValue = b.gender;
          break;
        case 'date_of_birth':
          aValue = a.date_of_birth ? new Date(a.date_of_birth).getTime() : 0;
          bValue = b.date_of_birth ? new Date(b.date_of_birth).getTime() : 0;
          break;
        default:
          aValue = a.member_id;
          bValue = b.member_id;
      }

      // Handle null values
      if (aValue === null || aValue === 0) return 1;
      if (bValue === null || bValue === 0) return -1;

      if (sortOrder === 'asc') {
        return aValue > bValue ? 1 : -1;
      } else {
        return aValue < bValue ? 1 : -1;
      }
    });

    return sorted;
  }, [members, searchQuery, sortField, sortOrder]);

  const paginatedMembers = useMemo(() => {
    const start = page * rowsPerPage;
    return filteredAndSortedMembers.slice(start, start + rowsPerPage);
  }, [filteredAndSortedMembers, page, rowsPerPage]);

  const handleChangePage = (_: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  // Infinite scroll observer for cursor-based pagination
  useEffect(() => {
    if (!loadMoreRef.current || !hasMore || loadingMore || !onLoadMore) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loadingMore) {
          onLoadMore();
        }
      },
      { threshold: 0.1 }
    );

    observer.observe(loadMoreRef.current);

    return () => {
      if (loadMoreRef.current) {
        observer.unobserve(loadMoreRef.current);
      }
    };
  }, [hasMore, loadingMore, onLoadMore]);

  return (
    <Paper sx={{ width: '100%', position: 'relative', minHeight: '400px' }}>
      {/* Search Bar */}
      <Box sx={{ p: 2 }}>
        <TextField
          fullWidth
          placeholder="Search by name..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value);
            setPage(0); // Reset to first page on search
          }}
          disabled={loading}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search />
              </InputAdornment>
            ),
          }}
        />
      </Box>

      {/* Loading Indicator */}
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

      {/* Table */}
      <TableContainer sx={{ opacity: loading && members.length === 0 ? 0.6 : 1 }}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Photo</TableCell>
              <TableCell>
                <TableSortLabel
                  active={sortField === 'arabic_name'}
                  direction={sortField === 'arabic_name' ? sortOrder : 'asc'}
                  onClick={() => handleSort('arabic_name')}
                >
                  Arabic Name
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={sortField === 'english_name'}
                  direction={sortField === 'english_name' ? sortOrder : 'asc'}
                  onClick={() => handleSort('english_name')}
                >
                  English Name
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={sortField === 'gender'}
                  direction={sortField === 'gender' ? sortOrder : 'asc'}
                  onClick={() => handleSort('gender')}
                >
                  Gender
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={sortField === 'date_of_birth'}
                  direction={sortField === 'date_of_birth' ? sortOrder : 'asc'}
                  onClick={() => handleSort('date_of_birth')}
                >
                  Birth Date
                </TableSortLabel>
              </TableCell>
              <TableCell>Status</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {(!members || members.length === 0) && !loading && (
              <TableRow>
                <TableCell colSpan={6} align="center" sx={{ py: 4, color: 'text.secondary' }}>
                  {searchQuery
                    ? 'No members found matching your search'
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
            {paginatedMembers.map((member) => (
              <TableRow
                key={member.member_id}
                hover
                sx={{ cursor: 'pointer' }}
                onClick={() => onViewMember(member)}
              >
                <TableCell>
                  <Avatar
                    src={getMemberPictureUrl(member.member_id, member.picture) || undefined}
                    sx={{ bgcolor: getGenderColor(member.gender) }}
                  >
                    {member.english_name.charAt(0)}
                  </Avatar>
                </TableCell>
                <TableCell>
                  <Box sx={{ fontWeight: 500 }}>{member.arabic_name}</Box>
                </TableCell>
                <TableCell>{member.english_name}</TableCell>
                <TableCell>
                  <Chip
                    label={member.gender === 'M' ? 'Male' : member.gender === 'F' ? 'Female' : 'Other'}
                    size="small"
                    sx={{
                      bgcolor: getGenderColor(member.gender),
                      color: 'white',
                    }}
                  />
                </TableCell>
                <TableCell>
                  {member.date_of_birth ? formatDate(member.date_of_birth) : '-'}
                </TableCell>
                <TableCell>
                  {member.is_married && <Chip label="Married" size="small" color="secondary" />}
                  {member.date_of_death && (
                    <Chip
                      label={`Deceased (${formatDate(member.date_of_death)})`}
                      size="small"
                      sx={{ ml: 0.5 }}
                    />
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Pagination */}
      <TablePagination
        component="div"
        count={filteredAndSortedMembers.length}
        page={page}
        onPageChange={handleChangePage}
        rowsPerPage={rowsPerPage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        rowsPerPageOptions={[10, 25, 50, 100]}
      />

      {/* Load More Sentinel for cursor-based pagination */}
      {hasMore && onLoadMore && (
        <Box
          ref={loadMoreRef}
          sx={{ display: 'flex', justifyContent: 'center', py: 2, minHeight: '60px' }}
        >
          {loadingMore && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <CircularProgress size={24} />
              <Typography variant="body2" color="text.secondary">
                Loading more members...
              </Typography>
            </Box>
          )}
        </Box>
      )}
    </Paper>
  );
};

export default MemberTableView;
