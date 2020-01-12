package discord

import (
	"context"
	"fmt"
	"sync"

	"github.com/andersfylling/disgord"
)

type Bot struct {
	token   string
	lock    *sync.RWMutex
	guild   disgord.Snowflake
	session disgord.Session
	roles   map[string]*disgord.Role
}

func newBot(token string) *Bot {
	return &Bot{
		token: token,
		lock:  &sync.RWMutex{},
	}
}

func (b *Bot) connect() {
	b.lock.Lock()

	s, e := disgord.NewClient(disgord.Config{
		BotToken: b.token,
	})

	if e != nil {
		panic(e)
	}

	b.session = s

	b.session.On(disgord.EvtGuildCreate, func(s disgord.Session, e *disgord.GuildCreate) {
		fmt.Println("connected to guild: ", e.Guild.Name)
		b.guild = e.Guild.ID
		b.roles = buildRoleMap(e.Guild.Roles)
		b.lock.Unlock()
	})

	b.session.On(disgord.EvtGuildRoleCreate, func(s disgord.Session, e *disgord.GuildRoleCreate) {
		fmt.Println("created role: ", e.Role.Name)
		b.refreshRoles()
	})

	b.session.On(disgord.EvtGuildRoleUpdate, func(s disgord.Session, e *disgord.GuildRoleUpdate) {
		fmt.Println("updated role: ", e.Role.Name)
		b.refreshRoles()
	})

	b.session.On(disgord.EvtGuildRoleDelete, func(s disgord.Session, e *disgord.GuildRoleCreate) {
		fmt.Println("deleted role: ", e.Role.Name)
		b.refreshRoles()
	})

	e = s.Connect(context.Background())
	if e != nil {
		panic(e)
	}

	fmt.Println("bot session started")
}

func (b *Bot) refreshRoles() {
	roles, e := b.session.GetGuildRoles(context.Background(), b.guild)
	if e != nil {
		fmt.Println(e)
		return
	}
	backing := buildRoleMap(roles)
	b.lock.Lock()
	defer b.lock.Unlock()
	b.roles = backing
}

func (b *Bot) hasAnyRole(id string, roles []string) bool {
	if len(roles) == 0 {
		return true
	}

	m, e := b.session.GetMember(context.Background(), b.guild, disgord.ParseSnowflakeString(id))
	if e != nil {
		return false
	}

	b.lock.Lock()
	guildRoles := b.roles
	b.lock.Unlock()

	for _, name := range roles {
		if guildRole, ok := guildRoles[name]; ok {
			for _, memberRole := range m.Roles {
				if guildRole.ID == memberRole {
					return true
				}
			}
		}
	}
	return false
}

func buildRoleMap(roles []*disgord.Role) map[string]*disgord.Role {
	result := map[string]*disgord.Role{}
	for _, r := range roles {
		result[r.Name] = r
	}
	return result
}
