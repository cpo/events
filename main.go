package main

import (
	"github.com/cpo/events/manager"
	logger "github.com/Sirupsen/logrus"
	"flag"
)

func main() {

	logLevel := flag.String("loglevel", "debug", "Set loglevel (debug|info|warn|error)")
	flag.Parse()

	formatter := new(logger.TextFormatter)
	formatter.ForceColors=true
	logger.SetFormatter(formatter)
	level,_:=logger.ParseLevel(*logLevel)
	logger.SetLevel(level)

	var evtMgr = manager.New()

	logger.Info("Startup")

	evtMgr.Start()
}
