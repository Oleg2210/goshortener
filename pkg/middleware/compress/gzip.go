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
	gz        *gzip.Writer
	enabled   bool
	wroteHdr  bool
	minLength int
	buf       []byte
}

func newGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
		minLength:      140,
		buf:            make([]byte, 0, 1024),
	}
}

func (gw *gzipWriter) WriteHeader(code int) {
	if code < 200 || code == 204 || code == 304 {
		gw.enabled = false
	}

	gw.wroteHdr = true

	if gw.enabled {
		gw.Header().Set("Content-Encoding", "gzip")
		gw.Header().Del("Content-Length")
	}

	gw.ResponseWriter.WriteHeader(code)
}

func (gw *gzipWriter) Write(p []byte) (int, error) {
	if !gw.wroteHdr {
		gw.buf = append(gw.buf, p...)
		return len(p), nil
	}

	if !gw.enabled {
		return gw.ResponseWriter.Write(p)
	}

	return gw.gz.Write(p)
}

func (gw *gzipWriter) startGzip() {
	gw.gz = gzip.NewWriter(gw.ResponseWriter)
	gw.enabled = true
}

func (gw *gzipWriter) Close() error {
	if !gw.wroteHdr {
		return nil
	}
	if gw.enabled {
		return gw.gz.Close()
	}
	return nil
}

func (gw *gzipWriter) Flush() {
	if f, ok := gw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
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

		clientSupportsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if !clientSupportsGzip {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)

		gw := newGzipWriter(w)
		defer gw.Close()

		ct := w.Header().Get("Content-Type")
		if strings.Contains(ct, "application/json") || strings.Contains(ct, "text/html") {
			gw.startGzip()
			_, _ = gw.gz.Write(gw.buf)
		}
	})
}
