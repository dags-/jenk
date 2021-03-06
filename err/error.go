package err

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
)

type Err interface {
	Error() string
	UnWrap() error
	Present() bool
	Warn()
	Panic()
}

type Error struct {
	e   error
	msg string
	ctx []string
}

var none = &Error{
	e:   nil,
	msg: "",
	ctx: []string{},
}

func New(e error) Error {
	if e == nil {
		return Nil()
	}
	if er, ok := e.(*Error); ok {
		return Error{ctx: er.ctx, e: er.e}
	}
	return Error{ctx: stack(), e: e}
}

func Nil() Error {
	return *none
}

func (e *Error) Error() string {
	er := e.unwrap()
	buf := bytes.Buffer{}
	buf.WriteString(er.UnWrap().Error())
	for _, s := range er.ctx {
		buf.WriteRune('\n')
		buf.WriteString(s)
	}
	return buf.String()
}

func (e *Error) UnWrap() error {
	if t, ok := e.e.(*Error); ok {
		return t.UnWrap()
	}
	return e.e
}

func (e *Error) Present() bool {
	return e.e != nil
}

func (e *Error) Warn() {
	if e.Present() {
		log.Println(e.Error())
	}
}

func (e *Error) Panic() {
	if e.Present() {
		panic(e.Error())
	}
}

func (e Error) Log() {
	e.Warn()
}

func (e Error) Fatal() {
	e.Panic()
}

func (e *Error) unwrap() *Error {
	er := e
	for true {
		t, ok := er.e.(*Error)
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
