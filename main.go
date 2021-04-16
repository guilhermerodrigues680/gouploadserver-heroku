package main

import (
	"gouploadserver/transport"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func getLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
	})
	logger.SetLevel(logrus.TraceLevel) // log all
	logger.SetOutput(os.Stdout)        // Output to stdout instead of the default stderr
	return logger
}

func main() {
	logger := getLogger()
	logger.Info("** Go Upload Server **")

	staticDirPath := "."

	handler := transport.NewServer(staticDirPath, logger.WithField("server", "handler"))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	logger.Infof("Listening on: %s", srv.Addr)
	err := srv.ListenAndServe()
	if err != nil {
		logger.WithError(err).Fatal("Server error")
	}
}
