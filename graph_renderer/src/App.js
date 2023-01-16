import { useEffect, useState } from 'react';
import './App.css';
import Dag from './components/dag';

export const stageStatuses = {
  RUNNING: 'Running',
  FAILED: 'Failed',
  SUCCESSFUL: 'Successful'
};

const statusColors = {
  [stageStatuses.RUNNING]: 'blue',
  [stageStatuses.FAILED]: 'red',
  [stageStatuses.SUCCESSFUL]: 'green'
};

function App() {

  const [pipelines, setPipelines] = useState();
  const [stages, setStages] = useState();

  // const dagData = [
  //   ["1", "2"],
  //   ["1", "5"],
  //   ["1", "7"],
  //   ["2", "3"],
  //   ["2", "4"],
  //   ["2", "5"],
  //   ["2", "7"],
  //   ["2", "8"],
  //   ["3", "6"],
  //   ["3", "8"],
  //   ["4", "7"],
  //   ["5", "7"],
  //   ["5", "8"],
  //   ["5", "9"],
  //   ["6", "8"],
  //   ["7", "8"],
  //   ["9", "10"],
  //   ["9", "11"]
  // ]

  useEffect(() => {
    // fetch data
    const dataFetch = async () => {
      const pipelinesData = await (
        await fetch(
          "http://localhost:8081/pipelines/"
        )
      ).json();

      // set state when the data received
      setPipelines(pipelinesData);
      console.log("Received pipelines: " + pipelinesData)

      const requestHeaders = new Headers();
      requestHeaders.append('Content-Type', 'application/json');
  
      let pipelineIds = pipelinesData.map(v => v.Id)
      console.log("Received pipelineIds: " + pipelineIds)
  
      const requestOptions = {
        method: 'POST',
        headers: requestHeaders,
        body: JSON.stringify(pipelineIds)
      };

      const stagesData = await (
        await fetch(
          "http://localhost:8081/stages",
          requestOptions
        )
      ).json();
      
      // set state when the data received
      setStages(stagesData);
      console.log("Received stages: " + stagesData)
    };

    dataFetch();
  }, []);  

  // TODOs
  // Fetch data correctly, check it
  // Create the table component
  // Populate the table with fetched data
  // Put the svg under the stages (scale the svg to be small enough or idk)
  // Listen on socket server for events ("stage changed state")
  // On event change, use the update function to update the svg

  return (
    <div>Hello</div>
    // <table>
    //   <thead>
    //     <tr>
    //       <th>Status</th>
    //       <th>Pipeline ID</th>
    //       <th>Triggerer</th>
    //       <th>Stages</th>
    //     </tr>
    //   </thead>
    //   <tbody>
    //     {pipelines.map(pipeline => (
    //       <tr key={pipeline.id}>
    //         {/* <td>
    //           <div 
    //             className="status-indicator"
    //             style={{ backgroundColor: statusColors[getPipelineState(stages[pipeline.id])] }}
    //           >
    //             {getPipelineState(stages[pipeline.id])}
    //           </div>
    //         </td>
    //         <td>{pipeline.id}</td>
    //         <td>Triggerer</td>
    //         <td><Dag pipelineData={pipeline} stagesData={stages} /></td> */}
    //         <td>Status</td>
    //         <td>ID</td>
    //         <td>Triggerer</td>
    //         <td>Stages</td>
    //       </tr>
    //     ))}
    //   </tbody>
    // </table>
  );
}

function getPipelineState(stages) {
  stages.forEach(s => {
    if (s.State == stageStatuses.FAILED) {
      return stageStatuses.FAILED;
    }

    if (s.State == stageStatuses.RUNNING) {
      return stageStatuses.RUNNING;
    }
  });

  return stageStatuses.SUCCESSFUL;
}

export default App;
