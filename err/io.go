package err

import "io"

func Close(c io.Closer) {
	if c != nil {
		New(c.Close()).Warn()
	}
}
