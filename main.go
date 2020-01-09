package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/dags-/jenk/err"
	"github.com/dags-/jenk/jenkins"
	"github.com/dags-/jenk/manager"
)

var (
	port   = flag.Int("port", 8123, "Server port")
	server = flag.String("server", "", "Jenkins address")
	user   = flag.String("user", "", "Jenkins API user")
	token  = flag.String("token", "", "Jenkins API token")
)

func init() {
	flag.Parse()

	str := func(s string) *string {
		return &s
	}

	server = str("https://ci.dags.me")
	user = str("dags")
	token = str("118b058ab56c8e1d1726e02c5b909d9c30")
}

func main() {
	fmt.Println("starting...")
	l := listen(*port)
	c := jenkins.NewClient(*server, *user, *token)
	m := manager.New(c)
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(fileHandler("assets")))
	mux.Handle("/file/", http.StripPrefix("/file/", http.HandlerFunc(m.ServeFile)))
	mux.Handle("/data/", http.StripPrefix("/data/", http.HandlerFunc(m.ServeData)))
	err.New(http.Serve(l, mux)).Panic()
}

func fileHandler(dir http.Dir) func(http.ResponseWriter, *http.Request) {
	handler := http.FileServer(dir)
	return func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 {
			// disallow sub path
			if strings.LastIndex(r.URL.Path, "/") > 0 {
				http.NotFound(w, r)
				return
			}
			// if not a file serve root
			if !strings.ContainsRune(r.URL.Path, '.') {
				r.URL.Path = ""
			}
		}
		handler.ServeHTTP(w, r)
	}
}

func listen(port int) net.Listener {
	l, e := net.Listen("tcp", fmt.Sprint("127.0.0.1:", port))
	if e != nil {
		panic(e)
	}
	return l
}
