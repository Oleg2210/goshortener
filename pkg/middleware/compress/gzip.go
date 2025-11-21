package compres

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipReader struct {
	body io.ReadCloser
	zr   *gzip.Reader
}

func newGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{body: r, zr: zr}, nil
}

func (gr *gzipReader) Read(p []byte) (int, error) {
	return gr.zr.Read(p)
}

func (gr *gzipReader) Close() error {
	_ = gr.zr.Close()
	return gr.body.Close()
}

type gzipWriter struct {
	http.ResponseWriter
	gz      *gzip.Writer
	enabled bool
}
.
func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
	}
}

func (gw *gzipWriter) startGzip() {
	gw.gz = gzip.NewWriter(gw.ResponseWriter)
	gw.enabled = true
	gw.Header().Set("Content-Encoding", "gzip")
	gw.Header().Set("Vary", "Accept-Encoding")
}
.
func (gw *gzipWriter) Write(p []byte) (int, error) {
	if !gw.enabled {
		return gw.ResponseWriter.Write(p)
	}
	return gw.gz.Write(p)
}

func (gw *gzipWriter) Close() error {
	if gw.gz != nil {
		return gw.gz.Close()
	}
	return nil
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := newGzipReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			r.Body = gr
			defer gr.Close()
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		ct := w.Header().Get("Content-Type")
		if strings.Contains(ct, "application/json") || strings.Contains(ct, "text/html") {
			w := NewGzipWriter(w)
			w.startGzip()
			defer w.Close()
		}

		next.ServeHTTP(w, r)
	})
}
