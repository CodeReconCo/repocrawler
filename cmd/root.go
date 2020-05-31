/*
Copyright © 2020 Mike de Libero

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"time"
)

var cfgFile string
var Organization string
var OutputFile string
var ScmURL string
var TokenName string
var IsGitlab bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "repocrawler",
	Short: "Crawls all source control repositories you have access to and reports back.",
	Long: `	Repocrawler was created to get a quick lay of the land of what source control
	repositories you have access to. It allows you to see which repos might have the most
	commits or users. Along with what repositories are active (meaning a check-in within the last six months). 

	Hopefully this allows you to narrow down what pieces of your software inventory you 
	need to look at.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {
	// },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.repocrawler.yaml)")
	rootCmd.PersistentFlags().StringVar(&Organization, "organization", "", "A specific organization/project name that the crawl should be scoped to")
	rootCmd.PersistentFlags().StringVar(&OutputFile, "output", "repocrawler.json", "The file that should have the output recorded to")
	rootCmd.PersistentFlags().StringVar(&ScmURL, "scmUrl", "", "The API URL for the source control management software you want to crawl")
	rootCmd.PersistentFlags().StringVar(&TokenName, "tokenName", "GIT_TOKEN", "The environment variable name we should retrieve the token for API authentication")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".repocrawler" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".repocrawler")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func IsActiveRepo(lastCommit time.Time) bool {
	now := time.Now()
	difference := now.Sub(lastCommit)
	sixMonths := 24 * 30 * 6 // Roughly six months

	if int(difference.Hours()) > sixMonths {
		return false
	}
	return true
}

func WriteOutput(results []RepoInformation) {
	output, _ := json.MarshalIndent(results, "", " ")
	_ = ioutil.WriteFile(OutputFile, output, 0644)
}
