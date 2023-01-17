import React from 'react';
import './App.css';
import axios from "axios"
import * as d3 from "d3";
import * as d3d from "d3-dag";

export const stageStatuses = {
  RUNNING: 'RUNNING',
  FAILED: 'FAILED',
  SUCCESS: 'SUCCESS'
};

export const statusColors = {
  'RUNNING': 'blue',
  'FAILED': 'red',
  'SUCCESS': 'green'
};

const nodeRadius = 50;
const edgeRadius = 5;

class App extends React.Component {

  // Default state
  state = {
    pipelines: [
      {
        "Id": "pipeline-id",
        "UserId": "user-id",
        "Dependencies": [
          [
            "build",
            "test"
          ]
        ]
      }
    ],
    stages: {
      "pipeline-id": [
        {
          "name": "build",
          "status": "SUCCESS"
        },
        {
          "name": "test",
          "status": "SUCCESS"
        }
      ]
    }
  }

  getPipelineState(pipelineId, stagesMap) {
    let stages = []
    let status = stageStatuses.SUCCESS;

    if (!stagesMap.hasOwnProperty(pipelineId)) {
      stages = stagesMap["pipeline-id"]
    } else {
      stages = stagesMap[pipelineId]
    }

    stages.forEach(s => {
      if (s.status == stageStatuses.FAILED) {
        status = stageStatuses.FAILED

      } else if (s.status == stageStatuses.RUNNING) {
        status = stageStatuses.RUNNING
      }
    });
  
    return status
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

  drawDag(pipelineId, dependencies, stages) {
    const svgId = `#svg-${pipelineId}`
    d3.selectAll(svgId).remove();

    const dagData = dependencies;
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
    const svgSelection = d3.select("#td-" + pipelineId).append("svg")
    .attr("id", "svg-" + pipelineId)
    .attr("width", width)
    .attr("height", height)
    .attr(
      "viewBox",
      `${-nodeRadius} ${-nodeRadius} ${width + 2 * nodeRadius} ${
        height + 2 * nodeRadius
      }`
    );

    const defs = svgSelection.append("defs");
    const colorMap = new Map();
    const descendants = dag.descendants();

    for (let i = 0; i < descendants.length; i++) {

      if (stages.hasOwnProperty(pipelineId)) {
        let pipelineStages = stages[pipelineId]
        pipelineStages.forEach(s => {
          if (s.name != descendants[i].data.id) {
            return;
          }
          colorMap.set(descendants[i].data.id, statusColors[s.status])
        });
      }
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
          .attr("stop-color", colorMap.get(source.data.id));
        grad
          .append("stop")
          .attr("offset", "100%")
          .attr("stop-color", colorMap.get(target.data.id));
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
        .attr("r", nodeRadius)
        .attr("fill", (n) => colorMap.get(n.data.id));
  
      nodes
        .append("text")
        .text((d) => d.data.id)
        .attr("font-weight", "bold")
        .attr("font-family", "sans-serif")
        .attr("text-anchor", "middle")
        .attr("alignment-baseline", "middle")
        .attr("fill", "white");
  }

  componentDidMount() {
    axios.get("http://localhost:8081/pipelines/")
    .then(pipelinesData => {
      this.setState({
        pipelines: pipelinesData.data
      })

      let pipelineIds = pipelinesData.data.map(v => v.Id)
      console.log("Received pipelineIds: " + pipelineIds)

      axios({
        method: "post",
        url: "http://localhost:8081/stages",
        data: JSON.stringify(pipelineIds)
      })
      .then(stagesData => {
        console.log("Received stages: " + Object.keys(stagesData.data))
        

        this.setState(prevState => {
          return {
            pipelines: prevState.pipelines,
            stages: stagesData.data
          }
        })
      })
    })

    this.setState(prevState => {
      return { prevState };
    })
  }

  render() {
    return (
      <div>
        <table class="styled-table table table-bordered text-center">
          <thead>
            <tr>
              <th><h2>Status</h2></th>
              <th><h2>Pipeline ID</h2></th>
              <th><h2>Triggerer</h2></th>
              <th><h2>Stages</h2></th>
            </tr>
          </thead>
          <tbody>
            {this.state.pipelines.map(pipeline => (
              <tr key={pipeline.Id}>
                <td class="text-center">
                  <div 
                    className="status-indicator"
                    style={{ color: statusColors[this.getPipelineState(pipeline.Id, this.state.stages)]}}
                  >
                    {this.getPipelineState(pipeline.Id, this.state.stages)}
                  </div>
                </td>
                <td class="text-center">{pipeline.Id}</td>
                <td class="text-center">{pipeline.UserId}</td>
                <td id={"td-" + pipeline.Id} class="text-center">{this.drawDag(pipeline.Id, pipeline.Dependencies, this.state.stages)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )
  }
}

export default App;
