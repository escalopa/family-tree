import React, { useEffect, useRef } from 'react';
import { Box, Chip, Paper, Typography } from '@mui/material';
import { AccountTree, Hub, Link as LinkIcon } from '@mui/icons-material';
import * as d3 from 'd3';
import { FamilyGraph, FamilyGraphPerson, FamilyGraphUnit, Member } from '../types';
import { getGenderColor, getMemberPictureUrl } from '../utils/helpers';
import { useTheme } from '../contexts/ThemeContext';

interface FamilyGraphVisualizationProps {
  data: FamilyGraph;
  onNodeClick: (member: Member) => void;
  onSetRoot: (memberId: number) => void;
  currentRootId?: number;
}

interface GraphNode extends d3.SimulationNodeDatum {
  id: string;
  kind: 'person' | 'family';
  person?: FamilyGraphPerson;
  unit?: FamilyGraphUnit;
  level: number;
}

interface GraphLink extends d3.SimulationLinkDatum<GraphNode> {
  id: string;
  type: 'partner' | 'child';
  relationType?: string;
  status?: string;
  isInPath: boolean;
}

const personNodeId = (personId: number) => `person:${personId}`;
const familyNodeId = (familyUnitId: number) => `family:${familyUnitId}`;

const FamilyGraphVisualization: React.FC<FamilyGraphVisualizationProps> = ({
  data,
  onNodeClick,
  onSetRoot,
  currentRootId,
}) => {
  const { mode } = useTheme();
  const isDarkMode = mode === 'dark';
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!svgRef.current || !containerRef.current || !data) return;

    d3.select(svgRef.current).selectAll('*').remove();

    const width = containerRef.current.clientWidth || 1000;
    const height = containerRef.current.clientHeight || 760;
    const levels = calculateLevels(data);

    const nodes: GraphNode[] = [
      ...data.people.map((person) => ({
        id: personNodeId(person.member.member_id),
        kind: 'person' as const,
        person,
        level: levels.get(personNodeId(person.member.member_id)) || 0,
      })),
      ...data.family_units.map((unit) => ({
        id: familyNodeId(unit.family_unit_id),
        kind: 'family' as const,
        unit,
        level: levels.get(familyNodeId(unit.family_unit_id)) || 0.5,
      })),
    ];

    const links: GraphLink[] = data.edges.map((edge) => ({
      id: edge.edge_id,
      source: edge.source_id,
      target: edge.target_id,
      type: edge.type,
      relationType: edge.relation_type,
      status: edge.status,
      isInPath: edge.is_in_path,
    }));

    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height);

    const defs = svg.append('defs');
    defs.append('filter')
      .attr('id', 'graph-node-shadow')
      .append('feDropShadow')
      .attr('dx', 0)
      .attr('dy', 6)
      .attr('stdDeviation', 5)
      .attr('flood-opacity', isDarkMode ? 0.35 : 0.18);

    const g = svg.append('g');
    const linkLayer = g.append('g').attr('class', 'graph-links');
    const nodeLayer = g.append('g').attr('class', 'graph-nodes');

    const simulation = d3
      .forceSimulation<GraphNode>(nodes)
      .force(
        'link',
        d3.forceLink<GraphNode, GraphLink>(links)
          .id((node) => node.id)
          .distance((link) => {
            if (link.type === 'partner') return 58;
            return link.relationType === 'step' || link.relationType === 'adopted' ? 150 : 130;
          })
          .strength((link) => (link.type === 'partner' ? 0.95 : 0.72))
      )
      .force('charge', d3.forceManyBody().strength((node) => ((node as GraphNode).kind === 'family' ? -260 : -520)))
      .force('collision', d3.forceCollide<GraphNode>().radius((node) => (node.kind === 'family' ? 46 : 70)))
      .force('x', d3.forceX<GraphNode>(width / 2).strength(0.045))
      .force('y', d3.forceY<GraphNode>((node) => 90 + node.level * 145).strength(0.42))
      .stop();

    for (let i = 0; i < 420; i += 1) {
      simulation.tick();
    }

    const link = linkLayer
      .selectAll('path')
      .data(links)
      .join('path')
      .attr('fill', 'none')
      .attr('stroke', (edge) => edgeColor(edge, isDarkMode))
      .attr('stroke-width', (edge) => (edge.isInPath ? 5 : edge.type === 'partner' ? 3 : 2.4))
      .attr('stroke-dasharray', (edge) => {
        if (edge.isInPath) return 'none';
        if (edge.status === 'divorced' || edge.status === 'separated') return '8 6';
        if (edge.relationType === 'adopted' || edge.relationType === 'step') return '5 5';
        return 'none';
      })
      .attr('opacity', (edge) => (edge.isInPath ? 1 : isDarkMode ? 0.74 : 0.66))
      .attr('d', linkPath);

    const node = nodeLayer
      .selectAll('g')
      .data(nodes)
      .join('g')
      .attr('transform', (graphNode) => `translate(${graphNode.x || 0},${graphNode.y || 0})`)
      .style('cursor', (graphNode) => (graphNode.kind === 'person' ? 'pointer' : 'default'));

    node
      .filter((graphNode) => graphNode.kind === 'family')
      .each(function renderFamilyUnit(graphNode) {
        const unit = graphNode.unit!;
        const group = d3.select(this);
        const isInPath = Boolean(data.path_family_unit_ids?.includes(unit.family_unit_id));

        group.append('rect')
          .attr('x', -19)
          .attr('y', -19)
          .attr('width', 38)
          .attr('height', 38)
          .attr('rx', 7)
          .attr('transform', 'rotate(45)')
          .attr('fill', isInPath ? '#FF9800' : isDarkMode ? '#20314f' : '#F7FAFC')
          .attr('stroke', unit.status === 'divorced' ? '#F97316' : isDarkMode ? '#8EA4D2' : '#64748B')
          .attr('stroke-width', isInPath ? 3 : 2)
          .style('filter', 'url(#graph-node-shadow)');

        group.append('text')
          .attr('text-anchor', 'middle')
          .attr('dy', '0.35em')
          .attr('font-size', '16px')
          .attr('font-weight', 800)
          .attr('fill', isInPath ? '#111827' : isDarkMode ? '#DCE7FF' : '#334155')
          .text(unit.status === 'divorced' ? '//' : '+');
      });

    node
      .filter((graphNode) => graphNode.kind === 'person')
      .each(function renderPerson(graphNode) {
        const person = graphNode.person!;
        const member = person.member;
        const group = d3.select(this);
        const isCurrentRoot = currentRootId === member.member_id;
        const isInPath = person.is_in_path;
        const radius = 36;
        const genderColor = getGenderColor(member.gender);

        group.append('circle')
          .attr('r', radius)
          .attr('fill', isInPath ? (isDarkMode ? '#92400E' : '#FFF7ED') : isDarkMode ? '#111C2E' : '#FFFFFF')
          .attr('stroke', isCurrentRoot ? '#60A5FA' : genderColor)
          .attr('stroke-width', isCurrentRoot || isInPath ? 4 : 3)
          .style('filter', 'url(#graph-node-shadow)');

        const pictureUrl = getMemberPictureUrl(member.member_id, member.picture);
        if (pictureUrl) {
          const clipId = `family-graph-clip-${member.member_id}`;
          const localDefs = group.append('defs');
          localDefs.append('clipPath')
            .attr('id', clipId)
            .append('circle')
            .attr('r', radius - 5);

          group.append('image')
            .attr('xlink:href', pictureUrl)
            .attr('x', -(radius - 5))
            .attr('y', -(radius - 5))
            .attr('width', (radius - 5) * 2)
            .attr('height', (radius - 5) * 2)
            .attr('clip-path', `url(#${clipId})`)
            .style('pointer-events', 'none');
        } else {
          const initials = getInitials(member);
          group.append('text')
            .attr('text-anchor', 'middle')
            .attr('dy', '0.35em')
            .attr('font-size', '15px')
            .attr('font-weight', 800)
            .attr('fill', genderColor)
            .text(initials);
        }

        if (person.is_reference_candidate) {
          group.append('circle')
            .attr('cx', 26)
            .attr('cy', -26)
            .attr('r', 11)
            .attr('fill', isDarkMode ? '#38BDF8' : '#0284C7')
            .attr('stroke', isDarkMode ? '#0F172A' : '#FFFFFF')
            .attr('stroke-width', 2);

          group.append('text')
            .attr('x', 26)
            .attr('y', -26)
            .attr('dy', '0.35em')
            .attr('text-anchor', 'middle')
            .attr('font-size', '13px')
            .attr('font-weight', 900)
            .attr('fill', '#FFFFFF')
            .text('R');
        }

        const names = getNameLines(member);
        names.forEach((name, index) => {
          group.append('text')
            .attr('y', radius + 18 + index * 15)
            .attr('text-anchor', 'middle')
            .attr('font-size', index === 0 ? '12px' : '11px')
            .attr('font-weight', index === 0 ? 700 : 600)
            .attr('fill', isDarkMode ? '#F8FAFC' : '#0F172A')
            .text(truncateText(name, 18));
        });

        group.on('click', (event) => {
          event.stopPropagation();
          onNodeClick(member);
        });

        group.on('contextmenu', (event) => {
          event.preventDefault();
          event.stopPropagation();
          onSetRoot(member.member_id);
        });
      });

    const drag = d3.drag<SVGGElement, GraphNode>()
      .on('start', function dragStart(_event, graphNode) {
        graphNode.fx = graphNode.x;
        graphNode.fy = graphNode.y;
        d3.select(this).raise();
      })
      .on('drag', function dragging(event, graphNode) {
        graphNode.fx = event.x;
        graphNode.fy = event.y;
        graphNode.x = event.x;
        graphNode.y = event.y;
        d3.select(this).attr('transform', `translate(${event.x},${event.y})`);
        link.attr('d', linkPath);
      });

    node.call(drag as unknown as (selection: typeof node) => void);

    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.08, 3.4])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });

    svg.call(zoom);

    const bounds = g.node()?.getBBox();
    if (bounds && bounds.width > 0 && bounds.height > 0) {
      const scale = Math.min(width / bounds.width, height / bounds.height) * 0.78;
      const boundedScale = Math.max(0.12, Math.min(1.25, scale));
      const translate = [
        width / 2 - boundedScale * (bounds.x + bounds.width / 2),
        height / 2 - boundedScale * (bounds.y + bounds.height / 2),
      ];
      svg.call(zoom.transform, d3.zoomIdentity.translate(translate[0], translate[1]).scale(boundedScale));
    }

    function linkPath(edge: GraphLink) {
      const source = edge.source as GraphNode;
      const target = edge.target as GraphNode;
      const sx = source.x || 0;
      const sy = source.y || 0;
      const tx = target.x || 0;
      const ty = target.y || 0;
      const midY = (sy + ty) / 2;
      if (edge.type === 'partner') {
        const midX = (sx + tx) / 2;
        return `M ${sx},${sy} Q ${midX},${sy - 24} ${tx},${ty}`;
      }
      return `M ${sx},${sy} C ${sx},${midY} ${tx},${midY} ${tx},${ty}`;
    }
  }, [data, currentRootId, isDarkMode, onNodeClick, onSetRoot]);

  const familyCount = data.family_units.length;
  const referenceCount = data.references.length;

  return (
    <Paper
      ref={containerRef}
      sx={{
        position: 'relative',
        width: '100%',
        height: { xs: '680px', md: '820px' },
        overflow: 'hidden',
        bgcolor: isDarkMode ? '#07101f' : '#F8FAFC',
        borderRadius: 2,
      }}
    >
      <svg ref={svgRef} style={{ width: '100%', height: '100%' }} />

      <Box
        sx={{
          position: 'absolute',
          top: 14,
          left: 14,
          display: 'flex',
          gap: 1,
          flexWrap: 'wrap',
          maxWidth: 'calc(100% - 28px)',
        }}
      >
        <Chip icon={<AccountTree />} label={`${data.people.length} people`} size="small" />
        <Chip icon={<Hub />} label={`${familyCount} family units`} size="small" />
        {referenceCount > 0 && <Chip icon={<LinkIcon />} label={`${referenceCount} linked references`} size="small" />}
      </Box>

      <Box
        sx={{
          position: 'absolute',
          right: 14,
          bottom: 14,
          display: 'grid',
          gap: 0.5,
          p: 1,
          borderRadius: 1,
          bgcolor: isDarkMode ? 'rgba(15, 23, 42, 0.82)' : 'rgba(255, 255, 255, 0.88)',
          boxShadow: 2,
        }}
      >
        <LegendLine color={isDarkMode ? '#F472B6' : '#EC407A'} label="partners" />
        <LegendLine color={isDarkMode ? '#94A3B8' : '#475569'} label="children" />
        <LegendLine color="#FF9800" label="relation path" />
      </Box>
    </Paper>
  );
};

