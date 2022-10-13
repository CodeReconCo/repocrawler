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
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"os"
	"strings"
	"time"
)

type GitlabCrawler struct {
	client *gitlab.Client
}

// gitlabCmd represents the gitlab command
var gitlabCmd = &cobra.Command{
	Use:   "gitlab",
	Short: "Executes a crawl against Gitlab",
	Long:  `Executes a crawl against Gitlab.com or Gitlab instance`,
	Run: func(cmd *cobra.Command, args []string) {
		crawler := GitlabCrawler{}
		crawler.runScan()
	},
}

func init() {
	rootCmd.AddCommand(gitlabCmd)
}

func (gc GitlabCrawler) calculateAverageCommits(numCommits int, createdOn time.Time) float64 {
	numDays := int(time.Now().Sub(createdOn).Hours() / 24)
	return float64(numCommits) / float64(numDays)
}

func (gc GitlabCrawler) retrieveProjectLanguages(projectId int) string {
	languages, _, _ := gc.client.Projects.GetProjectLanguages(projectId)
	var langs []string

	for k, _ := range *languages {
		langs = append(langs, k)
	}
	return strings.Join(langs, ",")
}

func (gc GitlabCrawler) retrieveNumberOfMembers(projectId int) int {

	opt := &gitlab.ListProjectUserOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
	}
	numMembers := 0
	for {

		users, resp, err := gc.client.Projects.ListProjectsUsers(projectId, opt)
		if err != nil {
			fmt.Println("Failed to get members: ", err)
		}

		numMembers += len(users)

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		opt.Page = resp.NextPage
	}
	return numMembers
}

func (gc GitlabCrawler) runScan() {
	token := os.Getenv(TokenName)
	if token == "" {
		fmt.Println(TokenName + " is empty")
		return
	}

	var err error
	if ScmURL != "" {
		gc.client, err = gitlab.NewClient(token, gitlab.WithBaseURL(ScmURL))
	} else {
		gc.client, err = gitlab.NewClient(token)
	}

	if err != nil {
		fmt.Println("Failed to create client: ", err)
	}

	var results []RepoInformation
	trueValue := true
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
		Membership: &trueValue,
		Statistics: &trueValue,
	}
	// Go through the Groups...
	for {
		projects, resp, err := gc.client.Projects.ListProjects(opt)
		if err != nil {
			fmt.Println("Failed to get projects: ", err)
		}

		for _, p := range projects {
			ri := RepoInformation{}
			ri.Name = p.Name

			var groupNames []string
			for _, group := range p.SharedWithGroups {
				groupNames = append(groupNames, group.GroupName)
			}
			ri.Organization = strings.Join(groupNames, ",")
			ri.URL = p.WebURL
			ri.Private = !p.Public
			ri.NumberOfForks = p.ForksCount
			ri.NumberOfStars = p.StarCount
			ri.CreatedOn = *p.CreatedAt
			ri.LastCommit = *p.LastActivityAt
			ri.IsActive = IsActiveRepo(ri.LastCommit)
			if p.Archived {
				ri.Status = "Archived"
			}
			ri.NumberOfCommits = p.Statistics.CommitCount
			ri.AverageCommitsPerDay = gc.calculateAverageCommits(ri.NumberOfCommits, ri.CreatedOn)
			ri.Languages = gc.retrieveProjectLanguages(p.ID)
			ri.NumberOfCollaborators = gc.retrieveNumberOfMembers(p.ID)

			results = append(results, ri)
			WriteOutput(results)
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		opt.Page = resp.NextPage
	}

}
