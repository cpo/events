package zwave

import (
	"encoding/json"
	"github.com/cpo/events/interfaces"
	"sync"
	"time"
	"github.com/stampzilla/gozwave"
	"github.com/stampzilla/gozwave/events"
	log "github.com/Sirupsen/logrus"
	"fmt"
)

var logger = log.New()

type ZWaveBridge struct {
	id           string
	port         string
	config       map[string]interface{}
	wg           sync.WaitGroup
	eventManager interfaces.EventManager
	controller   *gozwave.Controller
}

func NewZWaveBridge() interfaces.Bridge {
	return new(ZWaveBridge)
}

func (zw *ZWaveBridge) Initialize(eventManager interfaces.EventManager, config map[string]interface{}) {
	cfgStr, _ := json.Marshal(config)
	logger.Info("Initialize Z-Wave bridge %s with %s", zw.GetID(), cfgStr)
	zw.port = config["port"].(string)
	zw.id = config["name"].(string)
	zw.config = config
	zw.eventManager = eventManager
}

func (zw *ZWaveBridge) GetID() string {
	return zw.id
}

func (zw *ZWaveBridge) Connect() {
	defer zw.connectRecovery()

	var err error
	zw.controller, err = gozwave.Connect(zw.port, "")
	if err != nil {
		logger.Panicf("Panic: %s", err)
	}

	logger.Debugf("Z-Wave bridge %s connected.", zw.id)

	go func() {
		for {
			select {
			case event := <-zw.controller.GetNextEvent():
				logger.Println("----------------------------------------")
				logger.Debugf("Event: %#v\n", event)
				switch e := event.(type) {
				case events.NodeDiscoverd:
					znode := zw.controller.Nodes.Get(e.Address)
					znode.RLock()
					url := fmt.Sprintf("zwave://%s/node/%d", zw.id, znode.Id)
					logger.Debugf("Node detected: %s", url)
					znode.RUnlock()

				case events.NodeUpdated:
					znode := zw.controller.Nodes.Get(e.Address)
					znode.RLock()
					//logger.Debugf("Node: %#v\n", znode)

					params := ""
					for k, v := range znode.StateBool {
						if params != "" {
							params += "&"
						}
						params += fmt.Sprintf("%s=%t", k, v)
					}
					if (params != "") {
						params = "#" + params
					}
					evt := fmt.Sprintf("zwave://%s/node/%d/state%s", zw.id, znode.Id, params)
					zw.eventManager.Dispatch(evt)
					znode.RUnlock()
				}
			}
		}
	}()

	zw.wg = sync.WaitGroup{}
	zw.wg.Add(1)
	zw.wg.Wait()

	logger.Debugf("Stop Z-Wave bridge %s", zw.id)
}

func (zw *ZWaveBridge) connectRecovery() {
	if r := recover(); r != nil {
		logger.Debugf("Recovering connection for Z-Wave bridge %s", zw.id)
		time.Sleep(10 * time.Second)
		zw.Connect()
	}
}

func (zw *ZWaveBridge) Stop() {
	logger.Debugf("Setting stop signal for Z-Wave bridge %s", zw.id)
	zw.wg.Done()
}

func (zw *ZWaveBridge) Trigger(uri string) {
	logger.Debugf("(UNIMPLEMENTED) Publishing Z-Wave bridge %s: %s", zw.id, uri)
}
