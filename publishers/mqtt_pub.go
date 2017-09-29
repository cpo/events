package publishers

import (
	"encoding/json"
	"github.com/yosssi/gmq/mqtt/client"
	logger "github.com/Sirupsen/logrus"
	"fmt"
	"github.com/cpo/events/interfaces"
	"strings"
	"sync"
	"time"
)

type MQTTPublisher struct {
	id           string
	mqttClient   *client.Client
	config       map[string]interface{}
	wg           sync.WaitGroup
	prefix       string
	ready        bool
	eventManager interfaces.EventManager
}

func NewMQTTPublisher(manager interfaces.EventManager, config map[string]interface{}) interfaces.Publisher {
	p := new(MQTTPublisher).Initialize(manager, config)
	return p
}

func (mqp *MQTTPublisher) Initialize(eventManager interfaces.EventManager, config map[string]interface{}) interfaces.Publisher {
	cfgStr, _ := json.Marshal(config)
	mqp.id = config["name"].(string)
	mqp.config = config
	mqp.prefix = config["prefix"].(string)
	mqp.ready = false
	mqp.eventManager = eventManager
	logger.Debugf("Initialize MQTT publisher %s with %s", mqp.GetID(), cfgStr)
	return mqp
}

func (mqp *MQTTPublisher) GetID() string {
	return mqp.id
}

func (mqp *MQTTPublisher) Connect() {
	logger.Infof("Connecting MQTT publisher %s", mqp.id)
	mqp.mqttClient = client.New(&client.Options{})
	defer mqp.mqttClient.Terminate()
	defer mqp.connectRecovery()

	logger.Infof("MQTT connecting client %s to %s://%s:%.0f",
		mqp.config["clientId"].(string),
		mqp.config["proto"].(string),
		mqp.config["host"].(string),
		mqp.config["port"].(float64))

	err := mqp.mqttClient.Connect(&client.ConnectOptions{
		Network:  mqp.config["proto"].(string),
		Address:  mqp.config["host"].(string) + ":" + fmt.Sprintf("%.0f", mqp.config["port"].(float64)),
		ClientID: []byte(mqp.config["clientId"].(string)),
	})
	if err != nil {
		logger.Panic(err)
	}

	mqp.ready = true

	logger.Debugf("MQTT publisher %s connected.", mqp.id)

	mqp.wg = sync.WaitGroup{}
	mqp.wg.Add(1)
	mqp.wg.Wait()

	logger.Debugf("Stop MQTT publisher %s", mqp.id)

	mqp.ready = false
	if err := mqp.mqttClient.Disconnect(); err != nil {
		logger.Panic(err)
	}
}

func (mqp *MQTTPublisher) connectRecovery() {
	if r := recover(); r != nil {
		logger.Debugf("Recovering connection for MQTT publisher %s", mqp.id)
		time.Sleep(3 * time.Second)
		mqp.Connect()
	}
}

func (mqp *MQTTPublisher) Stop() {
	mqp.wg.Done()
}

func (mqp *MQTTPublisher) Publish(url string) {
	url2 := url[strings.Index(url, "://")+3:]
	logger.Debugf("Publishing MQTT publisher %s: %s", mqp.id, url2)
	lastIndex := strings.LastIndex(url2, "#")
	var message string = ""
	if lastIndex >= 0 {
		message = url[lastIndex:]
		url2 = url2[:lastIndex]
	}
	if mqp.ready {
		logger.Debugf(" topic: %s message: %s", url2, message)
		mqp.mqttClient.Publish(&client.PublishOptions{TopicName: []byte(url2), Message: []byte(message)})
	} else {
		logger.Warnf("Cannot publish %s", url)
	}

}
