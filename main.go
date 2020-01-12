package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/dags-/jenk/discord"
	"github.com/dags-/jenk/err"
	"github.com/dags-/jenk/jenkins"
	"github.com/dags-/jenk/manager"
)

type Config struct {
	Port                int      `json:"port"`
	Domain              string   `json:"domain"`
	JenkinsUser         string   `json:"jenkins_user"`
	JenkinsToken        string   `json:"jenkins_token"`
	JenkinsServer       string   `json:"jenkins_server"`
	DiscordBotToken     string   `json:"discord_bot_token"`
	DiscordClientId     string   `json:"discord_client_id"`
	DiscordClientSecret string   `json:"discord_client_secret"`
	DiscordRoles        []string `json:"discord_roles"`
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
		Roles:        c.DiscordRoles,
	})

	d.StartBot()

	m := manager.New(j, d)
	mux := http.NewServeMux()
	mux.HandleFunc("/", m.ServeDir("assets"))
	mux.HandleFunc("/auth", d.AuthHandler)
	mux.Handle("/file/", http.StripPrefix("/file/", http.HandlerFunc(m.ServeFile)))
	mux.Handle("/data/", http.StripPrefix("/data/", http.HandlerFunc(m.ServeData)))
	fmt.Println("starting server at", l.Addr().String())

	go serve(l, mux)

	exit()
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

func serve(l net.Listener, h http.Handler) {
	fmt.Println("starting server at", l.Addr().String())
	err.New(http.Serve(l, h)).Fatal()
}

func exit() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)
		line = strings.ToLower(line)
		if line == "exit" || line == "stop" {
			os.Exit(0)
		}
	}
}

func listen(port int) net.Listener {
	l, e := net.Listen("tcp", fmt.Sprint("127.0.0.1:", port))
	if e != nil {
		panic(e)
	}
	return l
}
