package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/xanzy/go-gitlab"
)

// extractProjectPath extracts the project path from a GitLab issues link
func extractProjectPath(issuesLink string) (string, error) {
	u, err := url.Parse(issuesLink)
	if err != nil {
		log.Printf("Error parsing GitLab issues link: %v\n", err)
		return "", err
	}

	pathSegments := strings.Split(u.Path, "/")
	if len(pathSegments) < 5 {
		err := fmt.Errorf("invalid GitLab issues link: %s", issuesLink)
		log.Printf("Error extracting project path: %v\n", err)
		return "", err
	}

	return strings.Join(pathSegments[1:5], "/"), nil
}

// readGitLabTokenFromFile reads GitLab token from file
func readGitLabTokenFromFile(tokenFile string) (string, error) {
	tokenBytes, err := os.ReadFile(tokenFile)
	if err != nil {
		log.Printf("Error reading GitLab token file: %v\n", err)
		return "", err
	}
	return strings.TrimSpace(string(tokenBytes)), nil
}

func main() {

	var releaseLabel string
	var readyForTest bool
	var blockerLabel string
	flag.StringVar(&releaseLabel, "release", "", "Specify the release label")
	flag.BoolVar(&readyForTest, "ready-for-test", false, "Flag to filter READY-FOR-TEST issues")
	flag.StringVar(&blockerLabel, "blocker", "", "Specify the blocker label (staging-upgrade or production-upgrade)")

	flag.Parse()

	if releaseLabel == "" && readyForTest {
		log.Fatal("Release label is required when filtering by READY-FOR-TEST")
	}
	log.Println("Starting the program...")

	// Set the path to the GitLab token file
	tokenFile := "~/.gitlab"
	resolveHome(&tokenFile)
	log.Printf("Token file path: %s\n", tokenFile)

	// Read GitLab token from file
	gitlabToken, err := readGitLabTokenFromFile(tokenFile)
	if err != nil {
		log.Fatalf("Failed to read GitLab token from file: %v", err)
	}

	// Remove the "token:" prefix if present
	gitlabToken = strings.TrimPrefix(gitlabToken, "token:")
	// log.Printf("GitLab token: %s\n", gitlabToken)

	gitlabIssuesLink := "https://gitlab.com/f5/volterra/support/technical/-/issues/?sort=created_date&state=opened"
	log.Printf("GitLab issues link: %s\n", gitlabIssuesLink)

	// Extract project path from the GitLab issues link
	projectPath, err := extractProjectPath(gitlabIssuesLink)
	if err != nil {
		log.Fatalf("Failed to extract project path: %v", err)
	}

	log.Printf("Project path: %s\n", projectPath)

	// Create a GitLab client
	git, err := gitlab.NewClient(gitlabToken)
	if err != nil {
		log.Fatalf("Failed to create GitLab client: %v", err)
	}

	log.Println("GitLab client created successfully.")
	opened := "opened"
	var lab, notLabels gitlab.LabelOptions

	if releaseLabel != "" {
		lab = append(lab, releaseLabel)
	}

	if readyForTest {
		lab = append(lab, "READY-FOR-TEST")
	} else {
		notLabels = append(notLabels, "READY-FOR-TEST")
	}

	if blockerLabel != "" {
		lab = append(lab, blockerLabel)
	}

	issues, _, err := git.Issues.ListProjectIssues(projectPath, &gitlab.ListProjectIssuesOptions{
		State:       &opened,
		ListOptions: gitlab.ListOptions{PerPage: 100},
		Labels:      &lab,
		NotLabels:   &notLabels,
	})
	if err != nil {
		log.Fatalf("Failed to list project issues: %v", err)
	}

	log.Println("Project issues listed successfully.")

	// Create a CSV file
	currentDateTime := time.Now().Format("2006-01-02_15-04-05")
	csvFileName := fmt.Sprintf("issues_output_%s.csv", currentDateTime)

	outputFile, err := os.Create(csvFileName)
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer outputFile.Close()

	log.Printf("CSV file created: %s\n", csvFileName)

	// Create a CSV writer
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write header to CSV
	header := []string{"Issue", "Summary", "Assignee", "Author", "Date of Creation"}
	if err := writer.Write(header); err != nil {
		log.Fatalf("Failed to write header to CSV: %v", err)
	}

	log.Println("Header written to CSV successfully.")

	// Write details to CSV for each issue
	for _, issue := range issues {
		assignee := "Unassigned"
		if issue.Assignee != nil {
			assignee = issue.Assignee.Name
		}

		// Fetch more details about the issue
		detailedIssue, _, err := git.Issues.GetIssue(projectPath, issue.IID)
		if err != nil {
			log.Fatalf("Failed to get detailed issue: %v", err)
		}

		// Create a formatted hyperlink for the issue number
		issueLink := fmt.Sprintf("[#%d](%s)", issue.IID, detailedIssue.WebURL)

		// Write issue details to CSV
		row := []string{
			issueLink,
			detailedIssue.Title,
			assignee,
			detailedIssue.Author.Username,
			detailedIssue.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if err := writer.Write(row); err != nil {
			log.Fatalf("Failed to write row to CSV: %v", err)
		}
	}

	fmt.Println("CSV file created successfully.")
}

// resolveHome replaces ~ with current home dir
func resolveHome(path *string) {
	expandedPath, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return
	}
	*path = strings.Replace(*path, "~", expandedPath, 1)
}
