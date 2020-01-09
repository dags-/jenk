package manager

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"github.com/dags-/downloads/jenkins"
	"github.com/dags-/err"
	"io"
	"net/http"
	"sync"
)

type Manager struct {
	address   string
	project   string
	client    *jenkins.Client
	lock      *sync.RWMutex
	downloads map[string]*Download
}

type Download struct {
	URL      string
	FileName string
}

func New(client *jenkins.Client, server, project string) *Manager {
	return &Manager{
		address:   server,
		project:   project,
		client:    client,
		lock:      &sync.RWMutex{},
		downloads: map[string]*Download{},
	}
}

func (m *Manager) GetAddress() string {
	return m.address + "/data"
}

func (m *Manager) ServeData(w http.ResponseWriter, r *http.Request) {
	data, e := m.client.GetJobData(m.project)
	if data == nil || e != nil && e.Present() {
		e.Warn()
		return
	}

	files := map[string]*Download{}
	for _, b := range data.Builds {
		for aid, a := range b.Artifacts {
			fid := getId(b.Timestamp, uint8(aid))
			url := m.client.GetArtifactURL(b, a)
			a.Path = m.address + "/file/" + fid
			files[fid] = &Download{
				URL:      url,
				FileName: a.FileName,
			}
		}
	}

	m.lock.Lock()
	m.downloads = files
	m.lock.Unlock()
	err.Encode(w, data).Warn()
}

func (m *Manager) ServeFile(w http.ResponseWriter, r *http.Request) {
	m.lock.Lock()
	dl, ok := m.downloads[r.URL.Path]
	m.lock.Unlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	e := m.download(w, dl)
	if e.Present() {
		e.Warn()
		http.NotFound(w, r)
		return
	}
}

func (m *Manager) download(w http.ResponseWriter, dl *Download) err.Error {
	rs, e := m.client.Get(dl.URL)
	if rs == nil || e != nil && e.Present() {
		return e
	}
	defer err.Close(rs.Body)

	for k, v := range rs.Header {
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+dl.FileName)

	_, er := io.Copy(w, rs.Body)

	if er != nil {
		return err.New(er)
	}

	return err.Nil()
}

func getId(buildId int64, artifactId uint8) string {
	buf := &bytes.Buffer{}
	en := base64.NewEncoder(base64.StdEncoding.WithPadding(base64.NoPadding), buf)
	binary.Write(en, binary.LittleEndian, buildId)
	binary.Write(en, binary.LittleEndian, artifactId)
	return buf.String()
}
