/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a pipeline based on a .yaml file given as parameter.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")
		f, _ := cmd.Flags().GetString("file")
		runPipeline(f)
	},
}

func runPipeline(file string) {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var pipeline map[string]interface{}
	err = yaml.Unmarshal([]byte(data), &pipeline)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	stages, ok := pipeline["stages"].(map[string]interface{})
	if !ok {
		log.Fatal("no stages found in yaml file\n")
	}

	for _, stage := range stages {
		script, ok := stage.(map[string]interface{})["script"].(string)
		if !ok {
			log.Fatal("no script found for stage in yaml file\n")
		}

		script = strings.TrimSuffix(script, "\n")
		cmds := strings.Split(script, "\n")
		joined := strings.Join(cmds, " && ")
		final := []string{"/bin/sh", "-c", joined}

		stage.(map[string]interface{})["script"] = final
	}

	body, err := json.Marshal(pipeline)
	if err != nil {
		panic(err)
	}

	bufferBody := bytes.NewBuffer(body)
	resp, err := http.Post("http://localhost:8081/execute", "application/json", bufferBody)

	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("resp with body %s with status %d\n", respBody, resp.StatusCode)
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringP("file", "f", "pipeline.yaml", "pipeline file (default is pipeline.yaml)")
}
