import React, { useEffect, useRef, useState } from 'react';
import { Paper, Dialog, DialogTitle, DialogContent, Typography } from '@mui/material';
import * as d3 from 'd3';
import { TreeNode, Member } from '../types';
import { getGenderColor, getMemberPictureUrl } from '../utils/helpers';
import { useTheme } from '../contexts/ThemeContext';
import '../styles/tree-theme.css';

interface TreeVisualizationProps {
  data: TreeNode;
  onNodeClick: (member: Member) => void;
  onSetRoot: (memberId: number) => void;
  currentRootId?: number;
}

interface D3Node extends d3.HierarchyPointNode<TreeNode> {
  _children?: D3Node[];
}

interface NodePosition {
  memberId: number;
  x: number;
  y: number;
  member: Member;
  isInPath: boolean;
}

const TreeVisualization: React.FC<TreeVisualizationProps> = ({ data, onNodeClick, onSetRoot, currentRootId }) => {
  const { mode } = useTheme();
  const isDarkMode = mode === 'dark';
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [infoDialogOpen, setInfoDialogOpen] = useState(false);

  useEffect(() => {
    if (!svgRef.current || !containerRef.current || !data) return;

    // Clear previous content
    d3.select(svgRef.current).selectAll('*').remove();

    const container = containerRef.current;
    const width = container.clientWidth;
    const height = container.clientHeight;

    // Circular node configuration
    const nodeRadius = 40;
    const horizontalSpacing = 80;
    const verticalSpacing = 180;
    const spouseOffset = nodeRadius * 2 + 30; // Distance between spouses

    // Create SVG
    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height);

    const g = svg.append('g').attr('transform', `translate(${width / 2},${nodeRadius + 40})`);

    // Create tree layout
    const treeLayout = d3
      .tree<TreeNode>()
      .nodeSize([nodeRadius * 2 + horizontalSpacing, verticalSpacing])
      .separation((a, b) => {
        // Increase separation for nodes with spouses
        const aHasSpouses = a.data.member.spouses && a.data.member.spouses.length > 0;
        const bHasSpouses = b.data.member.spouses && b.data.member.spouses.length > 0;
        return aHasSpouses || bHasSpouses ? 2.5 : 1.5;
      });

    // Create hierarchy
    const root = d3.hierarchy(data);
    const treeData = treeLayout(root);

    // Get all nodes and links from tree
    const treeNodes = treeData.descendants();
    const treeLinks = treeData.links();

    // Build node positions map including spouses
    const nodePositions: NodePosition[] = [];
    const nodePositionMap = new Map<number, NodePosition>();
    const links: Array<{ sourceId: number; targetId: number; type: 'parent' | 'spouse' }> = [];

    treeNodes.forEach((node) => {
      const member = node.data.member;
      const memberId = member.member_id;

      // Add main member position
      if (!nodePositionMap.has(memberId)) {
        const position: NodePosition = {
          memberId,
          x: node.x,
          y: node.y,
          member,
          isInPath: node.data.is_in_path || false
        };
        nodePositions.push(position);
        nodePositionMap.set(memberId, position);
      }

      // Add spouse positions next to the member
      if (member.spouses && member.spouses.length > 0) {
        member.spouses.forEach((spouse, index) => {
          if (!nodePositionMap.has(spouse.member_id)) {
            const spousePosition: NodePosition = {
              memberId: spouse.member_id,
              x: node.x + spouseOffset * (index + 1),
              y: node.y,
              member: {
                member_id: spouse.member_id,
                name: spouse.name,
                names: {},
                gender: spouse.gender,
                picture: spouse.picture,
                date_of_birth: null,
                date_of_death: null,
                father_id: null,
                mother_id: null,
                nicknames: [],
                profession: null,
                version: 0,
                is_married: true,
                generation_level: member.generation_level
              },
              isInPath: false
            };
            nodePositions.push(spousePosition);
            nodePositionMap.set(spouse.member_id, spousePosition);

            // Add spouse link
            links.push({
              sourceId: memberId,
              targetId: spouse.member_id,
              type: 'spouse'
            });
          }
        });
      }
    });

    // Add parent-child links
    treeLinks.forEach((link) => {
      const sourceId = (link.source as D3Node).data.member.member_id;
      const targetId = (link.target as D3Node).data.member.member_id;

      links.push({
        sourceId,
        targetId,
        type: 'parent'
      });
    });

    // Draw links
    const linkGroup = g.append('g').attr('class', 'links');

    links.forEach((link) => {
      const sourcePos = nodePositionMap.get(link.sourceId);
      const targetPos = nodePositionMap.get(link.targetId);

      if (!sourcePos || !targetPos) return;

      const path = linkGroup
        .append('path')
        .attr('class', `link ${link.type} tree-link-organic`)
        .attr('d', () => {
          if (link.type === 'spouse') {
            // Horizontal curved line for spouses
            const midX = (sourcePos.x + targetPos.x) / 2;
            const curveOffset = 15;
            return `M ${sourcePos.x},${sourcePos.y}
                    Q ${midX},${sourcePos.y - curveOffset} ${targetPos.x},${targetPos.y}`;
          } else {
            // Vertical curved line from parent to child
            const sourceX = sourcePos.x;
            const sourceY = sourcePos.y + nodeRadius;
            const targetX = targetPos.x;
            const targetY = targetPos.y - nodeRadius;
            const midY = (sourceY + targetY) / 2;
            const curveOffset = Math.abs(targetX - sourceX) * 0.15;

            return `M ${sourceX},${sourceY}
                    C ${sourceX},${midY - curveOffset}
                      ${sourceX},${midY + curveOffset}
                      ${sourceX},${midY}
                    L ${targetX},${midY}
                    C ${targetX},${midY + curveOffset}
                      ${targetX},${midY + curveOffset * 2}
                      ${targetX},${targetY}`;
          }
        })
        .attr('fill', 'none')
        .attr('stroke',
          link.type === 'spouse' ? (isDarkMode ? '#F472B6' : '#EC407A') :
          (isDarkMode ? '#6B7280' : '#424242')
        )
        .attr('stroke-width', link.type === 'spouse' ? 3 : 2.5)
        .attr('opacity', isDarkMode ? 0.7 : 0.6);

      // Highlight path links
      if (sourcePos.isInPath && targetPos.isInPath && link.type === 'parent') {
        path
          .attr('stroke', isDarkMode ? '#FFA726' : '#FF9800')
          .attr('stroke-width', 4)
          .attr('opacity', 1)
          .attr('class', 'tree-link-highlighted');
      }
    });

    // Draw circular nodes
    const nodeGroup = g.append('g').attr('class', 'nodes');

    nodePositions.forEach((posData) => {
      const { memberId, x, y, member, isInPath } = posData;

      const nodeG = nodeGroup
        .append('g')
        .attr('class', 'node tree-node-organic')
        .attr('transform', `translate(${x},${y})`)
        .style('cursor', 'pointer');

      const isCurrentRoot = currentRootId !== undefined && memberId === currentRootId;

      // Background circle
      nodeG
        .append('circle')
        .attr('r', nodeRadius)
        .attr('fill', () => {
          if (isCurrentRoot) return isDarkMode ? '#1E3A8A' : '#E3F2FD';
          if (isInPath) return isDarkMode ? '#92400E' : '#FFF3E0';
          return isDarkMode ? '#1E293B' : '#FFFFFF';
        })
        .attr('stroke', isCurrentRoot ? (isDarkMode ? '#60A5FA' : '#1976D2') : getGenderColor(member.gender))
        .attr('stroke-width', isCurrentRoot ? 5 : isInPath ? 4 : 3)
        .style('filter', isCurrentRoot
          ? 'drop-shadow(0 0 12px rgba(25, 118, 210, 0.5))'
          : 'drop-shadow(0 4px 12px rgba(0, 0, 0, 0.1))')
        .attr('class', isCurrentRoot ? 'root-node-glow' : '');

      // Growth rings
      nodeG.append('circle')
        .attr('r', nodeRadius + 6)
        .attr('fill', 'none')
        .attr('stroke', getGenderColor(member.gender))
        .attr('stroke-width', 1)
        .attr('opacity', isDarkMode ? 0.2 : 0.15)
        .style('pointer-events', 'none');

      nodeG.append('circle')
        .attr('r', nodeRadius + 10)
        .attr('fill', 'none')
        .attr('stroke', getGenderColor(member.gender))
        .attr('stroke-width', 0.5)
        .attr('opacity', isDarkMode ? 0.15 : 0.1)
        .style('pointer-events', 'none');

      // Avatar or initials
      const pictureUrl = getMemberPictureUrl(member.member_id, member.picture);
      const avatarRadius = nodeRadius - 5;

      if (pictureUrl) {
        const clipId = `clip-tree-${memberId}`;
        const defs = nodeG.append('defs');
        defs.append('clipPath')
          .attr('id', clipId)
          .append('circle')
          .attr('r', avatarRadius);

        nodeG.append('image')
          .attr('xlink:href', pictureUrl)
          .attr('x', -avatarRadius)
          .attr('y', -avatarRadius)
          .attr('width', avatarRadius * 2)
          .attr('height', avatarRadius * 2)
          .attr('clip-path', `url(#${clipId})`)
          .style('pointer-events', 'none');
      } else {
        const initials = member.name ? member.name.split(' ').map(n => n[0]).slice(0, 2).join('').toUpperCase() : '?';
        nodeG.append('text')
          .text(initials)
          .attr('text-anchor', 'middle')
          .attr('dy', '0.35em')
          .attr('font-size', nodeRadius * 0.5)
          .attr('font-weight', 'bold')
          .attr('fill', getGenderColor(member.gender))
          .style('pointer-events', 'none')
          .style('text-shadow', isDarkMode
            ? '0 1px 2px rgba(0, 0, 0, 0.8)'
            : '0 1px 2px rgba(255, 255, 255, 0.8)');
      }

      // Name label
      const name = member.name || 'Unknown';
      nodeG
        .append('text')
        .attr('y', nodeRadius + 18)
        .attr('text-anchor', 'middle')
        .attr('font-size', '13px')
        .attr('font-weight', '600')
        .attr('fill', isDarkMode ? '#F1F5F9' : '#000')
        .attr('class', isDarkMode ? 'tree-label-dark' : 'tree-label')
        .text(truncateText(name, 15))
        .style('pointer-events', 'none');

      // Generation badge
      if (member.generation_level !== undefined) {
        const badgeSize = 22;
        const badgeX = nodeRadius * 0.6;
        const badgeY = -nodeRadius * 0.6;

        nodeG
          .append('circle')
          .attr('cx', badgeX)
          .attr('cy', badgeY)
          .attr('r', badgeSize / 2)
          .attr('fill', isDarkMode ? '#60A5FA' : '#2196F3')
          .attr('class', 'generation-badge')
          .style('filter', 'drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3))')
          .style('pointer-events', 'none');

        nodeG
          .append('text')
          .attr('x', badgeX)
          .attr('y', badgeY)
          .attr('dy', '0.35em')
          .attr('text-anchor', 'middle')
          .attr('font-size', '11px')
          .attr('font-weight', 'bold')
          .attr('fill', 'white')
          .text(member.generation_level)
          .style('pointer-events', 'none')
          .style('text-shadow', '0 1px 2px rgba(0, 0, 0, 0.5)');
      }

      // Married indicator removed - visible from spouse edges

      // Click handlers
      nodeG.on('click', (event) => {
        event.stopPropagation();
        onNodeClick(member);
      });

      nodeG.on('contextmenu', (event) => {
        event.preventDefault();
        event.stopPropagation();
        onSetRoot(memberId);
      });

      // No hover animations - completely stable
      nodeG.on('mouseenter', function () {
        d3.select(this).style('opacity', 0.85);
      });

      nodeG.on('mouseleave', function () {
        d3.select(this).style('opacity', 1);
      });
    });

    // Zoom functionality
    const zoomBehavior = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 3])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });

    svg.call(zoomBehavior as any);

    // Center the view
    const bounds = g.node()?.getBBox();
    if (bounds) {
      const fullWidth = bounds.width;
      const fullHeight = bounds.height;
      const midX = bounds.x + fullWidth / 2;
      const midY = bounds.y + fullHeight / 2;

      const scale = Math.min(width / fullWidth, height / fullHeight) * 0.75;
      const translate = [width / 2 - scale * midX, height / 2 - scale * midY];

      svg.call(
        zoomBehavior.transform as any,
        d3.zoomIdentity.translate(translate[0], translate[1]).scale(scale)
      );
    }
  }, [data, onNodeClick, onSetRoot, currentRootId, isDarkMode]);

  // Zoom handlers removed

  return (
    <Paper
      ref={containerRef}
      className={`tree-container ${isDarkMode ? 'tree-container-dark tree-canvas-dark' : 'tree-canvas-light'} tree-fade-in`}
      sx={{
        position: 'relative',
        width: '100%',
        height: '800px',
        overflow: 'hidden',
        backgroundColor: isDarkMode ? '#0a0e1a' : '#fdfbf7',
      }}
    >
      <svg ref={svgRef} style={{ width: '100%', height: '100%' }} />

      {/* Controls - Info and zoom controls removed */}

      {/* Info Dialog */}
      <Dialog open={infoDialogOpen} onClose={() => setInfoDialogOpen(false)}>
        <DialogTitle>Tree View - Instructions</DialogTitle>
        <DialogContent>
          <Typography variant="body2" paragraph>
            <strong>Navigation:</strong>
          </Typography>
          <Typography variant="body2" component="div" sx={{ mb: 2, pl: 2 }}>
            • Click on a node to view detailed member information<br />
            • Right-click on a node to set it as the tree root<br />
            • Drag to pan the view<br />
            • Scroll to zoom in/out<br />
            • Siblings are connected through their shared parents
          </Typography>

          <Typography variant="body2" paragraph>
            <strong>Visual Indicators:</strong>
          </Typography>
          <Typography variant="body2" component="div" sx={{ pl: 2 }}>
            • <span style={{ color: '#1976D2', fontWeight: 'bold' }}>Blue border</span> = Current root node<br />
            • <span style={{ color: '#EC407A', fontWeight: 'bold' }}>Pink lines</span> = Spouse connections<br />
            • <span style={{ color: '#424242', fontWeight: 'bold' }}>Black lines</span> = Parent-child relationships<br />
            • <span style={{ color: '#FF9800', fontWeight: 'bold' }}>Orange highlights</span> = Relation path<br />
            • <span style={{ color: '#2196F3', fontWeight: 'bold' }}>Blue badge</span> = Generation number
          </Typography>
        </DialogContent>
      </Dialog>
    </Paper>
  );
};

// Helper function to truncate text
function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

export default TreeVisualization;
