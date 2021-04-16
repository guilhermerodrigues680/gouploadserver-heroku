package transport

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

type Server struct {
	r      *httprouter.Router
	logger *logrus.Entry
}

func NewServer(staticDirPath string, logger *logrus.Entry) *Server {
	router := httprouter.New()

	// router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	//http.ServeFile(w, r, "static/404.html")
	// 	logger.Info("oaoaoao")
	// })

	logReq := NewLoggingInterceptorOnFunc(logger.WithField("server", "interceptor-on-func"))

	fileServer := http.FileServer(http.Dir(staticDirPath))
	router.GET("/*filepath", logReq.log(func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		req.URL.Path = p.ByName("filepath")
		fileServer.ServeHTTP(w, req)
	}))

	return &Server{
		r:      router,
		logger: logger,
	}
}

func (f *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mw := NewLoggingInterceptorOnServer(f.r, f.logger.WithField("server", "interceptor-on-server"))
	mw.ServeHTTP(w, r)
}
