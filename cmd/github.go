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
	"context"
	"fmt"
	"github.com/google/go-github/v31/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"os"
	"time"
)

type GithubCrawler struct {
	client *github.Client
}

// githubCmd represents the github command
var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Crawls github",
	Long:  `Executes a crawl against Github or Github Enterprise instance`,
	Run: func(cmd *cobra.Command, args []string) {
		crawler := GithubCrawler{}
		crawler.runScan()
	},
}

func init() {
	rootCmd.AddCommand(githubCmd)
}

func (gc GithubCrawler) getCollaboratorInformation(ctx context.Context, org, repo *string) int {
	opt := &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	totalCollaborators := 0
	for {
		collaborators, resp, err := gc.client.Repositories.ListCollaborators(ctx, *org, *repo, opt)
		if _, ok := err.(*github.RateLimitError); ok {
			fmt.Println("Hit rate limit, sleeping for sixty minutes")
			time.Sleep(60 * time.Minute)
		}

		if err != nil {
			fmt.Println(err)
			return totalCollaborators
		}
		totalCollaborators += len(collaborators)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return totalCollaborators
}

func (gc GithubCrawler) getCommitInformation(ctx context.Context, org, repo *string, repoCreatedOn time.Time) (numCommits int, avgCommitsPerDay float64) {
	opt := &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		commits, resp, err := gc.client.Repositories.ListCommits(ctx, *org, *repo, opt)

		if _, ok := err.(*github.RateLimitError); ok {
			fmt.Println("Hit rate limit, sleeping for sixty minutes")
			time.Sleep(60 * time.Minute)
			err = nil
		}

		if err != nil {
			fmt.Println(err)
			return
		}

		numCommits += len(commits)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	numDays := int(time.Now().Sub(repoCreatedOn).Hours() / 24)
	avgCommitsPerDay = float64(numCommits) / float64(numDays)
	return
}

func (gc GithubCrawler) crawlRepositories(ctx context.Context, org *string, results []RepoInformation) []RepoInformation {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	for {
		repos, resp, err := gc.client.Repositories.ListByOrg(ctx, *org, opt)

		if _, ok := err.(*github.RateLimitError); ok {
			fmt.Println("Hit rate limit, sleeping for sixty minutes")
			time.Sleep(60 * time.Minute)
			err = nil
		}

		if err != nil {
			fmt.Println(err)
			return results
		}
		for _, repo := range repos {
			ri := RepoInformation{}
			ri.Name = *repo.Name
			ri.Organization = *org
			ri.URL = *repo.HTMLURL
			ri.Private = *repo.Private
			ri.NumberOfForks = *repo.ForksCount
			ri.NumberOfStars = *repo.StargazersCount
			ri.NumberOfWatchers = *repo.WatchersCount
			ri.Languages = *repo.Language
			ri.CreatedOn = repo.CreatedAt.Time
			ri.LastCommit = repo.UpdatedAt.Time
			ri.IsActive = IsActiveRepo(repo.UpdatedAt.Time)
			if *repo.Archived {
				ri.Status = "Archived"
			} else if *repo.Disabled {
				ri.Status = "Disabled"
			}
			ri.NumberOfCommits, ri.AverageCommitsPerDay = gc.getCommitInformation(ctx, org, repo.Name, ri.CreatedOn)
			ri.NumberOfCollaborators = gc.getCollaboratorInformation(ctx, org, repo.Name)
			results = append(results, ri)
		}
		WriteOutput(results)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return results
}

func (gc GithubCrawler) runScan() {
	token := os.Getenv(TokenName)
	if token == "" {
		fmt.Println(TokenName + " is empty")
		return
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	if ScmURL != "" {
		var clientErr error
		gc.client, clientErr = github.NewEnterpriseClient(ScmURL, ScmURL, tc)

		if clientErr != nil {
			fmt.Println(clientErr)
			return
		}
	} else {
		gc.client = github.NewClient(tc)
	}

	// Get organizations
	opt := &github.ListOptions{PerPage: 10}
	var results []RepoInformation

	if Organization != "" {
		gc.crawlRepositories(ctx, &Organization, results)
	} else {
		for {
			orgs, resp, err := gc.client.Organizations.List(ctx, Organization, opt)

			if err != nil {
				fmt.Println(err)
				return
			}

			if _, ok := err.(*github.RateLimitError); ok {
				fmt.Println("Hit rate limit, sleeping for sixty minutes")
				time.Sleep(60 * time.Minute)
			}

			for _, org := range orgs {
				tmpResults := gc.crawlRepositories(ctx, org.Login, results)
				results = append(results, tmpResults...)
			}

			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	}
}
