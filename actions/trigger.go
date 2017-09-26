package actions

import (
	"github.com/cpo/events/interfaces"
)

type TriggerAction struct {
	URL string
}

func (ta *TriggerAction) Initialize(config map[string]interface{}) *TriggerAction {
	ta.URL = config["trigger"].(string)
	return ta
}

func (ta *TriggerAction) Run(eventManager interfaces.EventManager, id string) {
	logger.Debugf(" [%s] action: trigger URL %s", id, ta.URL)
	eventManager.Trigger(ta.URL)
}
