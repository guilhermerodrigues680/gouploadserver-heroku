package main

/*
	gouploadserver [options] [path]
	ex: gouploadserver -p 8081 .
	options:
		-p or --port Port to use (defaults to 8080)

*/

import (
	"flag"
	"fmt"
	"gouploadserver/transport"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type cmdFlag struct {
	value        string
	name         string
	shortName    string
	defaultValue string
	usage        string
}

func NewCmdFlag(name, shortName, defaultValue, usage string) *cmdFlag {
	cf := cmdFlag{
		value:        "",
		name:         name,
		shortName:    shortName,
		defaultValue: defaultValue,
		usage:        usage,
	}
	flag.StringVar(&cf.value, name, defaultValue, usage)
	flag.StringVar(&cf.value, shortName, defaultValue, usage+" (shorthand)")
	return &cf
}

var portFlag *cmdFlag = NewCmdFlag("port", "p", "8000", "Port to use")
var pathArg string

func init() {
	flag.Usage = func() {
		fmt.Println("------------------------")
		defer fmt.Println("------------------------")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		// fmt.Fprintf(flag.CommandLine.Output(), "Custom help %s:\n", os.Args[0])
		// flag.VisitAll(func(f *flag.Flag) {
		// 	fmt.Fprintf(flag.CommandLine.Output(), "    %v\n", f.Usage) // f.Name, f.Value
		// })
	}
}

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
	go func() {
		for {
			PrintMemUsage()
			time.Sleep(time.Second)
		}
	}()

	logger := getLogger()
	logger.Debug(strings.Join(os.Args, " "))

	flag.Parse()
	pathArg = flag.Arg(0)
	if pathArg == "" {
		logger.Error("Path Arg Not Found")
		flag.Usage()
		os.Exit(1)
	}

	//
	logger.Info("** Go Upload Server **")
	logger.Infof("Path: %s, Port: %s", pathArg, portFlag.value)

	handler := transport.NewServer(pathArg, logger.WithField("server", "handler"))

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

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tHeapAlloc = %v MiB", bToMb(m.HeapAlloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
	// fmt.Printf("Alloc = %v B", m.Alloc)
	// fmt.Printf("\tTotalAlloc = %v B", m.TotalAlloc)
	// fmt.Printf("\tSys = %v B", m.Sys)
	// fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
