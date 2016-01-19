package main

import (
	"errors"
	"log"
	"net/smtp"
	"time"
)

type Email struct {
	From     string
	Password string
	To       string
	Subject  string
	Body     string
}

var emailSendQueue chan *Email = make(chan *Email)

func (this *Email) Send() error {
	if this.From == "" || this.Password == "" || this.To == "" || this.Body == "" {
		return errors.New("Email record not complete enough to send.")
	}

	emailSendQueue <- this

	return nil
}

func EmailSendQueueHandler() {
	for email := range emailSendQueue {
		sendEmail(email)

		time.Sleep(1000 * time.Millisecond)
	}
}

func sendEmail(email *Email) error {
	msg := "From: " + email.From + "\n" +
		"To: " + email.To + "\n" +
		"Subject: " + email.Subject + "\n\n" +
		email.Body

	log.Printf("Sending email to %s\n", email.To)

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", email.From, email.Password, "smtp.gmail.com"),
		email.From,
		[]string{email.To},
		[]byte(msg))

	if err != nil {
		log.Printf("Error sending email to %s -- %s\n", email.To, err.Error())
		return err
	}

	return nil
}
