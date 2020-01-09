package err

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
)

type Error interface {
	Error() string
	UnWrap() error
	Present() bool
	Warn()
	Panic()
}

type err struct {
	e   error
	msg string
	ctx []string
}

var none = &err{
	e:   nil,
	msg: "",
	ctx: []string{},
}

func New(e error) Error {
	if e == nil {
		return Nil()
	}
	if er, ok := e.(*err); ok {
		return &err{ctx: er.ctx, e: er.e}
	}
	return &err{ctx: stack(), e: e}
}

func Nil() Error {
	return none
}

func (e *err) Error() string {
	er := e.unwrap()
	buf := bytes.Buffer{}
	buf.WriteString(er.UnWrap().Error())
	for _, s := range er.ctx {
		buf.WriteRune('\n')
		buf.WriteString(s)
	}
	return buf.String()
}

func (e *err) UnWrap() error {
	if t, ok := e.e.(Error); ok {
		return t.UnWrap()
	}
	return e.e
}

func (e *err) Present() bool {
	return e.e != nil
}

func (e *err) Warn() {
	if e.Present() {
		log.Println(e.Error())
	}
}

func (e *err) Panic() {
	if e.Present() {
		panic(e.Error())
	}
}

func (e *err) unwrap() *err {
	er := e
	for true {
		t, ok := er.e.(*err)
		if ok {
			er = t
		} else {
			break
		}
	}
	return er
}

func stack() []string {
	var ctx []string
	for i := 2; i < 10; i++ {
		_, fn, line, ok := runtime.Caller(i)
		if ok {
			ctx = append(ctx, fmt.Sprint(fn, ":", line))
		} else {
			break
		}
	}
	return ctx
}
