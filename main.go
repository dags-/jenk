package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/dags-/jenk/discord"
	"github.com/dags-/jenk/err"
	"github.com/dags-/jenk/jenkins"
	"github.com/dags-/jenk/manager"
)

type Config struct {
	Port                int                 `json:"port"`
	Domain              string              `json:"domain"`
	JenkinsUser         string              `json:"jenkins_user"`
	JenkinsToken        string              `json:"jenkins_token"`
	JenkinsServer       string              `json:"jenkins_server"`
	DiscordBotToken     string              `json:"discord_bot_token"`
	DiscordClientId     string              `json:"discord_client_id"`
	DiscordClientSecret string              `json:"discord_client_secret"`
	Permissions         map[string][]string `json:"permissions"`
}

func main() {
	c := loadConfig()
	l := listen(c.Port)

	j := jenkins.NewClient(&jenkins.Config{
		Server: c.JenkinsServer,
		User:   c.JenkinsUser,
		Token:  c.JenkinsToken,
	})

	d := discord.New(&discord.Config{
		Domain:       c.Domain,
		BotToken:     c.DiscordBotToken,
		ClientId:     c.DiscordClientId,
		ClientSecret: c.DiscordClientSecret,
		Roles:        c.Permissions,
	})

	d.StartBot()

	m := manager.New(j, d)
	mux := http.NewServeMux()
	mux.HandleFunc("/", m.ServeDir("assets"))
	mux.HandleFunc("/auth", d.AuthHandler)
	mux.Handle("/file/", http.StripPrefix("/file/", http.HandlerFunc(m.ServeFile)))
	mux.Handle("/data/", http.StripPrefix("/data/", http.HandlerFunc(m.ServeData)))
	fmt.Println("starting server at", l.Addr().String())

	err.New(http.Serve(l, mux)).Fatal()
}

func loadConfig() *Config {
	var c Config
	d, e0 := ioutil.ReadFile("config.json")
	if e0 != nil {
		d, e1 := json.MarshalIndent(c, "", "  ")
		if e1 != nil {
			panic(e1)
		}

		e3 := ioutil.WriteFile("config.json", d, os.ModePerm)
		if e3 != nil {
			panic(e3)
		}

		panic(e0)
	}

	e4 := json.Unmarshal(d, &c)
	if e4 != nil {
		panic(e4)
	}
	return &c
}

func listen(port int) net.Listener {
	l, e := net.Listen("tcp", fmt.Sprint("127.0.0.1:", port))
	if e != nil {
		panic(e)
	}
	return l
}
