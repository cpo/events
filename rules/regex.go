package rules

import (
	"github.com/cpo/events/actions"
	"github.com/cpo/events/interfaces"
	"regexp"
)

type RegExRule struct {
	regex string
	actions []interfaces.Action
}

func (re *RegExRule) Initialize(config map[string]interface{}) {
	re.regex = config["regex"].(string)
	re.actions = actions.ParseActions(config["actions"].([]interface{}))
}

func (re *RegExRule) Matches(url string) bool {
	m, _ := regexp.Match(re.regex, []byte(url))
	return m
}

func (re *RegExRule) GetActions() []interfaces.Action {
	return re.actions
}
