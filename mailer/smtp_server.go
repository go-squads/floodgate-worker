package mailer

import (
	"net/smtp"
)

type SmtpServerServices interface {
	ConnectToServer() *smtp.Client
}
