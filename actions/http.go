package actions

import (
	"github.com/cpo/events/interfaces"
	"net/http"
)

type HttpAction struct {
	Method string
	Format string
}

func (ha *HttpAction) Initialize(config map[string]interface{}) *HttpAction {
	ha.Method = config["method"].(string)
	ha.Format = config["format"].(string)
	return ha
}

func (ha *HttpAction) Run(eventManager interfaces.EventManager, id string) {
	logger.Debugf(" [%s] action: HTTP %s to %s", id, ha.Method, ha.Format)
	req,_:=http.NewRequest(ha.Method, ha.Format, nil)
	response,err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Panicf("Error %s", err)
	}
	logger.Infof("Request ended with status %d: %s", response.StatusCode, response.Status)
}
