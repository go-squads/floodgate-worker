package mailer

import (
	"fmt"
	"io"
	"net/smtp"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

type smtpClientMock struct{}
type smtpServerMock struct {
	host string
	port string
}

type ioMock struct{}

func (s *smtpClientMock) Auth(smtp.Auth) error {
	return nil
}

func (s *smtpClientMock) Mail(from string) error {
	return nil
}

func (s *smtpClientMock) Rcpt(to string) error {
	return nil
}

func (s *smtpClientMock) Data() (io.WriteCloser, error) {
	var ioMock io.WriteCloser
	return ioMock, nil
}

func (s *smtpClientMock) Quit() error {
	return nil
}

func (s *smtpServerMock) writeMail(w io.WriteCloser, message string) error {
	return fmt.Errorf(message)
}

func (s *smtpServerMock) connectToServer() (SmtpClientTest, error) {
	var returnMock = smtpClientMock{}
	return &returnMock, nil
}

func TestIfMailSentCorrectly(t *testing.T) {
	testSmtpServer := smtpServerMock{
		host: "mailServer",
		port: "0000",
	}

	testMail := mail{
		toIds:      []string{"recipientMail"},
		subject:    "testNotification",
		body:       "thisIsNotificationEmail",
		smtpServer: &testSmtpServer,
	}
	err := godotenv.Load(os.ExpandEnv("$GOPATH/src/github.com/go-squads/floodgate-worker/.env"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	senderUsername := os.Getenv("SENDER_ACC_USERNAME")
	wantMessage := ""
	wantMessage += fmt.Sprintf("From: %s\r\n", senderUsername)
	wantMessage += fmt.Sprintf("To: recipientMail\r\n")
	wantMessage += fmt.Sprintf("Subject: INFO for test_topic\r\n")
	wantMessage += fmt.Sprintf("\r\nNumber of INFO logs for topic: test_topic, has reached 50. Subsequently, it exceeded the threshold limit of 40.\nPlease check the related application")

	resultHeader := testMail.SendMail("INFO", "test_topic", 50, 40)
	assert.Equal(t, fmt.Errorf(wantMessage), resultHeader)
}
