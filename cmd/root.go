/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

type ModuleStruct struct {
	Path      string    `json:"Path"`
	Version   string    `json:"Version"`
	Time      time.Time `json:"Time"`
	Dir       string    `json:"Dir"`
	GoMod     string    `json:"GoMod"`
	GoVersion string    `json:"GoVersion"`
}

type GoListSingleOutputStruct struct {
	Dir          string       `json:"Dir"`
	ImportPath   string       `json:"ImportPath"`
	Name         string       `json:"Name"`
	Doc          string       `json:"Doc"`
	Target       string       `json:"Target"`
	Root         string       `json:"Root"`
	Module       ModuleStruct `json:"Module"`
	Goroot       bool         `json:"Goroot"`
	Standard     bool         `json:"Standard"`
	DepOnly      bool         `json:"DepOnly"`
	GoFiles      []string     `json:"GoFiles"`
	Imports      []string     `json:"Imports"`
	Deps         []string     `json:"Deps"`
	XTestGoFiles []string     `json:"XTestGoFiles"`
	XTestImports []string     `json:"XTestImports"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gomodgraph-sixteen",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		pathToModule := make(map[string]string)
		isStandardPath := make(map[string]bool)
		pkgs := make(map[string][]string)

		goList := exec.Command("go", "list", "-json", "-deps")
		goListOutput, err := goList.Output()
		if err != nil {
			log.Fatal(err)
		}
		goListOutputString := string(goListOutput)

		scanner := bufio.NewScanner(strings.NewReader(goListOutputString))
		start := false
		jsonString := ""
		for scanner.Scan() {
			line := scanner.Text()
			words := strings.Fields(line)
			if words[0] == "{" {
				start = true
				jsonString = "{"
			} else if words[0] == "}" {
				start = false
				jsonString += "}"

				var goListSingleOutput GoListSingleOutputStruct
				json.Unmarshal([]byte(jsonString), &goListSingleOutput)

				pkgs[goListSingleOutput.ImportPath] = goListSingleOutput.Deps
				pathToModule[goListSingleOutput.ImportPath] = goListSingleOutput.Module.Path
				isStandardPath[goListSingleOutput.ImportPath] = goListSingleOutput.Standard
			} else {
				if start {
					jsonString += line
				}
			}
		}

		var goModGraphOutput [][]string
		for pkg, deps := range pkgs {
			if !isStandardPath[pkg] {
				for _, dep := range deps {

					// fmt.Println("pkg", pkg)
					// fmt.Println("deps", deps)
					// fmt.Println("pathToModule", pathToModule[pkg])
					// fmt.Println("isStandardPath", isStandardPath[pkg])
					if !isStandardPath[dep] {
						var goModGraphOutputLine []string
						goModGraphOutputLine = append(goModGraphOutputLine, pathToModule[pkg])
						goModGraphOutputLine = append(goModGraphOutputLine, pathToModule[dep])
						if !contains(goModGraphOutput, goModGraphOutputLine) && goModGraphOutputLine[0] != goModGraphOutputLine[1] {
							goModGraphOutput = append(goModGraphOutput, goModGraphOutputLine)
						}
					}

				}
			}
		}

		finalOutput := ""
		for _, line := range goModGraphOutput {
			finalOutput += (strings.Join(line, " -> ") + "\n")
		}
		fmt.Println(finalOutput)
	},
}

func contains(a [][]string, b []string) bool {
	for _, v := range a {
		if equal(v, b) {
			return true
		}
	}
	return false
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gomodgraph-sixteen.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gomodgraph-sixteen" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gomodgraph-sixteen")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