const LegendLine: React.FC<{ color: string; label: string }> = ({ color, label }) => (
  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
    <Box sx={{ width: 24, height: 3, bgcolor: color, borderRadius: 1 }} />
    <Typography variant="caption" color="text.secondary">{label}</Typography>
  </Box>
);

function calculateLevels(data: FamilyGraph) {
  const levels = new Map<string, number>();
  const childToUnits = new Map<number, number[]>();
  const unitByID = new Map(data.family_units.map((unit) => [unit.family_unit_id, unit]));

  data.family_units.forEach((unit) => {
    unit.child_ids.forEach((childID) => {
      const units = childToUnits.get(childID) || [];
      units.push(unit.family_unit_id);
      childToUnits.set(childID, units);
    });
  });

  data.people.forEach((person) => {
    levels.set(personNodeId(person.member.member_id), childToUnits.has(person.member.member_id) ? 2 : 0);
  });

  for (let iteration = 0; iteration < 12; iteration += 1) {
    let changed = false;
    data.family_units.forEach((unit) => {
      const partnerLevels = unit.partner_ids
        .map((personID) => levels.get(personNodeId(personID)))
        .filter((level): level is number => level !== undefined);
      const familyLevel = partnerLevels.length > 0 ? Math.max(...partnerLevels) + 0.75 : 0.75;
      const familyID = familyNodeId(unit.family_unit_id);
      if ((levels.get(familyID) ?? -1) < familyLevel) {
        levels.set(familyID, familyLevel);
        changed = true;
      }

      unit.child_ids.forEach((childID) => {
        const childIDKey = personNodeId(childID);
        const childLevel = familyLevel + 0.85;
        if ((levels.get(childIDKey) ?? -1) < childLevel) {
          levels.set(childIDKey, childLevel);
          changed = true;
        }

        const child = data.people.find((person) => person.member.member_id === childID);
        child?.partner_family_unit_ids?.forEach((partnerUnitID) => {
          const partnerUnit = unitByID.get(partnerUnitID);
          if (!partnerUnit) return;
          const partnerUnitKey = familyNodeId(partnerUnit.family_unit_id);
          if ((levels.get(partnerUnitKey) ?? -1) < childLevel + 0.75) {
            levels.set(partnerUnitKey, childLevel + 0.75);
            changed = true;
          }
        });
      });
    });
    if (!changed) break;
  }

  return levels;
}

function edgeColor(edge: GraphLink, isDarkMode: boolean) {
  if (edge.isInPath) return '#FF9800';
  if (edge.type === 'partner') return isDarkMode ? '#F472B6' : '#EC407A';
  if (edge.relationType === 'adopted') return '#22C55E';
  if (edge.relationType === 'step') return '#A855F7';
  return isDarkMode ? '#94A3B8' : '#475569';
}

function getInitials(member: Member) {
  const name = member.name || member.names?.en || member.names?.ar || '?';
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase();
}

function getNameLines(member: Member) {
  const names = [member.names?.ar, member.names?.en].filter((name): name is string => Boolean(name));
  if (names.length === 0 && member.name) names.push(member.name);
  return Array.from(new Set(names)).slice(0, 2);
}

function truncateText(text: string, maxLength: number) {
  if (text.length <= maxLength) return text;
  return `${text.slice(0, maxLength)}...`;
}

export default FamilyGraphVisualization;
