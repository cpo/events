package main

import (
	"github.com/cpo/events/manager"
	"log"
	"os"
)

var logger = log.New(os.Stderr, "[main] ", 1)

func main() {
	var evtMgr = manager.New()

	logger.Printf("Startup")

	evtMgr.Start()
}
