package jenkins

import (
	"github.com/dags-/err"
	"net/http"
	"sort"
	"strings"
	"sync"
)

const (
	suffix = "/api/json"
)

type Client struct {
	server string
	user   string
	token  string
}

type JobData struct {
	Name   string   `json:"name"`
	Builds []*Build `json:"builds"`
}

func NewClient(server, usr, token string) *Client {
	if server[len(server)-1] != '/' {
		server += "/"
	}
	return &Client{
		server: server,
		user:   usr,
		token:  token,
	}
}

func (c *Client) Get(path string) (*http.Response, err.Error) {
	rq, e := err.Request(http.MethodGet, path, nil)
	if rq == nil || e != nil && e.Present() {
		return nil, e
	}
	defer err.Close(rq.Body)

	rq.SetBasicAuth(c.user, c.token)
	rs, e := err.Send(rq)
	if rs == nil || e != nil && e.Present() {
		return nil, e
	}
	return rs, err.Nil()
}

func (c *Client) GetJob(name string) (*Job, err.Error) {
	var job Job
	endpoint := c.getEndpoint("job", name, suffix)
	e := c.get(endpoint, &job)
	return &job, e
}

func (c *Client) GetBuild(b *BuildMeta) (*Build, err.Error) {
	var build Build
	endpoint := b.URL + suffix
	e := c.get(endpoint, &build)
	return &build, e
}

func (c *Client) GetJobData(name string) (*JobData, err.Error) {
	j, e := c.GetJob(name)
	if j == nil || e != nil && e.Present() {
		return nil, e
	}

	ch := make(chan *Build)
	data := &JobData{
		Name:   name,
		Builds: make([]*Build, 0),
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

func (c *Client) get(path string, i interface{}) err.Error {
	rs, e := c.Get(path)
	if rs == nil || e != nil && e.Present() {
		return e
	}
	defer err.Close(rs.Body)
	return err.Decode(rs.Body, i)
}

func (c *Client) getEndpoint(path ...string) string {
	return c.server + strings.Join(path, "/") + suffix
}

func (c *Client) GetArtifactURL(build *Build, a *Artifact) string {
	return build.URL + "/artifact/" + a.Path
}

func (c *Client) getBuildsAsync(job *Job, ch chan *Build) {
	wg := &sync.WaitGroup{}
	wg.Add(len(job.Builds))
	for _, build := range job.Builds {
		go c.getBuildAsync(build, ch, wg)
	}
	wg.Wait()
	close(ch)
}

func (c *Client) getBuildAsync(meta *BuildMeta, ch chan *Build, wg *sync.WaitGroup) {
	defer wg.Done()
	b, e := c.GetBuild(meta)
	if b != nil && (e == nil || !e.Present()) {
		sort.Slice(b.Artifacts, func(i, j int) bool {
			a0 := b.Artifacts[i]
			a1 := b.Artifacts[j]
			return strings.Compare(a0.FileName, a1.FileName) < 0
		})
		ch <- b
		return
	}
	ch <- emptyBuild
}
