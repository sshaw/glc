package glc

import (
	"fmt"
	"time"
	"strings"

	"github.com/google/go-github/github"
)

const (
	helpURL string = "https://help.github.com/articles/getting-permanent-links-to-files"
	suggestionIntro string = "Annotating the GitHub links used by @%s with [permanent versions](" + helpURL + ").\n"

	repoURL string = "https://github.com/sshaw/glc"
	correctionFooter string = "— *GitHub links corrected by [The GLC](" + repoURL + ")*."
)

type Correction struct {
	OldURL GitHubURL
	NewURL GitHubURL
	Context string
}

type Event struct {
	gh *github.Client
	eventID string
	repoName string
	repoOwner string

	ID int
	Number int
	Type string
	Body string
	Repo string
	Actor string
	CreatedAt time.Time
	Corrections []*Correction
}

// Create a comment on the event that provides permanent versions of the non-permanent links.
func (e *Event) Comment() (int, error) {

	body := fmt.Sprintf(suggestionIntro, e.Actor)
	for _, correction := range(e.Corrections) {
		body += fmt.Sprintf("> %s\n\n", correction.Context)
		body += fmt.Sprintln(correction.NewURL.String())
	}

	comment, _, err := e.gh.Issues.CreateComment(e.repoOwner, e.repoName, e.Number, &github.IssueComment{Body: &body})
	if err != nil {
		return 0, err
	}

	return *comment.ID, nil
}

// Replace non-permanent GitHub links in with their permanent version and update the event.
func (e *Event) Correct() error {
	var err error

	body := e.Body
	for _, correction := range(e.Corrections) {
		body = strings.Replace(body, correction.OldURL.String(), correction.NewURL.String(), -1)
	}

	body += "\n\n" + correctionFooter

	if e.Type == PullRequest {
		_, _, err = e.gh.PullRequests.Edit(e.repoOwner, e.repoName, e.Number, &github.PullRequest{Body: &body})
	} else {
		_, _, err = e.gh.Issues.EditComment(e.repoOwner, e.repoName, e.ID, &github.IssueComment{Body: &body})
	}

	return err
}
