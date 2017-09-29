package publishers

import (
	"github.com/cpo/events/bridges/zwave"
	"github.com/cpo/events/bridges/hue"
	"github.com/cpo/events/bridges/mqtt"
	"github.com/cpo/events/interfaces"
)

// map with factory methods for producing bridges
var PublisherFactories = map[string]func() *interfaces.Publisher{
	// HUE bridge
	"mqtt":  NewMQTTPublisher,
}
