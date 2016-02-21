package glc

import (
	"net/url"
	"regexp"
)

type GitHubURL struct {
	*url.URL
	User string
	Repo string
	ID string   // Branch or SHA or tag or ...
}

const (
	Host = "github.com"
	MasterBranch = "master"
)

var shaRegex = regexp.MustCompile(`^[a-f0-9]{7,40}$`)
var pathRegex = regexp.MustCompile(`^/([^/]+)/([^/]+)/blob/([^/]+)/(\S+)`)

func parseGitHubURL (rawurl string) (*GitHubURL, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	if u.Host != Host {
		return nil, nil
	}

	ghURL := &GitHubURL{URL: u}
	parts := pathRegex.FindStringSubmatch(u.Path)
	if parts == nil {
		return ghURL, nil
	}

	ghURL.User = parts[1]
	ghURL.Repo = parts[2]
	ghURL.ID = parts[3]

	return ghURL, nil
}

func (url *GitHubURL) HasSHA() bool {
	return shaRegex.MatchString(url.ID)
}

func (url *GitHubURL) IsPermanent() bool {
	return !url.IsDeep() || url.HasSHA()
}

// Needed?
func (url *GitHubURL) UsesMasterBranch() bool {
	return url.ID == MasterBranch
}

func (url *GitHubURL) IsDeep() bool {
	return url.User != "" && url.Repo != "" && url.ID != ""
}

func (url *GitHubURL) String() string {
	return url.URL.String()
}
