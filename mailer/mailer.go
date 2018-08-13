package mailer

import (
	"crypto/tls"
	"fmt"
	"io"
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
	SendMail(string, string, int, int) error
}

type SmtpServerServices interface {
	connectToServer() (SmtpClientTest, error)
	writeMail(io.WriteCloser, string) error
}

type SmtpClientTest interface {
	Data() (io.WriteCloser, error)
	Mail(string) error
	Auth(smtp.Auth) error
	Rcpt(string) error
	Quit() error
}

type mail struct {
	toIds      []string
	subject    string
	body       string
	smtpServer SmtpServerServices
}

type smtpServer struct {
	host string
	port string
}

func NewMailerService() MailerService {
	newSmtpServer := smtpServer{
		host: "smtp.gmail.com",
		port: "465",
	}

	return &mail{
		smtpServer: &newSmtpServer,
		toIds:      []string{"hearthstone0298@gmail.com"},
	}
}

func (s *smtpServer) getServerName() string {
	return s.host + ":" + s.port
}

func (s *smtpServer) connectToServer() (SmtpClientTest, error) {
	log.Println(s.host)
	auth := smtp.PlainAuth("", os.Getenv("SENDER_ACC_USERNAME"), os.Getenv("SENDER_ACC_PASSWORD"), s.host)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.host,
	}
	conn, err := tls.Dial("tcp", s.getServerName(), tlsconfig)
	if err != nil {
		return nil, err
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return nil, err
	}

	if err = client.Auth(auth); err != nil {
		return nil, err
	}

	err = client.Mail(os.Getenv("SENDER_ACC_USERNAME"))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return client, nil
}

func (mail *mail) buildMessage(senderUsername string, logLevel string, topic string, levelValue int, threshold int) string {
	message := ""
	message += fmt.Sprintf("From: %s\r\n", senderUsername)
	if len(mail.toIds) > 0 {
		message += fmt.Sprintf("To: %s\r\n", strings.Join(mail.toIds, ";"))
	}

	subject := fmt.Sprintf(messageHeader, logLevel, topic)
	body := fmt.Sprintf(messageBody, logLevel, topic, levelValue, threshold)

	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "\r\n" + body

	return message
}

func (mail *mail) SendMail(level string, topic string, levelValue int, threshold int) error {
	client, err := mail.smtpServer.connectToServer()
	if err != nil {
		log.Error(err)
		return err
	}

	for _, k := range mail.toIds {
		if err = client.Rcpt(k); err != nil {
			log.Error(err)
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		log.Error(err)
		return err
	}

	message := mail.buildMessage(os.Getenv("SENDER_ACC_USERNAME"), level, topic, levelValue, threshold)
	err = mail.smtpServer.writeMail(w, message)
	if err != nil {
		return err
	}

	client.Quit()
	log.Info("Mail sent successfully")
	return nil
}

func (s *smtpServer) writeMail(w io.WriteCloser, message string) error {
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Error(err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
