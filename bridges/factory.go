package bridges

import (
	"github.com/cpo/events/bridges/hue"
	"github.com/cpo/events/bridges/mqtt"
	"github.com/cpo/events/interfaces"
)

// map with factory methods for producing bridges
var BridgeFactories = map[string]func() interfaces.Bridge{
	// HUE bridge
	"hue":  hue.NewHueBridge,
	"mqtt": mqtt.NewMQTTBridge,
}
