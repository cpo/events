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
	config       map[string]interface{}
	wg           sync.WaitGroup
	eventManager interfaces.EventManager
	controller   *gozwave.Controller
}

func NewZWaveBridge() interfaces.Bridge {
	return new(ZWaveBridge)
}

func (mq *ZWaveBridge) Initialize(eventManager interfaces.EventManager, config map[string]interface{}) {
	cfgStr, _ := json.Marshal(config)
	mq.id = config["name"].(string)
	mq.config = config
	mq.eventManager = eventManager
	logger.Info("Initialize Z-Wave bridge %s with %s", mq.GetID(), cfgStr)
	var err error
	mq.controller, err = gozwave.Connect(config["port"].(string), "")
	if err != nil {
		logger.Panicf("Panic: %s", err)
	}
}

func (mq *ZWaveBridge) GetID() string {
	return mq.id
}

func (mq *ZWaveBridge) Connect() {
	logger.Debugf("Z-Wave bridge %s connected.", mq.id)

	go func() {
		for {
			select {
			case event := <-mq.controller.GetNextEvent():
				logger.Println("----------------------------------------")
				logger.Debugf("Event: %#v\n", event)
				switch e := event.(type) {
				case events.NodeDiscoverd:
					znode := mq.controller.Nodes.Get(e.Address)
					znode.RLock()
					url := fmt.Sprintf("zwave://%s/node/%d", mq.id, znode.Id)
					logger.Debugf("Node detected: %s", url)
					znode.RUnlock()

				case events.NodeUpdated:
					znode := mq.controller.Nodes.Get(e.Address)
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
					evt := fmt.Sprintf("zwave://%s/node/%d/state%s", mq.id, znode.Id, params)
					mq.eventManager.Dispatch(evt)
					znode.RUnlock()
				}
			}
		}
	}()

	mq.wg = sync.WaitGroup{}
	mq.wg.Add(1)
	mq.wg.Wait()

	logger.Debugf("Stop Z-Wave bridge %s", mq.id)
}

func (mq *ZWaveBridge) connectRecovery() {
	if r := recover(); r != nil {
		logger.Debugf("Recovering connection for Z-Wave bridge %s", mq.id)
	}
	time.Sleep(3 * time.Second)
	mq.Connect()
}

func (mq *ZWaveBridge) Stop() {
	logger.Debugf("Setting stop signal for Z-Wave bridge %s", mq.id)
	mq.wg.Done()
}

func (mq *ZWaveBridge) Trigger(uri string) {
	logger.Debugf("(UNIMPLEMENTED) Publishing Z-Wave bridge %s: %s", mq.id, uri)
}
