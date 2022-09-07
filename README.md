# Repository Crawler
Repository crawler was created to easily get a usage snapshot of all the repositories that you have access to. This is useful when either joining a new company, an acquisition or just making sure you are focusing on the most important software being written by your organization.

## How to Compile
* Make sure Go is installed
* Put a Github or Gitlab API token in an environment variable
* Run the tool `go run .` there are two main commands `github` or `gitlab` as an example if I wanted to scan an example org in github the command would look like `go run . github --organization EXAMPLE`.

## Output 
* By default the output will be written to `repocrawler.json`, however, you can change that with the flag `--output` and put a different file name. The output is only in JSON.
* Fields that are outputted are:
	* Organization - The organization or groups associated with this repository
	* Name - Name of the repository
	* URL - The web URL for this repository
	* Private - If the repository is private or not
	* Number of Collaborators - Number of collaborators/users who have access to this repository
	* Number of Forks - Number of forks for a given repository
	* Number of Commits - Number of commits this repository has
	* Number of Stars - Number of stars that this repository has
	* Number of Watchers - Number of watchers (note this only works for Github)
	* Languages - The languages in this repository
	* IsActive - If the repository has had a check-in within the last six months
	* Last Commit - The date of the last commit
	* Created On - When the repository was created
	* Average Commits Per Day - The average commits that happen per-day
	* Status - Status of this repository

### Example Output
```
[
 {
  "Organization": "codereconco",
  "Name": "repocrawler",
  "URL": "https://github.com/CodeReconCo/repocrawler",
  "Private": false,
  "NumberOfCollaborators": 1,
  "NumberOfForks": 0,
  "NumberOfCommits": 7,
  "NumberOfStars": 1,
  "NumberOfWatchers": 1,
  "Languages": "Go",
  "IsActive": true,
  "LastCommit": "2020-06-02T04:18:13Z",
  "CreatedOn": "2020-04-22T13:06:47Z",
  "AverageCommitsPerDay": 0.04827586206896552,
  "Status": ""
 },
 {
  "Organization": "codereconco",
  "Name": "githubsecurityauditor",
  "URL": "https://github.com/CodeReconCo/githubsecurityauditor",
  "Private": false,
  "NumberOfCollaborators": 1,
  "NumberOfForks": 0,
  "NumberOfCommits": 7,
  "NumberOfStars": 0,
  "NumberOfWatchers": 0,
  "Languages": "Go",
  "IsActive": true,
  "LastCommit": "2020-09-15T03:36:51Z",
  "CreatedOn": "2020-06-09T04:20:22Z",
  "AverageCommitsPerDay": 0.07216494845360824,
  "Status": ""
 }
]
```