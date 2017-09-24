package actions

import (
	"github.com/cpo/events/interfaces"
	"time"
)

type WaitAction struct {
	Seconds int
}

func (wa *WaitAction) Initialize(config map[string]interface{}) *WaitAction {
	wa.Seconds = int(config["seconds"].(float64))
	return wa
}

func (wa *WaitAction) Run(eventManager interfaces.EventManager, id string) {
	logger.Printf(" [%s] action: Wait for %d seconds", id, wa.Seconds)
	time.Sleep(time.Duration(wa.Seconds) * time.Second)
}
