package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type GitHubEvent struct {
	Type    string       `json:"type"`
	Repo    Repo         `json:"repo"`
	Payload EventPayload `json:"payload"`
}

type Repo struct {
	Name string `json:"name"`
}

type EventPayload struct {
	Action  string `json:"action"`
	RefType string `json:"ref_type"`
	Commits []any  `json:"commits"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: github-activity <username>")
		return
	}

	username := os.Args[1]
	events, err := fetchUserActivity(username)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(events) == 0 {
		fmt.Println("No recent activity found.")
		return
	}

	fmt.Printf("Recent activity for %s:\n", username)
	for _, event := range events {
		printEvent(event)
	}
}

func fetchUserActivity(username string) ([]GitHubEvent, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/events", username)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GitHub API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("user '%s' not found", username)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status: %s", resp.Status)
	}

	var events []GitHubEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return events, nil
}

func printEvent(event GitHubEvent) {
	switch event.Type {
	case "PushEvent":
		count := len(event.Payload.Commits)
		fmt.Printf("- Pushed %d commits to %s\n", count, event.Repo.Name)

	case "IssuesEvent":
		action := event.Payload.Action
		if len(action) > 0 {
			action = string(action[0]-32) + action[1:]
		}
		fmt.Printf("- %s a new issue in %s\n", action, event.Repo.Name)

	case "WatchEvent":
		fmt.Printf("- Starred %s\n", event.Repo.Name)

	case "ForkEvent":
		fmt.Printf("- Forked %s\n", event.Repo.Name)

	case "CreateEvent":
		fmt.Printf("- Created %s in %s\n", event.Payload.RefType, event.Repo.Name)

	case "PullRequestEvent":
		action := event.Payload.Action
		if len(action) > 0 {
			action = string(action[0]-32) + action[1:]
		}
		fmt.Printf("- %s a pull request in %s\n", action, event.Repo.Name)

	default:
		fmt.Printf("- %s in %s\n", event.Type, event.Repo.Name)
	}
}
