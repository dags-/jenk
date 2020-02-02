package manager

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/dags-/jenk/discord"
	"github.com/dags-/jenk/err"
	"github.com/dags-/jenk/jenkins"
)

const (
	// frequency that download links should be checked for expiration
	checkInterval = time.Hour * 6

	// time after which cached data should expire
	dataExpireTime = time.Minute * 10

	// time after which download links should expire
	downloadExpireTime = time.Hour * 24 * 10
)

type Manager struct {
	address   string
	jenkins   *jenkins.Client
	discord   *discord.Client
	lock      *sync.RWMutex
	nextCheck time.Time
	cache     map[string]*cache
	downloads map[string]*download
}

type download struct {
	url      string
	fileName string
	project  string
	expires  time.Time
}

type cache struct {
	expires time.Time
	data    *jenkins.JobData
}

func New(jenkins *jenkins.Client, discord *discord.Client) *Manager {
	return &Manager{
		jenkins:   jenkins,
		discord:   discord,
		lock:      &sync.RWMutex{},
		nextCheck: time.Now().Add(checkInterval),
		cache:     map[string]*cache{},
		downloads: map[string]*download{},
	}
}

func (m *Manager) StartWatchdog() {
	go m.expireLinks()
}

func (m *Manager) getJob(name string) (*cache, err.Error) {
	data, e := m.jenkins.GetJobData(name)
	if data == nil || e.Present() {
		return nil, e
	}

	now := time.Now()
	downloadTimout := now.Add(downloadExpireTime)

	m.lock.Lock()
	for _, b := range data.Builds {
		for aid, a := range b.Artifacts {
			fid := getId(b.Timestamp, uint8(aid))
			url := m.jenkins.GetArtifactURL(b.Build, a)
			a.Path = "/file/" + fid
			m.downloads[fid] = &download{
				url:      url,
				project:  name,
				fileName: a.FileName,
				expires:  downloadTimout,
			}
		}
	}
	m.lock.Unlock()

	result := &cache{
		expires: now.Add(dataExpireTime),
		data:    data,
	}

	return result, err.Nil()
}

func (m *Manager) download(w http.ResponseWriter, dl *download) err.Error {
	rs, e := m.jenkins.Get(dl.url)
	if rs == nil || e.Present() {
		return e
	}
	defer err.Close(rs.Body)

	for k, v := range rs.Header {
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+dl.fileName)

	_, _ = io.Copy(w, rs.Body)

	return err.Nil()
}

func (m *Manager) expireLinks() {
	t := time.NewTicker(checkInterval)
	defer t.Stop()

	for now := range t.C {
		m.lock.Lock()
		for k, v := range m.downloads {
			if v.expires.Before(now) {
				delete(m.downloads, k)
			}
		}
		m.lock.Unlock()
	}
}

func getId(buildId int64, artifactId uint8) string {
	buf := &bytes.Buffer{}
	en := base64.NewEncoder(base64.StdEncoding.WithPadding(base64.NoPadding), buf)
	binary.Write(en, binary.LittleEndian, buildId)
	binary.Write(en, binary.LittleEndian, artifactId)
	return buf.String()
}
