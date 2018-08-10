package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	messageHeader = "%s for %s"
	messageBody   = "Number of %s logs for topic: %s, has reached %d. Subsequently, it exceeded the threshold limit of %d.\nPlease check the related application"
)

type MailerService interface {
	SendMail(string, string, int, int)
}

// Mock the SMTP Server
// Make an interface
// Mail implement Mailer
type mail struct {
	toIds   []string
	subject string
	body    string
}

type smtpServer struct {
	host string
	port string
}

func NewMailerService() MailerService {
	return &mail{}
}

func (s *smtpServer) getServerName() string {
	return s.host + ":" + s.port
}

func (s *smtpServer) connectToServer() *smtp.Client {
	log.Println(s.host)
	auth := smtp.PlainAuth("", os.Getenv("SENDER_ACC_USERNAME"), os.Getenv("SENDER_ACC_PASSWORD"), s.host)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.host,
	}
	conn, err := tls.Dial("tcp", s.getServerName(), tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		log.Panic(err)
	}

	if err = client.Auth(auth); err != nil {
		log.Panic(err)
	}
	return client
}

func (mail *mail) buildHeaderMessage() string {
	message := ""
	message += fmt.Sprintf("From: %s\r\n", os.Getenv("SENDER_ACC_USERNAME"))
	if len(mail.toIds) > 0 {
		message += fmt.Sprintf("To: %s\r\n", strings.Join(mail.toIds, ";"))
	}

	message += fmt.Sprintf("Subject: %s\r\n", mail.subject)
	message += "\r\n" + mail.body

	return message
}

func (mail *mail) buildMessageContent(logLevel string, topic string, levelValue int, threshold int) {
	mail.subject = fmt.Sprintf(messageHeader, logLevel, topic)
	mail.body = fmt.Sprintf(messageBody, logLevel, topic, levelValue, threshold)
}

func (mail *mail) SendMail(level string, topic string, levelValue int, threshold int) {
	smtpServer := smtpServer{host: "smtp.gmail.com", port: "465"}
	mail.toIds = []string{"hearthstone0298@gmail.com"}
	mail.buildMessageContent(level, topic, levelValue, threshold)

	client := smtpServer.connectToServer()

	err := client.Mail(os.Getenv("SENDER_ACC_USERNAME"))
	if err != nil {
		log.Panic(err)
	}

	for _, k := range mail.toIds {
		if err = client.Rcpt(k); err != nil {
			log.Panic(err)
		}
	}

	w, err := client.Data()
	if err != nil {
		log.Panic(err)
	}

	messageHeader := mail.buildHeaderMessage()
	_, err = w.Write([]byte(messageHeader))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}
	client.Quit()
	log.Info("Mail sent successfully")
}
