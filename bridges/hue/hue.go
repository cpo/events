package hue

import (
	"encoding/json"
	"github.com/cpo/events/interfaces"
	"github.com/cpo/go-hue/groups"
	"github.com/cpo/go-hue/lights"
	"github.com/cpo/go-hue/portal"
	"github.com/cpo/go-hue/sensors"
	"log"
	"os"
	"sync"
)

var logger = log.New(os.Stderr, "[HUEB] ", 1)

type HueBridge struct {
	id           string
	apiKey       string
	wg           sync.WaitGroup
	eventManager interfaces.EventManager
	pollInterval int64
}

func NewHueBridge() interfaces.Bridge {
	return new(HueBridge)
}

func (hue *HueBridge) Initialize(eventManager interfaces.EventManager, config map[string]interface{}) {
	cfgStr, _ := json.Marshal(config)
	hue.id = config["name"].(string)
	hue.apiKey = config["apiKey"].(string)
	hue.eventManager = eventManager
	hue.pollInterval = int64(config["pollInterval"].(float64))
	logger.Printf("Initialize HUE bridge %s with %s", hue.GetID(), cfgStr)
}

func (hue *HueBridge) GetID() string {
	return hue.id
}

func (hue *HueBridge) Connect() {
	logger.Printf("Connecting HUE bridge %s", hue.id)

	pp, err := portal.GetPortal()
	if err != nil {
		logger.Panic("portal.GetPortal() ERROR: ", err)
	}
	ll := lights.New(pp[0].InternalIPAddress, hue.apiKey)
	allLights, err := ll.GetAllLights()
	if err != nil {
		logger.Panic("lights.GetAllLights() ERROR: ", err)
	}
	logger.Println()
	logger.Println("Lights")
	logger.Println("------")
	for _, l := range allLights {
		logger.Printf("ID: %d Name: %s\n", l.ID, l.Name)
	}
	gg := groups.New(pp[0].InternalIPAddress, hue.apiKey)
	allGroups, err := gg.GetAllGroups()
	if err != nil {
		logger.Panic("groups.GetAllGroups() ERROR: ", err)
	}
	logger.Println()
	logger.Println("Groups")
	logger.Println("------")
	for _, g := range allGroups {
		logger.Printf("ID: %d Name: %s\n", g.ID, g.Name)
	}
	ss := sensors.New(pp[0].InternalIPAddress, hue.apiKey)
	allSensors, err := ss.GetAllSensors()
	if err != nil {
		logger.Panic("groups.GetAllSensors() ERROR: ", err)
	}
	logger.Println()
	logger.Println("Sensors")
	logger.Println("------")
	for _, g := range allSensors {
		logger.Printf("ID: %d Name: %s\n", g.ID, g.Name)
	}

	go hue.pollSensors(ss)

	hue.wg = sync.WaitGroup{}
	hue.wg.Add(1)
	hue.wg.Wait()

}

func (hue *HueBridge) Stop() {
	logger.Printf("Stop HUE bridge %s", hue.id)
}
func (hue *HueBridge) Trigger(uri string) {
	logger.Printf("Trigger bridge %s: %s", hue.id, uri)
}
