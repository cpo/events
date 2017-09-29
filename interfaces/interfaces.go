package interfaces

type EventManager interface {
	Dispatch(url string)
	Trigger(string)
	Start()
}

type Publisher interface {
	Connect()
	Publish(url string)
}

type Action interface {
	Run(EventManager, string)
}

type Rule interface {
	Initialize(config map[string]interface{})
	Matches(url string) bool
	GetActions() []Action
}

type Bridge interface {
	Initialize(eventManager EventManager, config map[string]interface{})
	Connect()
	GetID() string
	Stop()
	Trigger(string)
}
