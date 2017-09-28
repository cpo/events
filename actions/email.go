package actions

import (
	"github.com/cpo/events/interfaces"
	"net/smtp"
	"strings"
)

type EMailAction struct {
	Address  string
	User     string
	Password string
	From     string
	To       string
	Message  string
	Host     string
}

func (ea *EMailAction) Initialize(config map[string]interface{}) *EMailAction {
	ea.Address = config["address"].(string)
	ea.User = config["user"].(string)
	ea.Password = config["password"].(string)
	ea.From = config["from"].(string)
	ea.To = config["to"].(string)
	ea.Message = config["message"].(string)
	ea.Host = config["host"].(string)
	return ea
}

func (ea *EMailAction) Run(eventManager interfaces.EventManager, id string) {
	err := smtp.SendMail(ea.Address, smtp.PlainAuth("", ea.User, ea.Password, ea.Host), ea.From, []string{ea.To},
		[]byte(strings.Replace(ea.Message, "\\n", "\n", 0)))
	if err != nil {
		logger.Errorf("Error sending email: %s", err)
	}
}
