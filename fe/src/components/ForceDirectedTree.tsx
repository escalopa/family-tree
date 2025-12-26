import React, { useEffect, useRef, useState } from 'react';
import { Box, Paper, IconButton, Tooltip, Slider, Typography, Dialog, DialogTitle, DialogContent } from '@mui/material';
import { PlayArrow, Pause } from '@mui/icons-material';
import * as d3 from 'd3';
import { TreeNode, Member } from '../types';
import { getGenderColor, getMemberPictureUrl } from '../utils/helpers';
import { useTheme } from '../contexts/ThemeContext';
import '../styles/tree-theme.css';

interface ForceDirectedTreeProps {
  data: TreeNode;
  onNodeClick: (member: Member) => void;
  onSetRoot: (memberId: number) => void;
  currentRootId?: number;
}

interface GraphNode extends d3.SimulationNodeDatum {
  id: number;
  data: Member;
  isSpouse?: boolean;
  isSibling?: boolean;
  isInPath?: boolean;
}

interface GraphLink extends d3.SimulationLinkDatum<GraphNode> {
  type: 'parent' | 'spouse';
}

const ForceDirectedTree: React.FC<ForceDirectedTreeProps> = ({
  data,
  onNodeClick,
  onSetRoot,
  currentRootId,
}) => {
  const { mode } = useTheme();
  const isDarkMode = mode === 'dark';
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [simulation, setSimulation] = useState<d3.Simulation<GraphNode, GraphLink> | null>(null);
  const [isPlaying, setIsPlaying] = useState(true);
  const [linkDistance, setLinkDistance] = useState(150);
  const [chargeStrength, setChargeStrength] = useState(-300);
  const [infoDialogOpen, setInfoDialogOpen] = useState(false);

  useEffect(() => {
    if (!svgRef.current || !containerRef.current || !data) return;

    const container = containerRef.current;
    const width = container.clientWidth;
    const height = container.clientHeight;

    // Clear previous content
    d3.select(svgRef.current).selectAll('*').remove();

    // Convert tree to graph structure with de-duplication
    const nodeMap = new Map<number, GraphNode>();
    const links: GraphLink[] = [];
    const linkSet = new Set<string>(); // To avoid duplicate links

    const addNode = (member: Member, isInPath: boolean = false): GraphNode => {
      if (!nodeMap.has(member.member_id)) {
        const node: GraphNode = {
          id: member.member_id,
          data: member,
          isInPath: isInPath,
        };
        nodeMap.set(member.member_id, node);
      } else {
        // Update isInPath if this occurrence is in the path
        const existingNode = nodeMap.get(member.member_id)!;
        if (isInPath) {
          existingNode.isInPath = true;
        }
      }
      return nodeMap.get(member.member_id)!;
    };

    const addLink = (sourceId: number, targetId: number, type: 'parent' | 'spouse') => {
      const linkKey = `${sourceId}-${targetId}-${type}`;
      const reverseLinkKey = `${targetId}-${sourceId}-${type}`;

      // Avoid duplicate links (including reverse direction for spouse)
      if (!linkSet.has(linkKey) && !linkSet.has(reverseLinkKey)) {
        linkSet.add(linkKey);
        links.push({ source: sourceId, target: targetId, type });
      }
    };

    const processNode = (treeNode: TreeNode) => {
      const node = addNode(treeNode.member, treeNode.is_in_path || false);

      // Add spouse nodes from member's spouses array
      if (treeNode.member.spouses) {
        treeNode.member.spouses.forEach((spouse) => {
          const spouseMember: Member = {
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
            generation_level: treeNode.member.generation_level
          };
          addNode(spouseMember, false);
          addLink(node.id, spouse.member_id, 'spouse');
        });
      }

      // Add children
      if (treeNode.children) {
        treeNode.children.forEach((child) => {
          processNode(child);
          addLink(node.id, child.member.member_id, 'parent');
        });
      }
    };

    processNode(data);

    // Convert map to array
    const nodes = Array.from(nodeMap.values());

    // Create SVG
    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height);

    // Add gradient definitions for generation layers
    const defs = svg.append('defs');

    // Light mode gradient
    const lightGradient = defs.append('radialGradient')
      .attr('id', 'generationGradientLight')
      .attr('cx', '50%')
      .attr('cy', '50%')
      .attr('r', '50%');
    lightGradient.append('stop')
      .attr('offset', '0%')
      .attr('stop-color', '#8B4513')
      .attr('stop-opacity', '0.05');
    lightGradient.append('stop')
      .attr('offset', '100%')
      .attr('stop-color', '#228B22')
      .attr('stop-opacity', '0');

    // Dark mode gradient
    const darkGradient = defs.append('radialGradient')
      .attr('id', 'generationGradientDark')
      .attr('cx', '50%')
      .attr('cy', '50%')
      .attr('r', '50%');
    darkGradient.append('stop')
      .attr('offset', '0%')
      .attr('stop-color', '#4CAF50')
      .attr('stop-opacity', '0.1');
    darkGradient.append('stop')
      .attr('offset', '100%')
      .attr('stop-color', '#2E7D32')
      .attr('stop-opacity', '0');

    const g = svg.append('g');

    // Create force simulation
    const sim = d3
      .forceSimulation<GraphNode>(nodes)
      .force(
        'link',
        d3
          .forceLink<GraphNode, GraphLink>(links)
          .id((d) => d.id)
          .distance((d) => {
            if (d.type === 'spouse') return linkDistance * 0.5;
            return linkDistance;
          })
          .strength(1)
      )
      .force('charge', d3.forceManyBody().strength(chargeStrength))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(50));

    setSimulation(sim);

    // Draw links with organic styling
    const link = g
      .append('g')
      .selectAll('line')
      .data(links)
      .join('line')
      .attr('class', 'tree-link-organic')
      .attr('stroke', (d) => {
        if (d.type === 'spouse') return isDarkMode ? '#F472B6' : '#EC407A';
        return isDarkMode ? '#6B7280' : '#424242';
      })
      .attr('stroke-width', (d) => {
        if (d.type === 'spouse') return 3;
        return 2.5;
      })
      .attr('stroke-dasharray', 'none')
      .attr('opacity', isDarkMode ? 0.7 : 0.6);

    // Draw nodes
    const node = g
      .append('g')
      .selectAll('g')
      .data(nodes)
      .join('g')
      .call(
        d3
          .drag<SVGGElement, GraphNode>()
          .on('start', (event, d) => {
            if (!event.active) sim.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
          })
          .on('drag', (event, d) => {
            d.fx = event.x;
            d.fy = event.y;
          })
          .on('end', (event, d) => {
            if (!event.active) sim.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          }) as any
      );

    // Node circles and avatars
    const isCurrentRoot = (d: GraphNode) => currentRootId !== undefined && d.id === currentRootId;
    const nodeRadius = 35;

    // Add organic node group class
    node.attr('class', 'tree-node-organic');

    // Background circle with enhanced styling
    node
      .append('circle')
      .attr('r', nodeRadius)
      .attr('fill', (d) => {
        if (isCurrentRoot(d)) return isDarkMode ? '#1E3A8A' : '#E3F2FD';
        if (d.isInPath) return isDarkMode ? '#92400E' : '#FFF3E0';
        return isDarkMode ? '#1E293B' : '#FFFFFF';
      })
      .attr('stroke', (d) => {
        if (isCurrentRoot(d)) return isDarkMode ? '#60A5FA' : '#1976D2';
        return getGenderColor(d.data.gender);
      })
      .attr('stroke-width', (d) => (isCurrentRoot(d) ? 5 : d.isInPath ? 4 : 3))
      .attr('class', (d) => isCurrentRoot(d) ? 'root-node-glow' : '')
      .style('cursor', 'pointer')
      .style('filter', (d) => {
        if (isCurrentRoot(d)) return 'drop-shadow(0 0 12px rgba(25, 118, 210, 0.5))';
        return 'drop-shadow(0 4px 8px rgba(0, 0, 0, 0.15))';
      });

    // Add growth ring effect to avatar border
    node.each(function(d) {
      const nodeG = d3.select(this);

      // Growth rings around node (subtle tree ring pattern)
      nodeG.append('circle')
        .attr('r', nodeRadius + 6)
        .attr('fill', 'none')
        .attr('stroke', getGenderColor(d.data.gender))
        .attr('stroke-width', 1)
        .attr('opacity', isDarkMode ? 0.2 : 0.15)
        .style('pointer-events', 'none');

      nodeG.append('circle')
        .attr('r', nodeRadius + 10)
        .attr('fill', 'none')
        .attr('stroke', getGenderColor(d.data.gender))
        .attr('stroke-width', 0.5)
        .attr('opacity', isDarkMode ? 0.15 : 0.1)
        .style('pointer-events', 'none');
    });

    // Avatar images (moved to after growth rings)
    node.each(function(d) {
      const nodeG = d3.select(this);
      const pictureUrl = getMemberPictureUrl(d.data.member_id, d.data.picture);

      if (pictureUrl) {
        // Create circular clip path for avatar
        const clipId = `clip-${d.id}`;
        const defs = nodeG.append('defs');
        defs.append('clipPath')
          .attr('id', clipId)
          .append('circle')
          .attr('r', nodeRadius - 3);

        // Add avatar image
        nodeG.append('image')
          .attr('xlink:href', pictureUrl)
          .attr('x', -(nodeRadius - 3))
          .attr('y', -(nodeRadius - 3))
          .attr('width', (nodeRadius - 3) * 2)
          .attr('height', (nodeRadius - 3) * 2)
          .attr('clip-path', `url(#${clipId})`)
          .style('pointer-events', 'none')
          .style('transition', 'transform 0.3s ease');
      } else {
        // Show initials if no avatar (up to 2 letters)
        const initials = d.data.name ? d.data.name.split(' ').map(n => n[0]).slice(0, 2).join('').toUpperCase() : '?';
        nodeG.append('text')
          .text(initials)
          .attr('text-anchor', 'middle')
          .attr('dy', '0.35em')
          .attr('font-size', '18px')
          .attr('font-weight', 'bold')
          .attr('fill', getGenderColor(d.data.gender))
          .style('pointer-events', 'none')
          .style('text-shadow', isDarkMode
            ? '0 1px 2px rgba(0, 0, 0, 0.8)'
            : '0 1px 2px rgba(255, 255, 255, 0.8)');
      }
    });

    // Node labels with organic styling
    node
      .append('text')
      .text((d) => {
        const name = d.data.name || 'Unknown';
        return name.length > 15 ? name.substring(0, 15) + '...' : name;
      })
      .attr('text-anchor', 'middle')
      .attr('dy', 45)
      .attr('font-size', '13px')
      .attr('font-weight', '600')
      .attr('fill', isDarkMode ? '#F1F5F9' : '#000')
      .attr('class', isDarkMode ? 'tree-label-dark' : 'tree-label')
      .style('pointer-events', 'none');

    // Generation level badge with organic styling - positioned on top-right
    node
      .filter((d) => d.data.generation_level !== undefined)
      .append('circle')
      .attr('cx', 24)
      .attr('cy', -24)
      .attr('r', 11)
      .attr('fill', isDarkMode ? '#60A5FA' : '#2196F3')
      .attr('class', 'generation-badge')
      .style('filter', 'drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3))')
      .style('pointer-events', 'none');

    node
      .filter((d) => d.data.generation_level !== undefined)
      .append('text')
      .text((d) => d.data.generation_level!)
      .attr('x', 24)
      .attr('y', -24)
      .attr('dy', '0.35em')
      .attr('text-anchor', 'middle')
      .attr('font-size', '11px')
      .attr('font-weight', 'bold')
      .attr('fill', 'white')
      .style('pointer-events', 'none')
      .style('text-shadow', '0 1px 2px rgba(0, 0, 0, 0.5)');

    // Married indicator removed - visible from spouse edges

    // Click handlers
    node.on('click', (event, d) => {
      event.stopPropagation();
      onNodeClick(d.data);
    });

    node.on('contextmenu', (event, d) => {
      event.preventDefault();
      event.stopPropagation();
      if (d.id === currentRootId) {
        onSetRoot(-1);
      } else {
        onSetRoot(d.id);
      }
    });

    // No hover animations - completely stable
    node.on('mouseenter', function () {
      d3.select(this).style('opacity', 0.85);
    });

    node.on('mouseleave', function () {
      d3.select(this).style('opacity', 1);
    });

    // Update positions on tick
    sim.on('tick', () => {
      link
        .attr('x1', (d: any) => d.source.x)
        .attr('y1', (d: any) => d.source.y)
        .attr('x2', (d: any) => d.target.x)
        .attr('y2', (d: any) => d.target.y);

      node.attr('transform', (d) => `translate(${d.x},${d.y})`);
    });

    // Zoom functionality
    const zoomBehavior = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 3])
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });

    svg.call(zoomBehavior as any);

    return () => {
      sim.stop();
    };
  }, [data, onNodeClick, onSetRoot, currentRootId, linkDistance, chargeStrength]);

  useEffect(() => {
    if (simulation) {
      if (isPlaying) {
        simulation.restart();
      } else {
        simulation.stop();
      }
    }
  }, [isPlaying, simulation]);

  useEffect(() => {
    if (simulation) {
      simulation
        .force(
          'link',
          d3
            .forceLink<GraphNode, GraphLink>()
            .id((d) => d.id)
            .distance((d) => {
              if (d.type === 'spouse') return linkDistance * 0.5;
              return linkDistance;
            })
        )
        .force('charge', d3.forceManyBody().strength(chargeStrength))
        .alpha(0.3)
        .restart();
    }
  }, [linkDistance, chargeStrength, simulation]);

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

      {/* Controls */}
      <Box
        sx={{
          position: 'absolute',
          top: 16,
          right: 16,
          display: 'flex',
          flexDirection: 'column',
          gap: 1,
        }}
      >
        <Tooltip title={isPlaying ? 'Pause' : 'Play'}>
          <IconButton
            onClick={() => setIsPlaying(!isPlaying)}
            className={isDarkMode ? 'tree-control-button-dark' : 'tree-control-button'}
          >
            {isPlaying ? <Pause /> : <PlayArrow />}
          </IconButton>
        </Tooltip>
      </Box>

      {/* Physics Controls */}
      <Box
        className={isDarkMode ? 'tree-control-button-dark' : 'tree-control-button'}
        sx={{
          position: 'absolute',
          top: 16,
          left: 16,
          p: 2,
          minWidth: 200,
        }}
      >
        <Typography variant="subtitle2" gutterBottom>
          Physics Controls
        </Typography>
        <Box sx={{ mb: 2 }}>
          <Typography variant="caption" color="text.secondary">
            Link Distance: {linkDistance}
          </Typography>
          <Slider
            value={linkDistance}
            onChange={(_, value) => setLinkDistance(value as number)}
            min={50}
            max={300}
            size="small"
          />
        </Box>
        <Box>
          <Typography variant="caption" color="text.secondary">
            Repulsion: {Math.abs(chargeStrength)}
          </Typography>
          <Slider
            value={Math.abs(chargeStrength)}
            onChange={(_, value) => setChargeStrength(-(value as number))}
            min={100}
            max={1000}
            size="small"
          />
        </Box>
      </Box>

      {/* Info Dialog */}
      <Dialog open={infoDialogOpen} onClose={() => setInfoDialogOpen(false)}>
        <DialogTitle>Force Directed Tree View - Instructions</DialogTitle>
        <DialogContent>
          <Typography variant="body2" paragraph>
            <strong>Navigation:</strong>
          </Typography>
          <Typography variant="body2" component="div" sx={{ mb: 2, pl: 2 }}>
            • Click on a node to view detailed member information<br />
            • Right-click on a node to set/unset it as the tree root<br />
            • Drag nodes to manually reposition them<br />
            • Drag background to pan the view<br />
            • Scroll to zoom in/out
          </Typography>

          <Typography variant="body2" paragraph>
            <strong>Physics Controls:</strong>
          </Typography>
          <Typography variant="body2" component="div" sx={{ mb: 2, pl: 2 }}>
            • Use the sliders to adjust link distance and node repulsion<br />
            • Click Play/Pause to start/stop the physics simulation<br />
            • Fine-tune the layout to your preference
          </Typography>

          <Typography variant="body2" paragraph>
            <strong>Visual Indicators:</strong>
          </Typography>
          <Typography variant="body2" component="div" sx={{ pl: 2 }}>
            • <span style={{ color: '#1976D2', fontWeight: 'bold' }}>Blue border</span> = Current root node<br />
            • <span style={{ color: '#EC407A', fontWeight: 'bold' }}>Pink lines</span> = Spouse connections<br />
            • <span style={{ color: '#424242', fontWeight: 'bold' }}>Black lines</span> = Parent-child relationships<br />
            • Siblings are connected naturally through shared parents<br />
            • <span style={{ color: '#FF9800', fontWeight: 'bold' }}>Orange highlights</span> = Relation path (when finding relations)<br />
            • <span style={{ color: '#2196F3', fontWeight: 'bold' }}>Blue badge</span> = Generation level
          </Typography>
        </DialogContent>
      </Dialog>
    </Paper>
  );
};

export default ForceDirectedTree;
