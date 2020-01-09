package err

import (
	"encoding/json"
	"io"
)

func Decode(r io.Reader, i interface{}) Error {
	return New(json.NewDecoder(r).Decode(i))
}

func Encode(w io.Writer, i interface{}) Error {
	return New(json.NewEncoder(w).Encode(i))
}

func EncodePretty(w io.Writer, i interface{}) Error {
	en := json.NewEncoder(w)
	en.SetIndent("", "  ")
	return New(en.Encode(i))
}
