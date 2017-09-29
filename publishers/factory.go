package publishers

import (
	"github.com/cpo/events/interfaces"
)

// map with factory methods for producing bridges
var PublisherFactories = map[string]func(interfaces.EventManager, map[string]interface{}) interfaces.Publisher{
	// HUE bridge
	"mqtt":  NewMQTTPublisher,
}
