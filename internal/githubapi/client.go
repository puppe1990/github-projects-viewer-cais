package githubapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.github.com"

var (
	ErrNotFound    = errors.New("github: not found")
	ErrForbidden   = errors.New("github: forbidden")
	ErrRateLimited = errors.New("github: rate limited")
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	me         *User
}

func New(baseURL, token string) *Client {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = defaultBaseURL
	}
	return &Client{
		baseURL:    base,
		token:      token,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) HasToken() bool {
	return c.token != ""
}

func (c *Client) Me(ctx context.Context) (User, error) {
	if c.me != nil {
		return *c.me, nil
	}
	var raw ghUser
	if err := c.get(ctx, "/user", &raw); err != nil {
		return User{}, err
	}
	user := raw.toUser()
	c.me = &user
	return user, nil
}

func (c *Client) User(ctx context.Context, login string) (User, error) {
	var raw ghUser
	if err := c.get(ctx, "/users/"+url.PathEscape(login), &raw); err != nil {
		return User{}, err
	}
	return raw.toUser(), nil
}

func (c *Client) UserRepos(ctx context.Context, login string) ([]Repo, error) {
	if c.isSelf(ctx, login) {
		return c.pagedRepos(ctx, "/user/repos?affiliation=owner")
	}
	return c.pagedRepos(ctx, "/users/"+url.PathEscape(login)+"/repos")
}

func (c *Client) UserOrgs(ctx context.Context, login string) ([]Org, error) {
	path := "/users/" + url.PathEscape(login) + "/orgs"
	if c.isSelf(ctx, login) {
		path = "/user/orgs"
	}
	return c.pagedOrgs(ctx, path)
}

func (c *Client) isSelf(ctx context.Context, login string) bool {
	if !c.HasToken() {
		return false
	}
	me, err := c.Me(ctx)
	return err == nil && strings.EqualFold(me.Login, login)
}

func (c *Client) OrgRepos(ctx context.Context, org string) ([]Repo, error) {
	return c.pagedRepos(ctx, "/orgs/"+url.PathEscape(org)+"/repos")
}

func (c *Client) Traffic(ctx context.Context, owner, repo string) (Traffic, error) {
	prefix := "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo) + "/traffic/"
	var views, clones ghTraffic
	if err := c.get(ctx, prefix+"views", &views); err != nil {
		return Traffic{}, err
	}
	if err := c.get(ctx, prefix+"clones", &clones); err != nil {
		return Traffic{}, err
	}
	return Traffic{
		Views:        views.Count,
		ViewUniques:  views.Uniques,
		Clones:       clones.Count,
		CloneUniques: clones.Uniques,
		Available:    true,
		ViewsByDay:   views.Views,
		ClonesByDay:  clones.Clones,
	}, nil
}

func (c *Client) pagedRepos(ctx context.Context, path string) ([]Repo, error) {
	var out []Repo
	next := firstPage(path)
	for page := 1; page <= 10 && next != ""; page++ {
		var raw []ghRepo
		link, err := c.getWithLink(ctx, next, &raw)
		if err != nil {
			return nil, err
		}
		for _, item := range raw {
			out = append(out, item.toRepo())
		}
		next = nextPage(link)
	}
	return out, nil
}

func (c *Client) pagedOrgs(ctx context.Context, path string) ([]Org, error) {
	var out []Org
	next := firstPage(path)
	for page := 1; page <= 10 && next != ""; page++ {
		var raw []ghOrg
		link, err := c.getWithLink(ctx, next, &raw)
		if err != nil {
			return nil, err
		}
		for _, item := range raw {
			out = append(out, item.toOrg())
		}
		next = nextPage(link)
	}
	return out, nil
}

func firstPage(path string) string {
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	if strings.Contains(path, "per_page=") {
		return path
	}
	return path + sep + "per_page=100&sort=updated&page=1"
}

func (c *Client) get(ctx context.Context, path string, dest any) error {
	_, err := c.getWithLink(ctx, path, dest)
	return err
}

func (c *Client) getWithLink(ctx context.Context, path string, dest any) (string, error) {
	target := path
	if strings.HasPrefix(path, "/") {
		target = c.baseURL + path
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("github %s: %w", path, err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return "", fmt.Errorf("github read %s: %w", path, err)
	}
	if err := statusError(res); err != nil {
		return "", err
	}
	if dest != nil {
		if err := json.Unmarshal(body, dest); err != nil {
			return "", fmt.Errorf("github decode %s: %w", path, err)
		}
	}
	return res.Header.Get("Link"), nil
}

func statusError(res *http.Response) error {
	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusForbidden:
		if remaining := res.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
			return ErrRateLimited
		}
		return ErrForbidden
	default:
		return fmt.Errorf("github: HTTP %d", res.StatusCode)
	}
}

func nextPage(link string) string {
	for _, part := range strings.Split(link, ",") {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, `rel="next"`) {
			continue
		}
		start := strings.Index(part, "<")
		end := strings.Index(part, ">")
		if start >= 0 && end > start {
			return part[start+1 : end]
		}
	}
	return ""
}
