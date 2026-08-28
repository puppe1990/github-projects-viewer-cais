package githubapi

type User struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatar_url"`
	HTMLURL     string `json:"html_url"`
	Location    string `json:"location"`
	Company     string `json:"company"`
	Blog        string `json:"blog"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
}

type Repo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	HTMLURL     string `json:"html_url"`
	Homepage    string `json:"homepage"`
	UpdatedAt   string `json:"updated_at"`
	OwnerLogin  string `json:"owner_login"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Fork        bool   `json:"fork"`
	Archived    bool   `json:"archived"`
}

type Org struct {
	Login       string `json:"login"`
	AvatarURL   string `json:"avatar_url"`
	Description string `json:"description"`
}

type TrafficPoint struct {
	Timestamp string `json:"timestamp"`
	Count     int    `json:"count"`
	Uniques   int    `json:"uniques"`
}

type Traffic struct {
	Views        int            `json:"views"`
	ViewUniques  int            `json:"view_uniques"`
	Clones       int            `json:"clones"`
	CloneUniques int            `json:"clone_uniques"`
	Available    bool           `json:"available"`
	ViewsByDay   []TrafficPoint `json:"views_by_day"`
	ClonesByDay  []TrafficPoint `json:"clones_by_day"`
}

type ghUser struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatar_url"`
	HTMLURL     string `json:"html_url"`
	Location    string `json:"location"`
	Company     string `json:"company"`
	Blog        string `json:"blog"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
}

func (u ghUser) toUser() User {
	return User{
		Login:       u.Login,
		Name:        u.Name,
		Bio:         u.Bio,
		AvatarURL:   u.AvatarURL,
		HTMLURL:     u.HTMLURL,
		Location:    u.Location,
		Company:     u.Company,
		Blog:        u.Blog,
		PublicRepos: u.PublicRepos,
		Followers:   u.Followers,
		Following:   u.Following,
	}
}

type ghRepo struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	HTMLURL     string `json:"html_url"`
	Homepage    string `json:"homepage"`
	UpdatedAt   string `json:"updated_at"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Fork        bool   `json:"fork"`
	Archived    bool   `json:"archived"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func (r ghRepo) toRepo() Repo {
	return Repo{
		Name:        r.Name,
		FullName:    r.FullName,
		Description: r.Description,
		Language:    r.Language,
		HTMLURL:     r.HTMLURL,
		Homepage:    r.Homepage,
		UpdatedAt:   r.UpdatedAt,
		OwnerLogin:  r.Owner.Login,
		Stars:       r.Stars,
		Forks:       r.Forks,
		Fork:        r.Fork,
		Archived:    r.Archived,
	}
}

type ghOrg struct {
	Login       string `json:"login"`
	AvatarURL   string `json:"avatar_url"`
	Description string `json:"description"`
}

func (o ghOrg) toOrg() Org {
	return Org{Login: o.Login, AvatarURL: o.AvatarURL, Description: o.Description}
}

type ghTraffic struct {
	Count   int            `json:"count"`
	Uniques int            `json:"uniques"`
	Views   []TrafficPoint `json:"views"`
	Clones  []TrafficPoint `json:"clones"`
}
