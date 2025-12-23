import React, { useEffect, useRef, useState } from 'react';
import { Box, Paper, IconButton, Tooltip, Slider, Typography, Dialog, DialogTitle, DialogContent } from '@mui/material';
import { ZoomIn, ZoomOut, ZoomOutMap, PlayArrow, Pause, Info } from '@mui/icons-material';
import * as d3 from 'd3';
import { TreeNode, Member } from '../types';
import { getGenderColor, getMemberPictureUrl } from '../utils/helpers';

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
  type: 'parent' | 'spouse' | 'sibling';
}

const ForceDirectedTree: React.FC<ForceDirectedTreeProps> = ({
  data,
  onNodeClick,
  onSetRoot,
  currentRootId,
}) => {
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

    // Convert tree to graph structure
    const nodes: GraphNode[] = [];
    const links: GraphLink[] = [];
    const nodeMap = new Map<number, GraphNode>();

    const processNode = (treeNode: TreeNode) => {
      const node: GraphNode = {
        id: treeNode.member.member_id,
        data: treeNode.member,
        isInPath: treeNode.is_in_path,
      };
      nodes.push(node);
      nodeMap.set(node.id, node);

      // Add spouse nodes
      if (treeNode.spouse_nodes) {
        treeNode.spouse_nodes.forEach((spouse) => {
          const spouseNode: GraphNode = {
            id: spouse.member.member_id,
            data: spouse.member,
            isSpouse: true,
            isInPath: spouse.is_in_path,
          };
          nodes.push(spouseNode);
          nodeMap.set(spouseNode.id, spouseNode);
          links.push({ source: node.id, target: spouseNode.id, type: 'spouse' });
        });
      }

      // Add sibling nodes
      if (treeNode.sibling_nodes) {
        treeNode.sibling_nodes.forEach((sibling) => {
          const siblingNode: GraphNode = {
            id: sibling.member.member_id,
            data: sibling.member,
            isSibling: true,
            isInPath: sibling.is_in_path,
          };
          nodes.push(siblingNode);
          nodeMap.set(siblingNode.id, siblingNode);
          links.push({ source: node.id, target: siblingNode.id, type: 'sibling' });

          // Add sibling's spouses
          if (sibling.spouse_nodes) {
            sibling.spouse_nodes.forEach((siblingSpouse) => {
              const siblingSpouseNode: GraphNode = {
                id: siblingSpouse.member.member_id,
                data: siblingSpouse.member,
                isSpouse: true,
                isInPath: siblingSpouse.is_in_path,
              };
              nodes.push(siblingSpouseNode);
              nodeMap.set(siblingSpouseNode.id, siblingSpouseNode);
              links.push({ source: siblingNode.id, target: siblingSpouseNode.id, type: 'spouse' });
            });
          }
        });
      }

      // Add children
      if (treeNode.children) {
        treeNode.children.forEach((child) => {
          processNode(child);
          links.push({ source: node.id, target: child.member.member_id, type: 'parent' });
        });
      }
    };

    processNode(data);

    // Create SVG
    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height);

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
            if (d.type === 'sibling') return linkDistance * 0.7;
            return linkDistance;
          })
          .strength(1)
      )
      .force('charge', d3.forceManyBody().strength(chargeStrength))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(50));

    setSimulation(sim);

    // Draw links
    const link = g
      .append('g')
      .selectAll('line')
      .data(links)
      .join('line')
      .attr('stroke', (d) => {
        if (d.type === 'spouse') return '#EC407A';
        if (d.type === 'sibling') return '#9E9E9E';
        return '#424242';
      })
      .attr('stroke-width', (d) => (d.type === 'spouse' ? 3 : 2))
      .attr('stroke-dasharray', (d) => (d.type === 'sibling' ? '5,5' : 'none'))
      .attr('opacity', 0.6);

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
          })
      );

    // Node circles
    const isCurrentRoot = (d: GraphNode) => currentRootId !== undefined && d.id === currentRootId;

    node
      .append('circle')
      .attr('r', 30)
      .attr('fill', (d) => (isCurrentRoot(d) ? '#E3F2FD' : '#ffffff'))
      .attr('stroke', (d) => (isCurrentRoot(d) ? '#1976D2' : getGenderColor(d.data.gender)))
      .attr('stroke-width', (d) => (isCurrentRoot(d) ? 5 : d.isInPath ? 4 : 3))
      .style('cursor', 'pointer');

    // Path highlighting
    node.filter((d) => d.isInPath).select('circle').attr('fill', '#FFF3E0');

    // Show first letter (no avatars/images)
    node
      .append('text')
      .text((d) => d.data.english_name.charAt(0).toUpperCase())
      .attr('text-anchor', 'middle')
      .attr('dy', '0.35em')
      .attr('font-size', '18px')
      .attr('font-weight', 'bold')
      .attr('fill', (d) => getGenderColor(d.data.gender))
      .style('pointer-events', 'none');

    // Node labels
    node
      .append('text')
      .text((d) => d.data.english_name)
      .attr('text-anchor', 'middle')
      .attr('dy', 45)
      .attr('font-size', '12px')
      .attr('font-weight', 'bold')
      .attr('fill', '#333')
      .style('pointer-events', 'none');

    // Generation level badge
    node
      .filter((d) => d.data.generation_level !== undefined)
      .append('circle')
      .attr('cx', 20)
      .attr('cy', -20)
      .attr('r', 12)
      .attr('fill', '#2196F3')
      .style('pointer-events', 'none');

    node
      .filter((d) => d.data.generation_level !== undefined)
      .append('text')
      .text((d) => d.data.generation_level!)
      .attr('x', 20)
      .attr('y', -16)
      .attr('text-anchor', 'middle')
      .attr('font-size', '10px')
      .attr('font-weight', 'bold')
      .attr('fill', 'white')
      .style('pointer-events', 'none');

    // Married indicator
    node
      .filter((d) => d.data.is_married)
      .append('circle')
      .attr('cx', -20)
      .attr('cy', -20)
      .attr('r', 6)
      .attr('fill', '#EC407A')
      .style('pointer-events', 'none');

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

    // Hover effects
    node.on('mouseenter', function () {
      d3.select(this).select('circle').attr('filter', 'drop-shadow(0 4px 8px rgba(0,0,0,0.3))');
    });

    node.on('mouseleave', function () {
      d3.select(this).select('circle').attr('filter', 'none');
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
              if (d.type === 'sibling') return linkDistance * 0.7;
              return linkDistance;
            })
        )
        .force('charge', d3.forceManyBody().strength(chargeStrength))
        .alpha(0.3)
        .restart();
    }
  }, [linkDistance, chargeStrength, simulation]);

  const handleZoomIn = () => {
    if (svgRef.current) {
      const svg = d3.select(svgRef.current);
      svg.transition().call(d3.zoom<SVGSVGElement, unknown>().scaleBy as any, 1.3);
    }
  };

  const handleZoomOut = () => {
    if (svgRef.current) {
      const svg = d3.select(svgRef.current);
      svg.transition().call(d3.zoom<SVGSVGElement, unknown>().scaleBy as any, 0.7);
    }
  };

  const handleResetZoom = () => {
    if (svgRef.current) {
      const svg = d3.select(svgRef.current);
      svg.transition().call(d3.zoom<SVGSVGElement, unknown>().transform as any, d3.zoomIdentity);
    }
  };

  return (
    <Paper
      ref={containerRef}
      sx={{
        position: 'relative',
        width: '100%',
        height: '800px',
        overflow: 'hidden',
        backgroundColor: '#fafafa',
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
        <Tooltip title="Info">
          <IconButton onClick={() => setInfoDialogOpen(true)} sx={{ bgcolor: 'white' }}>
            <Info />
          </IconButton>
        </Tooltip>
        <Tooltip title={isPlaying ? 'Pause' : 'Play'}>
          <IconButton onClick={() => setIsPlaying(!isPlaying)} sx={{ bgcolor: 'white' }}>
            {isPlaying ? <Pause /> : <PlayArrow />}
          </IconButton>
        </Tooltip>
        <Tooltip title="Zoom In">
          <IconButton onClick={handleZoomIn} sx={{ bgcolor: 'white' }}>
            <ZoomIn />
          </IconButton>
        </Tooltip>
        <Tooltip title="Zoom Out">
          <IconButton onClick={handleZoomOut} sx={{ bgcolor: 'white' }}>
            <ZoomOut />
          </IconButton>
        </Tooltip>
        <Tooltip title="Reset Zoom">
          <IconButton onClick={handleResetZoom} sx={{ bgcolor: 'white' }}>
            <ZoomOutMap />
          </IconButton>
        </Tooltip>
      </Box>

      {/* Physics Controls */}
      <Box
        sx={{
          position: 'absolute',
          top: 16,
          left: 16,
          bgcolor: 'white',
          p: 2,
          borderRadius: 1,
          boxShadow: 2,
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
            • <span style={{ color: '#9E9E9E', fontWeight: 'bold' }}>Gray dashed lines</span> = Sibling connections<br />
            • <span style={{ color: '#FF9800', fontWeight: 'bold' }}>Orange highlights</span> = Relation path (when finding relations)<br />
            • <span style={{ color: '#EC407A', fontWeight: 'bold' }}>Pink dot</span> = Married member<br />
            • <span style={{ color: '#2196F3', fontWeight: 'bold' }}>Blue badge</span> = Generation level
          </Typography>
        </DialogContent>
      </Dialog>
    </Paper>
  );
};

export default ForceDirectedTree;
