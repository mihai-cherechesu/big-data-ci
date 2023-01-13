import React, {Component} from 'react';
import * as d3 from "d3";
import * as d3d from "d3-dag";

const nodeRadius = 30;
const edgeRadius = 5;

class Dag extends Component {
  componentDidMount() {
    this.drawDag();
  }

  horizontalizeLayout(baseLayout, dag) {
    const { width, height } = baseLayout(dag);
    for (const node of dag) {
      [node.x, node.y] = [node.y, node.x];
    }
    for (const { points } of dag.ilinks()) {
      for (const point of points) {
        [point.x, point.y] = [point.y, point.x];
      }
    }
    return { width: height, height: width };
  }

  drawDag() {
    d3.selectAll("svg").remove()

    const dagData = this.props.data;
    const dag = d3d.dagConnect()(dagData);
    const baseLayout = d3d
      .zherebko()
      .nodeSize([
        nodeRadius * 2,
        (nodeRadius + edgeRadius) * 2,
        edgeRadius * 2
      ])
    
    const layout = this.horizontalizeLayout(baseLayout, dag);
    const width = layout.width;
    const height = layout.height;
    const svgSelection = d3.select("body").append("svg")
    .attr("width", width)
    .attr("height", height)
    .attr(
      "viewBox",
      `${-nodeRadius} ${-nodeRadius} ${width + 2 * nodeRadius} ${
        height + 2 * nodeRadius
      }`
    );

    const defs = svgSelection.append("defs");
    const steps = dag.size();
    const interp = d3.interpolateRainbow;
    const colorMap = new Map();

    const descendants = dag.descendants()
    for (let i = 0; i < descendants.length; i++) {
      colorMap.set(descendants[i].data.id, interp(i / steps));
    }

    const line = d3
      .line()
      .curve(d3.curveMonotoneX)
      .x((d) => d.x)
      .y((d) => d.y);
    
    svgSelection
      .append("g")
      .selectAll("path")
      .data(dag.links())
      .enter()
      .append("path")
      .attr("d", ({ points }) => line(points))
      .attr("fill", "none")
      .attr("stroke-width", 3)
      .attr("stroke", ({ source, target }) => {
        const gradId = encodeURIComponent(
          `${source.data.id}-${target.data.id}`
        );
        const grad = defs
          .append("linearGradient")
          .attr("id", gradId)
          .attr("gradientUnits", "userSpaceOnUse")
          .attr("x1", source.x)
          .attr("x2", target.x)
          .attr("y1", source.y)
          .attr("y2", target.y);
        grad
          .append("stop")
          .attr("offset", "0%")
          // .attr("stop-color", colorMap.get(source.data.id));
        grad
          .append("stop")
          .attr("offset", "100%")
          // .attr("stop-color", colorMap.get(target.data.id));
        return `url(#${gradId})`;
      });
      
      const nodes = svgSelection
        .append("g")
        .selectAll("g")
        .data(dag.descendants())
        .enter()
        .append("g")
        .attr("transform", ({ x, y }) => `translate(${x}, ${y})`);
  
      nodes
        .append("circle")
        .attr("r", nodeRadius);
        // .attr("fill", (n) => colorMap.get(n.data.id));
  
      nodes
        .append("text")
        .text((d) => d.data.id)
        .attr("font-weight", "bold")
        .attr("font-family", "sans-serif")
        .attr("text-anchor", "middle")
        .attr("alignment-baseline", "middle")
        .attr("fill", "white");
  }

  render() {  
      return (
        <div></div>
      )
  }
}



export default Dag;