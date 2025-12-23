import React, { useEffect, useRef } from 'react';
import { Box, Paper, IconButton, Tooltip } from '@mui/material';
import { ZoomIn, ZoomOut, ZoomOutMap } from '@mui/icons-material';
import * as d3 from 'd3';
import { TreeNode, Member } from '../types';
import { getGenderColor, getMemberPictureUrl } from '../utils/helpers';

interface TreeVisualizationProps {
  data: TreeNode;
  onNodeClick: (member: Member) => void;
  onSetRoot: (memberId: number) => void;
  currentRootId?: number;
}

interface D3Node extends d3.HierarchyPointNode<TreeNode> {
  _children?: D3Node[];
}

const TreeVisualization: React.FC<TreeVisualizationProps> = ({ data, onNodeClick, onSetRoot, currentRootId }) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!svgRef.current || !containerRef.current || !data) return;

    // Clear previous content
    d3.select(svgRef.current).selectAll('*').remove();

    const container = containerRef.current;
    const width = container.clientWidth;
    const height = container.clientHeight;

    // Node dimensions
    const nodeWidth = 180;
    const nodeHeight = 100;
    const horizontalSpacing = 40;
    const verticalSpacing = 150;

    // Create SVG
    const svg = d3
      .select(svgRef.current)
      .attr('width', width)
      .attr('height', height);

    const g = svg.append('g').attr('transform', `translate(${width / 2},50)`);

    // Create tree layout
    const treeLayout = d3
      .tree<TreeNode>()
      .nodeSize([nodeWidth + horizontalSpacing, nodeHeight + verticalSpacing])
      .separation((a, b) => {
        // Increase separation for nodes with spouses
        const aHasSpouses = a.data.spouse_nodes && a.data.spouse_nodes.length > 0;
        const bHasSpouses = b.data.spouse_nodes && b.data.spouse_nodes.length > 0;
        return aHasSpouses || bHasSpouses ? 2 : 1;
      });

    // Create hierarchy
    const root = d3.hierarchy(data);
    const treeData = treeLayout(root);

    // Get all nodes
    const nodes = treeData.descendants();
    const links = treeData.links();

    // Process spouse and sibling nodes
    const allNodes: Array<{ node: D3Node; isSpouse: boolean; isSibling: boolean; relatedTo?: D3Node }> = [];
    const allLinks: Array<{ source: D3Node; target: D3Node; type: 'parent' | 'spouse' | 'sibling' }> = [];

    nodes.forEach((node) => {
      allNodes.push({ node: node as D3Node, isSpouse: false, isSibling: false });

      let offset = 0;

      // Add spouse nodes
      if (node.data.spouse_nodes && node.data.spouse_nodes.length > 0) {
        node.data.spouse_nodes.forEach((spouseData) => {
          offset++;
          const spouseNode: any = {
            data: spouseData,
            x: node.x + offset * (nodeWidth + 20),
            y: node.y,
            parent: node.parent,
          };
          allNodes.push({ node: spouseNode, isSpouse: true, isSibling: false, relatedTo: node as D3Node });

          // Add spouse link (pink line)
          allLinks.push({ source: node as D3Node, target: spouseNode, type: 'spouse' });
        });
      }

      // Add sibling nodes
      if (node.data.sibling_nodes && node.data.sibling_nodes.length > 0) {
        node.data.sibling_nodes.forEach((siblingData) => {
          offset++;
          const siblingNode: any = {
            data: siblingData,
            x: node.x + offset * (nodeWidth + 20),
            y: node.y,
            parent: node.parent,
          };
          allNodes.push({ node: siblingNode, isSpouse: false, isSibling: true, relatedTo: node as D3Node });

          // Add sibling link (gray dashed line)
          allLinks.push({ source: node as D3Node, target: siblingNode, type: 'sibling' });

          // Add sibling's spouses
          if (siblingData.spouse_nodes && siblingData.spouse_nodes.length > 0) {
            siblingData.spouse_nodes.forEach((siblingSpouseData) => {
              offset++;
              const siblingSpouseNode: any = {
                data: siblingSpouseData,
                x: node.x + offset * (nodeWidth + 20),
                y: node.y,
                parent: node.parent,
              };
              allNodes.push({ node: siblingSpouseNode, isSpouse: true, isSibling: false, relatedTo: siblingNode });

              // Add spouse link for sibling's spouse
              allLinks.push({ source: siblingNode, target: siblingSpouseNode, type: 'spouse' });
            });
          }
        });
      }
    });

    // Add parent-child links (black lines)
    links.forEach((link) => {
      allLinks.push({ source: link.source as D3Node, target: link.target as D3Node, type: 'parent' });
    });

    // Draw links
    const linkGroup = g.append('g').attr('class', 'links');

    allLinks.forEach((link) => {
      const path = linkGroup
        .append('path')
        .attr('class', `link ${link.type}`)
        .attr('d', () => {
          if (link.type === 'spouse' || link.type === 'sibling') {
            // Horizontal line for spouses and siblings
            return `M ${link.source.x},${link.source.y} L ${link.target.x},${link.target.y}`;
          } else {
            // Vertical line from parent to child
            const sourceX = link.source.x;
            const sourceY = link.source.y + nodeHeight / 2;
            const targetX = link.target.x;
            const targetY = link.target.y - nodeHeight / 2;
            const midY = (sourceY + targetY) / 2;

            return `M ${sourceX},${sourceY}
                    L ${sourceX},${midY}
                    L ${targetX},${midY}
                    L ${targetX},${targetY}`;
          }
        })
        .attr('fill', 'none')
        .attr('stroke',
          link.type === 'spouse' ? '#EC407A' :
          link.type === 'sibling' ? '#9E9E9E' :
          '#424242'
        )
        .attr('stroke-width', link.type === 'spouse' ? 3 : 2)
        .attr('stroke-dasharray', link.type === 'sibling' ? '5,5' : 'none')
        .attr('opacity', 0.7);

      // Highlight path links
      if (link.source.data.is_in_path && link.target.data.is_in_path && link.type === 'parent') {
        path.attr('stroke', '#FF9800').attr('stroke-width', 4).attr('opacity', 1);
      }
    });

    // Draw nodes
    const nodeGroup = g.append('g').attr('class', 'nodes');

    allNodes.forEach(({ node }) => {
      const nodeG = nodeGroup
        .append('g')
        .attr('class', 'node')
        .attr('transform', `translate(${node.x - nodeWidth / 2},${node.y - nodeHeight / 2})`)
        .style('cursor', 'pointer');

      // Check if this is the current root
      const isCurrentRoot = currentRootId !== undefined && node.data.member.member_id === currentRootId;

      // Node background with gender color border
      const rect = nodeG
        .append('rect')
        .attr('width', nodeWidth)
        .attr('height', nodeHeight)
        .attr('rx', 8)
        .attr('fill', isCurrentRoot ? '#E3F2FD' : '#ffffff')
        .attr('stroke', isCurrentRoot ? '#1976D2' : getGenderColor(node.data.member.gender))
        .attr('stroke-width', isCurrentRoot ? 5 : node.data.is_in_path ? 4 : 3);

      // Highlight path nodes (but root takes precedence)
      if (!isCurrentRoot && node.data.is_in_path) {
        rect.attr('fill', '#FFF3E0');
      }

      // Add picture or avatar
      const pictureUrl = getMemberPictureUrl(
        node.data.member.member_id,
        node.data.member.picture
      );

      if (pictureUrl) {
        nodeG
          .append('image')
          .attr('x', 10)
          .attr('y', 10)
          .attr('width', 50)
          .attr('height', 50)
          .attr('href', pictureUrl)
          .attr('clip-path', 'circle(25px at 25px 25px)');
      } else {
        nodeG
          .append('circle')
          .attr('cx', 35)
          .attr('cy', 35)
          .attr('r', 25)
          .attr('fill', getGenderColor(node.data.member.gender))
          .attr('opacity', 0.3);

        nodeG
          .append('text')
          .attr('x', 35)
          .attr('y', 40)
          .attr('text-anchor', 'middle')
          .attr('fill', '#000')
          .attr('font-size', '20px')
          .attr('font-weight', 'bold')
          .text(node.data.member.english_name.charAt(0));
      }

      // Add names
      nodeG
        .append('text')
        .attr('x', 75)
        .attr('y', 25)
        .attr('font-size', '12px')
        .attr('font-weight', 'bold')
        .attr('fill', '#000')
        .text(truncateText(node.data.member.arabic_name, 15));

      nodeG
        .append('text')
        .attr('x', 75)
        .attr('y', 42)
        .attr('font-size', '11px')
        .attr('fill', '#666')
        .text(truncateText(node.data.member.english_name, 15));

      // Add generation level or age
      if (node.data.member.generation_level !== undefined) {
        nodeG
          .append('text')
          .attr('x', 75)
          .attr('y', 57)
          .attr('font-size', '10px')
          .attr('fill', '#999')
          .text(`Gen: ${node.data.member.generation_level}`);
      } else if (node.data.member.age) {
        nodeG
          .append('text')
          .attr('x', 75)
          .attr('y', 57)
          .attr('font-size', '10px')
          .attr('fill', '#999')
          .text(`Age: ${node.data.member.age}`);
      }

      // Married indicator
      if (node.data.member.is_married) {
        nodeG
          .append('circle')
          .attr('cx', nodeWidth - 15)
          .attr('cy', 15)
          .attr('r', 5)
          .attr('fill', '#EC407A');
      }

      // Click handlers
      nodeG.on('click', (event) => {
        event.stopPropagation();
        onNodeClick(node.data.member);
      });

      nodeG.on('contextmenu', (event) => {
        event.preventDefault();
        event.stopPropagation();
        // If this is already the root, unselect it (pass undefined)
        if (node.depth === 0) {
          onSetRoot(-1); // Special value to indicate root reset
        } else {
          onSetRoot(node.data.member.member_id);
        }
      });

      // Hover effect
      nodeG.on('mouseenter', function () {
        d3.select(this).select('rect').attr('filter', 'drop-shadow(0 4px 8px rgba(0,0,0,0.2))');
      });

      nodeG.on('mouseleave', function () {
        d3.select(this).select('rect').attr('filter', 'none');
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

      const scale = Math.min(width / fullWidth, height / fullHeight) * 0.8;
      const translate = [width / 2 - scale * midX, height / 2 - scale * midY];

      svg.call(
        zoomBehavior.transform as any,
        d3.zoomIdentity.translate(translate[0], translate[1]).scale(scale)
      );
    }
  }, [data, onNodeClick, onSetRoot]);

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

      {/* Zoom controls */}
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

      {/* Instructions */}
      <Box
        sx={{
          position: 'absolute',
          bottom: 16,
          left: 16,
          bgcolor: 'white',
          p: 2,
          borderRadius: 1,
          boxShadow: 2,
        }}
      >
        <Box sx={{ fontSize: '12px', color: '#666' }}>
          <div>• Click node to view details</div>
          <div>• Right-click node to set/unset as root</div>
          <div>• Blue border = current root</div>
          <div>• Drag to pan • Scroll to zoom</div>
        </Box>
      </Box>
    </Paper>
  );
};

// Helper function to truncate text
function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

export default TreeVisualization;
