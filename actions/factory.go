package actions

import (
	"github.com/cpo/events/interfaces"
	log "github.com/Sirupsen/logrus"

)

// map with factory methods for producing bridges
var ActionFactories = map[string]func(map[string]interface{}) interfaces.Action{
	// HUE bridge
	"wait": NewWaitAction,
	"trigger": NewTriggerAction,
	"http": NewHttpAction,
	"email": NewEmailAction,
}

var logger = log.New()

func NewWaitAction(config map[string]interface{}) interfaces.Action {
	return new(WaitAction).Initialize(config)
}

func NewEmailAction(config map[string]interface{}) interfaces.Action {
	return new(EMailAction).Initialize(config)
}

func NewTriggerAction(config map[string]interface{}) interfaces.Action {
	return new(TriggerAction).Initialize(config)
}

func NewHttpAction(config map[string]interface{}) interfaces.Action {
	return new(HttpAction).Initialize(config)
}

func ParseActions(config []interface{}) []interfaces.Action {
	actions := make([]interfaces.Action,0)
	logger.Debugf("Parse actions")
	for n,aConfig := range config {
		config := aConfig.(map[string]interface{})
		actionType := config["type"].(string)
		logger.Debugf(" -> %d: Action type %s", n, actionType)
		if factory,found := ActionFactories[actionType]; found {
			actions = append(actions, factory(config))
		} else {
			logger.Debugf("** ERROR type %s not found **", )
		}


	}
	return actions
}

