package manager

import (
	"encoding/json"
	"fmt"
	"github.com/cpo/events/bridges"
	"github.com/cpo/events/interfaces"
	"github.com/cpo/events/rules"
	"github.com/satori/go.uuid"
	"io/ioutil"
	logger "github.com/Sirupsen/logrus"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"time"
	"github.com/cpo/events/publishers"
)

type EventManagerImpl struct {
	id        string
	bridges   map[string]interfaces.Bridge
	rules     []interfaces.Rule
	publisher interfaces.Publisher
}

func New() interfaces.EventManager {
	return new(EventManagerImpl).initialize()
}

func (em *EventManagerImpl) initialize() interfaces.EventManager {
	em.id = uuid.NewV4().String()
	em.bridges = make(map[string]interfaces.Bridge)
	logger.Debugf("Initializing EventManager %s", em.id)
	return em
}

func (em *EventManagerImpl) AddBridge(bridge interfaces.Bridge, bridgeConfig map[string]interface{}) {
	bridge.Initialize(em, bridgeConfig)
	go bridge.Connect()
	logger.Debugf("Adding bridge %s", bridge.GetID())
	em.bridges[bridge.GetID()] = bridge
	logger.Debugf("#bridges: %d", len(em.bridges))
}

func (em *EventManagerImpl) Dispatch(url string) {
	if em.publisher != nil {
		logger.Debugf("Publishing event %s", url)
		go em.publisher.Publish(url)
	}

	logger.Debugf("Dispatching event %s", url)
	matches := 0
	for ruleN, rule := range em.rules {
		if rule.Matches(url) {
			matches++
			logger.Infof("Rule %d matches. Execute %d actions", ruleN, len(rule.GetActions()))
			myId := uuid.NewV4().String()
			for _, action := range rule.GetActions() {
				acJson, _ := json.Marshal(action)
				logger.Infof(" [%s] running %T action %s", myId, action, acJson)
				action.Run(em, myId)
			}
		}
	}
	logger.Debugf("Matched %d rules", matches)
}

func (em *EventManagerImpl) run() {
	go func() {
		for true {
			time.Sleep(30 * time.Second)
			ms := runtime.MemStats{}
			runtime.ReadMemStats(&ms)
			fmtTime := time.Unix(int64(ms.LastGC)/1000/1000/1000, 0).Local().Format("Mon 02-01-2006 15:04:05")
			logger.Debugf(" ==> goroutines: %d, heap: %d, lastgc: %s", runtime.NumGoroutine(), ms.HeapAlloc, fmtTime)
		}
	}()
	logger.Info("Running EventManager")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

	logger.Info("EventManager stopping all bridges...")
	for _, bridge := range em.bridges {
		bridge.Stop()
	}
}

func (em *EventManagerImpl) AddRule(rule interfaces.Rule, i map[string]interface{}) {
	em.rules = append(em.rules, rule)
	logger.Debugf("Adding rule %s", rule)
}

func (em *EventManagerImpl) Trigger(url string) {
	logger.Debugf("Triggering %s", url)
	re, _ := regexp.Compile("^([^:]*)://([^/]*)/(.*)")
	subs := re.FindStringSubmatch(url)
	logger.Debugf("Sub matches: %s", subs)
	switch subs[1] {
	case "bridge":
		logger.Debugf("Routing to %s bridge %s", subs[1], subs[2])
		em.triggerBridge(subs[2], subs[3])
	}
}

func (em *EventManagerImpl) triggerBridge(name string, uri string) {
	logger.Debugf("triggering bridge %s uri %s", name, uri)
	bridge, found := em.bridges[name]
	if found {
		logger.Debugf("Found bridge %s", name)
		bridge.Trigger(uri)
	} else {
		logger.Panicf("Bridge %s not found. Remaining URI %s", name, uri)
	}
}

func (em *EventManagerImpl) Start() {
	logger.Debugf("Reading configuration")
	config, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	jsonObject := make(map[string]interface{})
	err = json.Unmarshal(config, &jsonObject)
	if err != nil {
		panic(err)
	}

	if pubConfig := jsonObject["publisher"].(map[string]interface{}); pubConfig != nil {
		em.publisher = publishers.PublisherFactories[pubConfig["type"].(string)](em, pubConfig)
		go em.publisher.Connect()
	}

	logger.Debugf("Initializing bridges")
	for _, bridgeConfig := range jsonObject["bridges"].([]interface{}) {
		bridgeType := bridgeConfig.(map[string]interface{})["type"].(string)
		logger.Debugf("Instantiating bridge type %s", bridgeType)
		bridgeFactory, found := bridges.BridgeFactories[bridgeType]
		if found {
			newBridge := bridgeFactory()
			em.AddBridge(newBridge, bridgeConfig.(map[string]interface{}))
		} else {
			logger.Fatalf("Error while instantiating bridge: %b", err)
		}
	}

	logger.Debugf("Initializing rules")
	for _, ruleConfig := range jsonObject["rules"].([]interface{}) {
		ruleType := ruleConfig.(map[string]interface{})["type"].(string)
		logger.Debugf("Instantiating rule type %s", ruleType)
		ruleFactory, found := rules.RuleFactories[ruleType]
		if found {
			newRule := ruleFactory(ruleConfig.(map[string]interface{}))
			em.AddRule(newRule, ruleConfig.(map[string]interface{}))
		} else {
			logger.Fatalf("Error while instantiating bridge: %b", err)
		}
	}

	logger.Debugf(" === Bridges: %d, Rules: %d ===", len(em.bridges), len(em.rules))
	logger.Debugf("Entering main loop")
	em.run()

}
