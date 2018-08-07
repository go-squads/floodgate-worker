package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type Mail struct {
	senderAccount account
	toIds         []string
	subject       string
	body          string
}

type SmtpServer struct {
	host string
	port string
}

type account struct {
	username string
	password string
}

func (s *SmtpServer) ServerName() string {
	return s.host + ":" + s.port
}

func (mail *Mail) BuildMessage() string {
	message := ""
	message += fmt.Sprintf("From: %s\r\n", mail.senderAccount.username)
	if len(mail.toIds) > 0 {
		message += fmt.Sprintf("To: %s\r\n", strings.Join(mail.toIds, ";"))
	}

	message += fmt.Sprintf("Subject: %s\r\n", mail.subject)
	message += "\r\n" + mail.body

	return message
}

func (s *SmtpServer) connectToServer(senderAccount account) *smtp.Client {
	log.Println(s.host)
	auth := smtp.PlainAuth("", senderAccount.username, senderAccount.password, s.host)
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.host,
	}
	conn, err := tls.Dial("tcp", s.ServerName(), tlsconfig)
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

func SendMail() {
	smtpServer := SmtpServer{host: "smtp.gmail.com", port: "465"}
	mail := Mail{}
	senderAccount := account{}
	senderAccount.username = "gosquad20@gmail.com"
	senderAccount.password = "gojekgosquad2.0"
	mail.toIds = []string{"vso_f1@yahoo.com"}
	mail.subject = "This is the email subject"
	mail.body = "Harry Potter and threat to Israel\n\nGood editing!!"

	client := smtpServer.connectToServer(senderAccount)

	err := client.Mail(senderAccount.username)
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

	messageBody := mail.BuildMessage()
	_, err = w.Write([]byte(messageBody))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	}
	client.Quit()
	log.Println("Mail sent successfully")
}
