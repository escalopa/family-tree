import React, { useState, useMemo } from 'react';
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
  IconButton,
  Tooltip,
  TablePagination,
} from '@mui/material';
import { Search, Visibility, AccountTree } from '@mui/icons-material';
import { Member } from '../types';
import { getGenderColor, formatDate, getMemberPictureUrl } from '../utils/helpers';

interface MemberTableViewProps {
  members: Member[];
  onViewMember: (member: Member) => void;
  onSetRoot: (memberId: number) => void;
}

type SortField = 'arabic_name' | 'english_name' | 'gender' | 'date_of_birth' | 'age';
type SortOrder = 'asc' | 'desc';

const MemberTableView: React.FC<MemberTableViewProps> = ({ members, onViewMember, onSetRoot }) => {
  const [sortField, setSortField] = useState<SortField>('date_of_birth');
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc');
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(25);

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
          member.english_name.toLowerCase().includes(query) ||
          (member.nicknames && member.nicknames.some((n) => n.toLowerCase().includes(query))) ||
          (member.profession && member.profession.toLowerCase().includes(query))
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
        case 'age':
          aValue = a.age || 0;
          bValue = b.age || 0;
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

  return (
    <Paper sx={{ width: '100%' }}>
      {/* Search Bar */}
      <Box sx={{ p: 2 }}>
        <TextField
          fullWidth
          placeholder="Search by name, nickname, or profession..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value);
            setPage(0); // Reset to first page on search
          }}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search />
              </InputAdornment>
            ),
          }}
        />
      </Box>

      {/* Table */}
      <TableContainer>
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
              <TableCell>Profession</TableCell>
              <TableCell align="center">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
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
                  {member.nicknames && member.nicknames.length > 0 && (
                    <Box sx={{ fontSize: '0.75rem', color: 'text.secondary', mt: 0.5 }}>
                      {member.nicknames.join(', ')}
                    </Box>
                  )}
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
                <TableCell>{member.profession || '-'}</TableCell>
                <TableCell align="center">
                  <Tooltip title="View Details">
                    <IconButton
                      size="small"
                      onClick={(e) => {
                        e.stopPropagation();
                        onViewMember(member);
                      }}
                    >
                      <Visibility fontSize="small" />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Set as Tree Root">
                    <IconButton
                      size="small"
                      onClick={(e) => {
                        e.stopPropagation();
                        onSetRoot(member.member_id);
                      }}
                    >
                      <AccountTree fontSize="small" />
                    </IconButton>
                  </Tooltip>
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
    </Paper>
  );
};

export default MemberTableView;
