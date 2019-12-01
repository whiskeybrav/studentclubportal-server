package mail

import (
	"github.com/NoteToScreen/maily-go/maily"
	"github.com/whiskeybrav/studentclubportal-server/configuration"
)

var Mail maily.Context

func ConfigureMail(config configuration.Config) {
	Mail = maily.Context{
		FromAddress:  config.Mail.FromAddress,
		FromDisplay:  config.Mail.FromDisplay,
		SendDomain:   config.Mail.SendDomain,
		SMTPHost:     config.Mail.SMTPHost,
		SMTPPort:     config.Mail.SMTPPort,
		SMTPUsername: config.Mail.SMTPUsername,
		SMTPPassword: config.Mail.SMTPPassword,
		TemplatePath: "./mail/templates",
	}
}
