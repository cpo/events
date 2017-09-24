package manager

import (
	"encoding/json"
	"fmt"
	"github.com/cpo/events/bridges"
	"github.com/cpo/events/interfaces"
	"github.com/cpo/events/rules"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"time"
)

var logger = log.New(os.Stderr, "[EVTM] ", 1)

type EventManagerImpl struct {
	id      string
	bridges map[string]interfaces.Bridge
	rules   []interfaces.Rule
}

func New() interfaces.EventManager {
	return new(EventManagerImpl).initialize()
}

func (manager *EventManagerImpl) initialize() interfaces.EventManager {
	manager.id = uuid.NewV4().String()
	manager.bridges = make(map[string]interfaces.Bridge)
	logger.Printf("Initializing EventManager %s", manager.id)
	return manager
}

func (eventManager *EventManagerImpl) AddBridge(bridge interfaces.Bridge, bridgeConfig map[string]interface{}) {
	bridge.Initialize(eventManager, bridgeConfig)
	go bridge.Connect()
	logger.Printf("Adding bridge %s", bridge.GetID())
	eventManager.bridges[bridge.GetID()] = bridge
	logger.Printf("#bridges: %d", len(eventManager.bridges))
}

func (eventManager *EventManagerImpl) Dispatch(url string) {
	logger.Printf("Dispatching event %s", url)
	for ruleN, rule := range eventManager.rules {
		if rule.Matches(url) {
			logger.Printf("Rule %d matches. Execute %d actions", ruleN, len(rule.GetActions()))
			myId := uuid.NewV4().String()
			for _, action := range rule.GetActions() {
				acJson,_ := json.Marshal(action)
				logger.Printf(" [%s] running action %s", myId, acJson)
				action.Run(eventManager, myId)
			}
		}
	}
}

func (eventManager *EventManagerImpl) run() {
	go func() {
		for true {
			time.Sleep(30 * time.Second)
			ms := runtime.MemStats{}
			runtime.ReadMemStats(&ms)
			fmtTime := time.Unix(int64(ms.LastGC)/1000/1000/1000, 0).Local().Format("Mon 02-01-2006 15:04:05")
			logger.Printf(" ==> goroutines: %d, heap: %d, lastgc: %s", runtime.NumGoroutine(), ms.HeapAlloc, fmtTime)
		}
	}()
	logger.Printf("Running EventManager")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)

	logger.Printf("EventManager stopping all bridges...")
	for _, bridge := range eventManager.bridges {
		bridge.Stop()
	}
}

func (eventManager *EventManagerImpl) AddRule(rule interfaces.Rule, i map[string]interface{}) {
	eventManager.rules = append(eventManager.rules, rule)
	logger.Printf("Adding rule %s", rule)
}

func (eventManager *EventManagerImpl) Trigger(url string) {
	logger.Printf("Triggering %s", url)
	re, _ := regexp.Compile("^([^:]*)://([^/]*)/(.*)")
	subs := re.FindStringSubmatch(url)
	logger.Printf("Sub matches: %s", subs)
	switch subs[1] {
	case "bridge":
		logger.Printf("Routing to %s bridge %s", subs[1], subs[2])
		eventManager.triggerBridge(subs[2], subs[3])
	}
}

func (eventManager *EventManagerImpl) triggerBridge(name string, uri string) {
	logger.Printf("triggering bridge %s uri %s", name, uri)
	bridge,found := eventManager.bridges[name]
	if found {
		logger.Printf("Found bridge %s", name)
		bridge.Trigger(uri)
	} else {
		logger.Panicf("Bridge %s not found. Remaining URI %s", name, uri)
	}
}

func (ev *EventManagerImpl) Start() {
	logger.Printf("Reading configuration")
	config, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	jsonObject := make(map[string]interface{})
	err = json.Unmarshal(config, &jsonObject)
	if err != nil {
		panic(err)
	}

	logger.Printf("Initializing bridges")
	for _, bridgeConfig := range jsonObject["bridges"].([]interface{}) {
		bridgeType := bridgeConfig.(map[string]interface{})["type"].(string)
		logger.Printf("Instantiating bridge type %s", bridgeType)
		bridgeFactory, found := bridges.BridgeFactories[bridgeType]
		if found {
			newBridge := bridgeFactory()
			ev.AddBridge(newBridge, bridgeConfig.(map[string]interface{}))
		} else {
			logger.Fatalf("Error while instantiating bridge: %b", err)
		}
	}

	logger.Printf("Initializing rules")
	for _, ruleConfig := range jsonObject["rules"].([]interface{}) {
		ruleType := ruleConfig.(map[string]interface{})["type"].(string)
		logger.Printf("Instantiating rule type %s", ruleType)
		ruleFactory, found := rules.RuleFactories[ruleType]
		if found {
			newRule := ruleFactory(ruleConfig.(map[string]interface{}))
			ev.AddRule(newRule, ruleConfig.(map[string]interface{}))
		} else {
			logger.Fatalf("Error while instantiating bridge: %b", err)
		}
	}

	logger.Printf(" === Bridges: %d, Rules: %d ===", len(ev.bridges), len(ev.rules))
	logger.Printf("Entering main loop")
	ev.run()

}
