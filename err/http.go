package err

import (
	"io"
	"net/http"
)

func Request(method, url string, body io.Reader) (*http.Request, Error) {
	r, e := http.NewRequest(method, url, body)
	return r, New(e)
}

func Get(url string) (*http.Response, Error) {
	r, e := http.Get(url)
	return r, New(e)
}

func Post(url, content string, body io.Reader) (*http.Response, Error) {
	r, e := http.Post(url, content, body)
	return r, New(e)
}

func Send(r *http.Request) (*http.Response, Error) {
	rs, e := http.DefaultClient.Do(r)
	return rs, New(e)
}
