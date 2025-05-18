package middleware

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "cannot decode gzip body", http.StatusBadRequest)
				return
			}
			defer func(gr *gzip.Reader) {
				err := gr.Close()
				if err != nil {
					log.Printf("failed to close gzip reader: %v", err)
				}
			}(gr)
			r.Body = io.NopCloser(gr)
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gw := gzip.NewWriter(w)
		defer func(gw *gzip.Writer) {
			err := gw.Close()
			if err != nil {
				log.Printf("failed to close gzip writer: %v", err)
			}
		}(gw)

		grw := &gzipResponseWriter{
			ResponseWriter: w,
			Writer:         gw,
		}

		next.ServeHTTP(grw, r)
	})
}
