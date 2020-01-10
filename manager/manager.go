package manager

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net/http"
	"sync"
	"time"

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
	client    *jenkins.Client
	lock      *sync.RWMutex
	nextCheck time.Time
	cache     map[string]*cache
	downloads map[string]*download
}

type download struct {
	url      string
	fileName string
	expires  time.Time
}

type cache struct {
	expires time.Time
	data    *jenkins.JobData
}

func New(client *jenkins.Client) *Manager {
	return &Manager{
		client:    client,
		lock:      &sync.RWMutex{},
		nextCheck: time.Now().Add(checkInterval),
		cache:     map[string]*cache{},
		downloads: map[string]*download{},
	}
}

func (m *Manager) ServeData(w http.ResponseWriter, r *http.Request) {
	// periodically check and expire download links
	m.expireLinks()

	// fetch cached data
	m.lock.Lock()
	data, ok := m.cache[r.URL.Path]
	m.lock.Unlock()

	// cached data exists but has expired
	if ok && data.expires.Before(time.Now()) {
		data = nil
		ok = false
		m.lock.Lock()
		delete(m.cache, r.URL.Path)
		m.lock.Unlock()
	}

	// cached value didn't exist or expired
	if data == nil || !ok {
		newData, e := m.getData(r.URL.Path)
		if newData == nil || e.Present() {
			e.Warn()
			http.NotFound(w, r)
			return
		}

		// cache the data
		m.lock.Lock()
		m.cache[r.URL.Path] = newData
		m.lock.Unlock()

		// set the data
		data = newData
	}

	// send to client
	e := err.Encode(w, data.data)
	e.Warn()
}

func (m *Manager) ServeFile(w http.ResponseWriter, r *http.Request) {
	// periodically check and expire download links
	m.expireLinks()

	// get the download link
	m.lock.Lock()
	dl, ok := m.downloads[r.URL.Path]
	m.lock.Unlock()

	// link didn't exist or expired
	if !ok {
		http.NotFound(w, r)
		return
	}

	// handle the download process & any errors
	e := m.download(w, dl)
	if e.Present() {
		e.Warn()
		return
	}
}

func (m *Manager) expireLinks() {
	now := time.Now()

	m.lock.Lock()
	defer m.lock.Unlock()

	if m.nextCheck.Before(now) {
		for k, v := range m.downloads {
			if v.expires.Before(now) {
				delete(m.downloads, k)
			}
		}
		m.nextCheck = now.Add(checkInterval)
	}
}

func (m *Manager) getData(name string) (*cache, err.Error) {
	data, e := m.client.GetJobData(name)
	if data == nil || e.Present() {
		return nil, e
	}

	now := time.Now()
	downloadTimout := now.Add(downloadExpireTime)
	for _, b := range data.Builds {
		for aid, a := range b.Artifacts {
			fid := getId(b.Timestamp, uint8(aid))
			url := m.client.GetArtifactURL(b.Build, a)
			a.Path = "/file/" + fid
			m.lock.Lock()
			m.downloads[fid] = &download{
				url:      url,
				fileName: a.FileName,
				expires:  downloadTimout,
			}
			m.lock.Unlock()
		}
	}

	result := &cache{
		expires: now.Add(dataExpireTime),
		data:    data,
	}

	return result, err.Nil()
}

func (m *Manager) download(w http.ResponseWriter, dl *download) err.Error {
	rs, e := m.client.Get(dl.url)
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

func getId(buildId int64, artifactId uint8) string {
	buf := &bytes.Buffer{}
	en := base64.NewEncoder(base64.StdEncoding.WithPadding(base64.NoPadding), buf)
	binary.Write(en, binary.LittleEndian, buildId)
	binary.Write(en, binary.LittleEndian, artifactId)
	return buf.String()
}
