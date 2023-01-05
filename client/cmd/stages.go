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
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

// stagesCmd represents the stages command
var stagesCmd = &cobra.Command{
	Use:   "stages",
	Short: "Lists the stages for a specific pipeline.",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("stages called")
		id, _ := cmd.Flags().GetString("id")

		fmt.Printf("id: %s\n", id)

		resp, err := http.Get("http://localhost:8081/pipelines/" + id)
		if err != nil {
			log.Fatalln(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		sb := string(body)
		log.Printf("Response body %s\n", sb)
	},
}

func init() {
	lsCmd.AddCommand(stagesCmd)
	stagesCmd.PersistentFlags().StringP("id", "i", "", "The id of the pipeline to be queried.")
}
