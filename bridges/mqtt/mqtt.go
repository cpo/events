package mqtt

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

type MQTTBridge struct {
	id           string
	mqttClient   *client.Client
	config       map[string]interface{}
	wg           sync.WaitGroup
	eventManager interfaces.EventManager
}

func NewMQTTBridge() interfaces.Bridge {
	return new(MQTTBridge)
}

func (mq *MQTTBridge) Initialize(eventManager interfaces.EventManager, config map[string]interface{}) {
	cfgStr, _ := json.Marshal(config)
	mq.id = config["name"].(string)
	mq.config = config
	mq.eventManager = eventManager
	logger.Debugf("Initialize MQTT bridge %s with %s", mq.GetID(), cfgStr)
}

func (mq *MQTTBridge) GetID() string {
	return mq.id
}

func (mq *MQTTBridge) Connect() {
	logger.Infof("Connecting MQTT bridge %s", mq.id)
	mq.mqttClient = client.New(&client.Options{})
	defer mq.mqttClient.Terminate()
	defer mq.connectRecovery()

	logger.Infof("MQTT connecting client %s to %s://%s:%.0f",
		mq.config["clientId"].(string),
		mq.config["proto"].(string),
		mq.config["host"].(string),
		mq.config["port"].(float64))

	err := mq.mqttClient.Connect(&client.ConnectOptions{
		Network:  mq.config["proto"].(string),
		Address:  mq.config["host"].(string) + ":" + fmt.Sprintf("%.0f", mq.config["port"].(float64)),
		ClientID: []byte(mq.config["clientId"].(string)),
	})
	if err != nil {
		logger.Panic(err)
	}

	// Subscribe to topics.
	err = mq.mqttClient.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte("#"),
				QoS:         mqtt.QoS2,
				// Define the processing of the message handler.
				Handler: func(topicName, message []byte) {
					logger.Debugf("MQTT topic=%s message=%s", string(topicName), string(message))
					go mq.eventManager.Dispatch(fmt.Sprintf("mqtt://%s/%s#%s", mq.config["name"], topicName, string(message)))
				},
			},
		},
	})

	logger.Debugf("MQTT bridge %s connected.", mq.id)

	mq.wg = sync.WaitGroup{}
	mq.wg.Add(1)
	mq.wg.Wait()

	logger.Debugf("Stop MQTT bridge %s", mq.id)

	// Unsubscribe from topics.
	err = mq.mqttClient.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte("/#"),
		},
	})
	if err != nil {
		logger.Panic(err)
	}
	if err := mq.mqttClient.Disconnect(); err != nil {
		logger.Panic(err)
	}
}

func (mq *MQTTBridge) connectRecovery() {
	if r := recover(); r != nil {
		logger.Debugf("Recovering connection for MQTT bridge %s", mq.id)
		time.Sleep(3 * time.Second)
		mq.Connect()
	}
}

func (mq *MQTTBridge) Stop() {
	mq.wg.Done()
}

func (mq *MQTTBridge) Trigger(uri string) {
	logger.Debugf("Publishing MQTT bridge %s: %s", mq.id, uri)
	lastIndex := strings.LastIndex(uri, "#")
	var message string = ""
	if lastIndex >= 0 {
		message = uri[lastIndex:]
		uri = uri[:lastIndex]
	}
	mq.mqttClient.Publish(&client.PublishOptions{TopicName: []byte(uri), Message: []byte(message)})
}
