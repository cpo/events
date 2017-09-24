package rules

import (
	"github.com/cpo/events/interfaces"
)

// map with factory methods for producing bridges
var RuleFactories = map[string]func(map[string]interface{}) interfaces.Rule{
	// HUE bridge
	"regex": NewRegExRule,
}

func NewRegExRule(config map[string]interface{}) interfaces.Rule {
	re := RegExRule{}
	re.Initialize(config)
	return &re
}

