import { useState } from 'react';
import './App.css';
import Dag from './components/dag';

function App() {
  const dagData = [
    ["1", "2"],
    ["1", "5"],
    ["1", "7"],
    ["2", "3"],
    ["2", "4"],
    ["2", "5"],
    ["2", "7"],
    ["2", "8"],
    ["3", "6"],
    ["3", "8"],
    ["4", "7"],
    ["5", "7"],
    ["5", "8"],
    ["5", "9"],
    ["6", "8"],
    ["7", "8"],
    ["9", "10"],
    ["9", "11"]
  ]

  const [pipelines, setPipelines] = useState([]);
  const [stages, setStages] = useState([]);

  fetch("http://localhost:8081/pipelines/")
    .then(r => r.json)
    .then(data => setPipelines(data));

  
  fetch("http://localhost:8081/stages")
    .then(r => r.json)
    .then(data => setStages(data));

  // TODOs
  // Fetch data correctly, check it
  // Create the table component
  // Populate the table with fetched data
  // Put the svg under the stages (scale the svg to be small enough or idk)
  // Listen on socket server for events ("stage changed state")
  // On event change, use the update function to update the svg
  

  return (
    <Dag data={dagData}/>
  );
}

export default App;
