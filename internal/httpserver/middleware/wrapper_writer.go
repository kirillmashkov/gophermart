package middleware

import (
	"net/http"
)

type Writer struct {
    http.ResponseWriter
    code int
    bytes int
}

func (wr *Writer) Write(buf []byte) (n int, err error) {
	n, err = wr.ResponseWriter.Write(buf)
	wr.bytes += n
	return n, err
}

func (wr *Writer) WriteHeader(code int) {
	wr.code = code
	wr.ResponseWriter.WriteHeader(code)
}

func (wr *Writer) Status() int {
	return wr.code
}

func (wr *Writer) Bytes() int {
	return wr.bytes
}
