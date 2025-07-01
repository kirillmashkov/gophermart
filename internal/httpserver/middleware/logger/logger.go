package logger

import (
	"net/http"
	"time"
	"github.com/kirillmashkov/gophermart/internal/app"
    "github.com/kirillmashkov/gophermart/internal/httpserver/middleware"
	"go.uber.org/zap"
)


type ResponseWrapper interface {
	http.ResponseWriter
	Status() int
	Bytes() int
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		
		app.Log.Info("request",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
        )

		ww := wrap(w)

        next.ServeHTTP(ww, r)		

		app.Log.Info("response",
			zap.Duration("elapsed", time.Since(t1)),
            zap.Int("status", ww.Status()),
            zap.Int("content length", ww.Bytes()))
	})
}	

func wrap(w http.ResponseWriter) ResponseWrapper {
    return &middleware.Writer{ResponseWriter: w}
}

