package jenkins

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/dags-/jenk/err"
)

const (
	suffix = "/api/json"
)

var (
	emptyBuild = &BuildData{
		Build: &Build{
			Number:    -1,
			Result:    "FAIL",
			Artifacts: []*Artifact{},
		},
		Commit: &CommitData{
			Hash: "unknown",
			URL:  "",
		},
	}
)

type Config struct {
	Server string
	User   string
	Token  string
}

type Client struct {
	server string
	user   string
	token  string
}

type JobData struct {
	Name   string       `json:"name"`
	Builds []*BuildData `json:"builds"`
}

type BuildData struct {
	*Build
	Commit *CommitData `json:"commit"`
}

type CommitData struct {
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

func NewClient(c *Config) *Client {
	return &Client{
		user:   c.User,
		token:  c.Token,
		server: c.address(),
	}
}

func (c *Client) Get(path string) (*http.Response, err.Error) {
	rq, e := err.Request(http.MethodGet, path, nil)
	if e.Present() {
		return nil, e
	}
	defer err.Close(rq.Body)

	rq.SetBasicAuth(c.user, c.token)
	rs, e := err.Send(rq)
	if e.Present() {
		return nil, e
	}

	return rs, err.Nil()
}

func (c *Client) GetJob(name string) (*Job, err.Error) {
	var job Job
	endpoint := c.getEndpoint("job", name, suffix)
	e := c.get(endpoint, &job)
	if e.Present() {
		return nil, e
	}
	return &job, e
}

func (c *Client) GetBuild(b *BuildMeta) (*Build, err.Error) {
	var build Build
	endpoint := b.URL + suffix
	e := c.get(endpoint, &build)
	if e.Present() {
		return nil, e
	}
	return &build, e
}

func (c *Client) GetGit(b *Build) (*Git, err.Error) {
	var git Git
	endpoint := b.URL + "git" + suffix
	e := c.get(endpoint, &git)
	if e.Present() {
		return nil, e
	}
	return &git, err.Nil()
}

func (c *Client) GetJobData(name string) (*JobData, err.Error) {
	j, e := c.GetJob(name)
	if e.Present() {
		return nil, e
	}

	ch := make(chan *BuildData)
	data := &JobData{
		Name:   name,
		Builds: make([]*BuildData, 0),
	}

	go c.getBuildsAsync(j, ch)

	for b := range ch {
		data.Builds = append(data.Builds, b)
	}

	sort.Slice(data.Builds, func(i, j int) bool {
		b0 := data.Builds[i]
		b1 := data.Builds[j]
		return b0.Number > b1.Number
	})

	return data, err.Nil()
}

func (c *Client) GetCommit(b *Build) *CommitData {
	git, e := c.GetGit(b)
	if e.Present() || len(git.RemoteURL) != 1 {
		return emptyBuild.Commit
	}

	url := git.RemoteURL[0]
	url = strings.TrimSuffix(url, ".git")
	url = url + "/commit/" + git.Revision.SHA1

	commit := &CommitData{
		Hash: git.Revision.SHA1,
		URL:  url,
	}

	return commit
}

func (c *Client) get(path string, i interface{}) err.Error {
	rs, e := c.Get(path)
	if e.Present() {
		return e
	}
	defer err.Close(rs.Body)
	return err.Decode(rs.Body, i)
}

func (c *Client) GetArtifactURL(build *Build, a *Artifact) string {
	return build.URL + "/artifact/" + a.Path
}

func (c *Client) getEndpoint(path ...string) string {
	return c.server + strings.Join(path, "/") + suffix
}

func (c *Client) getBuildsAsync(job *Job, ch chan *BuildData) {
	wg := &sync.WaitGroup{}
	wg.Add(len(job.Builds))
	for _, build := range job.Builds {
		go c.getBuildAsync(build, ch, wg)
	}
	wg.Wait()
	close(ch)
}

func (c *Client) getBuildAsync(meta *BuildMeta, ch chan *BuildData, wg *sync.WaitGroup) {
	defer wg.Done()
	build, e := c.GetBuild(meta)
	if e.Present() {
		ch <- emptyBuild
		return
	}

	sort.Slice(build.Artifacts, func(i, j int) bool {
		a0 := build.Artifacts[i]
		a1 := build.Artifacts[j]
		return strings.Compare(a0.FileName, a1.FileName) < 0
	})

	ch <- &BuildData{
		Build:  build,
		Commit: c.GetCommit(build),
	}
}

func (c *Config) address() string {
	address := c.Server
	if len(address) == 0 {
		return address
	}
	if c.Server[len(address)-1] != '/' {
		address += "/"
	}
	return address
}
