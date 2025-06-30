package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/kirillmashkov/gophermart/internal/app"
	"go.uber.org/zap"
)

type compressWriter struct {
    w  http.ResponseWriter
    zw *gzip.Writer
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(payload []byte) (int, error) {
	return c.zw.Write(payload)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w: w,
		zw: gzip.NewWriter(w),
	}
}

type compressReader struct {
	r io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error){
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r: r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	err := c.r.Close()
	if err != nil {
		return err
	}
	return c.zr.Close()
}

func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resultWriter := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")
		if supportGzip {
			compressWriter := newCompressWriter(w)
			resultWriter = compressWriter
			defer func() {
				if errClose := compressWriter.Close(); errClose != nil {
					app.Log.Error("Can't close writer when compress", zap.Error(errClose))
				}
			} ()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		needDecompress := strings.Contains(contentEncoding, "gzip")
		if needDecompress {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func() {
				if errClose := cr.Close(); errClose != nil {
					app.Log.Error("Can't read request when decompress", zap.Error(errClose))
				}
			}()
		}

		next.ServeHTTP(resultWriter, r)
	})
}