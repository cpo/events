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

func NewMQTTPublisher() *interfaces.Publisher {
	return new(MQTTPublisher)
}

func (mq *MQTTPublisher) Initialize(eventManager interfaces.EventManager, config map[string]interface{}) {
	cfgStr, _ := json.Marshal(config)
	mq.id = config["name"].(string)
	mq.config = config
	mq.prefix = config["prefix"].(string)
	mq.eventManager = eventManager
	logger.Debugf("Initialize MQTT publisher %s with %s", mq.GetID(), cfgStr)
}

func (mq *MQTTPublisher) GetID() string {
	return mq.id
}

func (mq *MQTTPublisher) Connect() {
	logger.Infof("Connecting MQTT publisher %s", mq.id)
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

	logger.Debugf("MQTT publisher %s connected.", mq.id)

	mq.wg = sync.WaitGroup{}
	mq.wg.Add(1)
	mq.wg.Wait()

	logger.Debugf("Stop MQTT publisher %s", mq.id)

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

func (mq *MQTTPublisher) connectRecovery() {
	if r := recover(); r != nil {
		logger.Debugf("Recovering connection for MQTT publisher %s", mq.id)
		time.Sleep(3 * time.Second)
		mq.Connect()
	}
}

func (mq *MQTTPublisher) Stop() {
	mq.wg.Done()
}

func (mq *MQTTPublisher) Publish(url string) {
	logger.Debugf("Publishing MQTT publisher %s: %s", mq.id, url)
	lastIndex := strings.LastIndex(url, "#")
	var message string = ""
	if lastIndex >= 0 {
		message = url[lastIndex:]
		url = url[:lastIndex]
	}
	mq.mqttClient.Publish(&client.PublishOptions{TopicName: []byte(url), Message: []byte(message)})
}
