package err

import "io"

func Close(c io.Closer) {
	if c != nil {
		e := New(c.Close())
		e.Warn()
	}
}
