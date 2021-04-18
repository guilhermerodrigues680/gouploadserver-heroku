package handler

import (
	"net"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// LoggingInterceptorOnServer é um interceptor que tem acesso a requisicao e resposta
// antes e depois da chamada do Handle
type LoggingInterceptorOnServer struct {
	next   http.Handler
	logger *logrus.Entry
}

func NewLoggingInterceptorOnServer(next http.Handler, logger *logrus.Entry) *LoggingInterceptorOnServer {
	return &LoggingInterceptorOnServer{
		next:   next,
		logger: logger,
	}
}

func (l *LoggingInterceptorOnServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	remoteIp, _, _ := net.SplitHostPort(r.RemoteAddr)
	forwardedFor := r.Header.Get("X-Forwarded-For")
	lrw := newLoggingResponseWriter(w)
	l.next.ServeHTTP(lrw, r)
	l.logger.Infof("%s - %s '%s %s' %d %s", forwardedFor, remoteIp, r.Method, r.RequestURI, lrw.StatusCode, time.Since(start))
}

// loggingResponseWriter é um ResponseWriter para fazer o log do código HTTP enviado ao cliente
type loggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// WriteHeader(int) não é chamado se nossa resposta retornar implicitamente 200 OK, então
	// configura-se por default este status code.
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.StatusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

// LoggingInterceptorOnFunc é uma objeto capaz de interceptar 'httprouter.Handle'
type LoggingInterceptorOnFunc struct {
	logger *logrus.Entry
}

func NewLoggingInterceptorOnFunc(logger *logrus.Entry) *LoggingInterceptorOnFunc {
	return &LoggingInterceptorOnFunc{
		logger: logger,
	}
}

func (l *LoggingInterceptorOnFunc) log(next func(w http.ResponseWriter, r *http.Request, ps httprouter.Params)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		l.logger.Info(r.URL.Path)
		next(w, r, ps)
	}
}
