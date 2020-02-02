package discord

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Soumil07/authcord"
)

var (
	invalidSession      = fmt.Errorf("invalid session")
	unauthorizedSession = fmt.Errorf("unauthorized session")
)

type Session struct {
	session    *authcord.Session
	path       string
	authorized bool
}

func (c *Client) RequestLogin(w http.ResponseWriter, r *http.Request, path string) {
	session := authcord.New(c.clientId, c.clientSecret, c.redirect, []string{"identify"})

	c.add(&Session{
		session: session,
		path:    path,
	})

	http.Redirect(w, r, session.AuthURL()+"&prompt=none", http.StatusFound)
}

func (c *Client) AuthHandler(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseForm(); e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	state := r.FormValue("state")
	session, ok := c.pop(state)
	if !ok {
		http.Error(w, invalidSession.Error(), http.StatusBadRequest)
		return
	}

	if e := session.session.Callback(r.FormValue("code")); e != nil {
		http.Error(w, unauthorizedSession.Error(), http.StatusUnauthorized)
		return
	}

	user, e := session.session.User()
	if e != nil {
		http.Error(w, unauthorizedSession.Error(), http.StatusUnauthorized)
		return
	}

	name := getCookieName(session.path)
	roles, ok := c.roles[name]
	if !ok {
		http.Error(w, unauthorizedSession.Error(), http.StatusUnauthorized)
		return
	}

	if !c.bot.hasAnyRole(user.ID, roles) {
		http.Error(w, "You do not have access to this page :[", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    name,
		Value:   "yes",
		Path:    "/",
		Expires: time.Now().Add(time.Hour * 24),
	})

	http.Redirect(w, r, c.domain+session.path, http.StatusFound)
}

func (c *Client) IsLoggedIn(r *http.Request, path string) bool {
	name := strings.ToLower(getCookieName(path))
	for _, c := range r.Cookies() {
		if strings.ToLower(c.Name) == name {
			return true
		}
	}
	return false
}

func getCookieName(path string) string {
	if len(path) == 0 {
		return ""
	}
	if path[0] == '/' {
		path = path[1:]
	}
	i := strings.Index(path, "/")
	if i > 0 {
		path = path[:i]
	}
	return path
}
