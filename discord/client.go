package discord

import (
	"sync"
)

type Config struct {
	Domain       string
	ClientId     string
	ClientSecret string
	BotToken     string
	Roles        []string
}

type Client struct {
	roles        []string
	domain       string
	redirect     string
	clientId     string
	clientSecret string
	bot          *Bot
	lock         *sync.RWMutex
	permissions  []string
	sessions     map[string]*Session
}

func New(c *Config) *Client {
	return &Client{
		domain:       c.Domain,
		redirect:     c.Domain + "/auth",
		clientId:     c.ClientId,
		clientSecret: c.ClientSecret,
		roles:        c.Roles,
		bot:          newBot(c.BotToken),
		lock:         &sync.RWMutex{},
		sessions:     map[string]*Session{},
	}
}

func (c *Client) StartBot() {
	c.bot.connect()
}

func (c *Client) add(s *Session) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.sessions[s.session.State] = s
}

func (c *Client) get(s string) (*Session, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	session, ok := c.sessions[s]
	return session, ok
}

func (c *Client) pop(s string) (*Session, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	session, ok := c.sessions[s]
	if ok {
		delete(c.sessions, s)
	}
	return session, ok
}
