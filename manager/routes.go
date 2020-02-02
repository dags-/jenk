package manager

import (
	"net/http"
	"strings"
	"time"

	"github.com/dags-/jenk/err"
)

func (m *Manager) ServeDir(dir http.Dir) func(http.ResponseWriter, *http.Request) {
	handler := http.FileServer(dir)
	return func(w http.ResponseWriter, r *http.Request) {
		if !m.discord.IsLoggedIn(r, r.URL.Path) {
			m.discord.RequestLogin(w, r, r.URL.Path)
			return
		}

		if len(r.URL.Path) > 1 {
			// disallow sub path
			i := strings.LastIndex(r.URL.Path, "/")
			if i > 0 {
				r.URL.Path = r.URL.Path[i:]
			} else if !strings.ContainsRune(r.URL.Path, '.') && !strings.HasSuffix(r.URL.Path, "/") {
				r.URL.Path += "/"
				http.Redirect(w, r, r.URL.Path, http.StatusPermanentRedirect)
				return
			}
		}

		handler.ServeHTTP(w, r)
	}
}

func (m *Manager) ServeData(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")

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
		newData, e := m.getJob(r.URL.Path)
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

	// send to jenkins
	err.Encode(w, data.data).Log()
}

func (m *Manager) ServeFile(w http.ResponseWriter, r *http.Request) {
	// get the download link
	m.lock.Lock()
	dl, ok := m.downloads[r.URL.Path]
	m.lock.Unlock()

	if !m.discord.IsLoggedIn(r, dl.project) {
		m.discord.RequestLogin(w, r, "/file/"+r.URL.Path)
		return
	}

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
