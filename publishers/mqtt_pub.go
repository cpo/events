package publishers

import (
	"encoding/json"
	"github.com/yosssi/gmq/mqtt/client"
	logger "github.com/Sirupsen/logrus"
	"fmt"
	"github.com/cpo/events/interfaces"
	"github.com/yosssi/gmq/mqtt"
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
	eventManager interfaces.EventManager
}

func NewMQTTPublisher(manager interfaces.EventManager, config map[string]interface{}) interfaces.Publisher {
	return new(MQTTPublisher).Initialize(manager,config)
}

func (mqp *MQTTPublisher) Initialize(eventManager interfaces.EventManager, config map[string]interface{}) interfaces.Publisher {
	cfgStr, _ := json.Marshal(config)
	mqp.id = config["name"].(string)
	mqp.config = config
	mqp.prefix = config["prefix"].(string)
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

	// Subscribe to topics.
	err = mqp.mqttClient.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte("#"),
				QoS:         mqtt.QoS2,
				// Define the processing of the message handler.
				Handler: func(topicName, message []byte) {
					logger.Debugf("MQTT topic=%s message=%s", string(topicName), string(message))
					go mqp.eventManager.Dispatch(fmt.Sprintf("mqtt://%s/%s#%s", mqp.config["name"], topicName, string(message)))
				},
			},
		},
	})

	logger.Debugf("MQTT publisher %s connected.", mqp.id)

	mqp.wg = sync.WaitGroup{}
	mqp.wg.Add(1)
	mqp.wg.Wait()

	logger.Debugf("Stop MQTT publisher %s", mqp.id)

	// Unsubscribe from topics.
	err = mqp.mqttClient.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte("/#"),
		},
	})
	if err != nil {
		logger.Panic(err)
	}
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
	url = url[strings.Index(url,"://")+3:]
	logger.Debugf("Publishing MQTT publisher %s: %s", mqp.id, url)
	lastIndex := strings.LastIndex(url, "#")
	var message string = ""
	if lastIndex >= 0 {
		message = url[lastIndex:]
		url = url[:lastIndex]
	}
	mqp.mqttClient.Publish(&client.PublishOptions{TopicName: []byte(url), Message: []byte(message)})
}
