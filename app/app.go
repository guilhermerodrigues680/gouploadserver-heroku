package app

import (
	"gouploadserver/handler"
	"net/http"

	"github.com/sirupsen/logrus"
)

func Run(cwd string, port string, logger *logrus.Entry) error {
	logger.Info("** Go Upload Server **")
	logger.Infof("Dir: %s, Port: %s", cwd, port)

	h := handler.NewServer(cwd, logger.WithField("server", "handler"))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	logger.Infof("Listening on: %s", srv.Addr)
	err := srv.ListenAndServe()
	if err != nil {
		logger.WithError(err).Error("Server error")
		return err
	}

	return nil
}
