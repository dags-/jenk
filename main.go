package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/dags-/jenk/err"
	"github.com/dags-/jenk/jenkins"
	"github.com/dags-/jenk/manager"
)

var (
	port    = flag.Int("port", 8123, "Server port")
	server  = flag.String("server", "", "Jenkins address")
	project = flag.String("project", "", "Jenkins project name")
	user    = flag.String("user", "", "Jenkins API user")
	token   = flag.String("token", "", "Jenkins API token")
)

func init() {
	flag.Parse()
}

func main() {
	l := listen(*port)
	c := jenkins.NewClient(*server, *user, *token)
	m := manager.New(c, "http://"+l.Addr().String(), *project)
	fmt.Println(m.GetAddress())
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("assets")))
	mux.Handle("/file/", http.StripPrefix("/file/", http.HandlerFunc(m.ServeFile)))
	mux.Handle("/data/", http.StripPrefix("/data/", http.HandlerFunc(m.ServeData)))
	err.New(http.Serve(l, mux)).Panic()
}

func listen(port int) net.Listener {
	l, e := net.Listen("tcp", fmt.Sprint("127.0.0.1:", port))
	if e != nil {
		panic(e)
	}
	return l
}
