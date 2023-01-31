package main

import (
	"fmt"

	"github.com/cli/go-gh"
	"github.com/mgutz/ansi"
)

var (
	confusion    = "😕"
	mergeable    = map[string]string{"": "⏳", "CONFLICTING": "❌", "MERGEABLE": "✅"}
	reviewed     = map[string]string{"": "❗", "CHANGES_REQUESTED": "❌", "APPROVED": "✅", "REVIEW_REQUIRED": "⏳"}
	checkResults = map[string]string{
		"":                "❗",
		"IN_PROGRESS":     "⚙️ ",
		"QUEUED":          "⏳",
		"REQUESTED":       "⏳",
		"ACTION_REQUIRED": "🛑",
		"CANCELLED":       "🛑",
		"FAILURE":         "❌",
		"NEUTRAL":         "🔵",
		"SKIPPED":         "🔵",
		"STARTUP_FAILURE": "❌",
		"SUCCESS":         "✅",
		"TIMED_OUT":       "🛑",
	}
)

func main() {
	client, err := gh.GQLClient(nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	var query struct {
		Viewer struct {
			PullRequests struct {
				Nodes []*struct {
					Number  int
					Title   string
					HeadRef struct {
						Name string
					}
					Mergeable        string
					MergeStateStatus string
					URL              string
					ReviewDecision   string
					Repository       struct {
						Name  string
						Owner struct {
							Login string
						}
					}
					Commits struct {
						Nodes []struct {
							Commit struct {
								StatusCheckRollup struct {
									State string
								}
							}
						}
					} `graphql:"commits(last: 1)"`
				}
			} `graphql:"pullRequests(last: 50, states: [OPEN], orderBy: {field: CREATED_AT, direction: ASC})"`
		}
	}

	variables := map[string]interface{}{}
	err = client.Query("MyPullRequests", &query, variables)
	if err != nil {
		fmt.Printf("Error calling graphql api: %s\n", err)
		return
	}
	var mrl, msl int
	for _, pr := range query.Viewer.PullRequests.Nodes {
		if pr.Repository.Owner.Login != "github" {
			continue
		}
		l := 1 + len(pr.Repository.Name) + len(pr.HeadRef.Name)
		if l > mrl {
			mrl = l
		}
		if len(pr.Title) > msl {
			msl = len(pr.Title)
		}
	}
	if len(query.Viewer.PullRequests.Nodes) > 0 {
		fmt.Printf("%*s", mrl+11, "⛙ 👀✔️ \n")
	}
	for _, pr := range query.Viewer.PullRequests.Nodes {
		if pr.Repository.Owner.Login != "github" {
			continue
		}
		ident := fmt.Sprintf("%s/%s", pr.Repository.Name, pr.HeadRef.Name)
		mergeStatus, ok := mergeable[pr.Mergeable]
		if !ok {
			mergeStatus = confusion
		}
		reviewStatus, ok := reviewed[pr.ReviewDecision]
		if !ok {
			reviewStatus = confusion
		}
		checkStatus := checkResults[pr.Commits.Nodes[0].Commit.StatusCheckRollup.State]

		fmt.Printf("[%-*s]  %s%s%s  %-*s %s\n", mrl, ident, mergeStatus, reviewStatus, checkStatus, msl, pr.Title, ansi.Color(pr.URL, "white+d"))
	}
}
