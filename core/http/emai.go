package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/jakubDoka/sterr"
)

// BotAccount for obvious security reasons email and password of a bot is loaded from private file
var BotAccount EmailAccount

// EmailAccount stores Email address and password
type EmailAccount struct {
	Email, Password string
}

func init() {
	path := "email.json"
	if strings.HasSuffix(os.Args[0], ".test.exe") {
		path = "C:/Users/jakub/Documents/programming/golang/src/myNotes/email.json"
	}
	bts, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bts, &BotAccount)
	if err != nil {
		panic(err)
	}
}

// errors  related to email handling
var (
	ErrInvalidEmail   = sterr.New("invalid email")
	ErrEmailVerifFail = sterr.New("verification of email failed")
)

// EmailStatus is for unmarshaling api responce
type EmailStatus struct {
	Status string `json:"status"`
}

// ValidEmail returns nil if inputted addres is valid existing email or ErrInvalidEmail
func ValidEmail(email string) (err error) {
	const apiURL = "https://isitarealemail.com/api/email/validate?email="

	url := apiURL + url.QueryEscape(email)
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ErrEmailVerifFail.Wrap(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ErrEmailVerifFail.Wrap(err)
	}

	var es EmailStatus
	err = json.Unmarshal(body, &es)
	if err != nil {
		return ErrEmailVerifFail.Wrap(err)
	}

	if es.Status != "valid" {
		return ErrInvalidEmail
	}

	return nil
}

// EmailSender handles sending of emails to targets
type EmailSender struct {
	Service, Sender string
	Auth            smtp.Auth
}

// NEmailSender ...
func NEmailSender(sender, password string, port int16) *EmailSender {
	const host = "smtp.gmail.com"

	return &EmailSender{
		Sender:  sender,
		Auth:    smtp.PlainAuth("", sender, password, host),
		Service: fmt.Sprintf("%s:%d", host, port),
	}
}

// Send sends email with message to targets
func (e *EmailSender) Send(message []byte, targets ...string) error {
	return smtp.SendMail(e.Service, e.Auth, e.Sender, targets, message)
}

// FormatVerificationEmail creates verification email
func FormatVerificationEmail(code, name string) []byte {
	t, err := template.ParseFiles("C:/Users/jakub/Documents/programming/golang/src/myNotes/core/http/template.html")
	if err != nil {
		panic(err)
	}

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	_, err = body.Write([]byte(fmt.Sprintf("Subject: This is a test subject \n%s\n\n", mimeHeaders)))
	if err != nil {
		panic(err)
	}

	err = t.Execute(&body, struct {
		Name string
		Code string
	}{name, code})
	if err != nil {
		panic(err)
	}

	return body.Bytes()
}
