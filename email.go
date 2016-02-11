package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/ses"
)

type Email struct {
	From    string
	To      string
	Subject string
	Body    string
}

var emailSendQueue chan *Email = make(chan *Email)

func (this *Email) Send() error {
	if this.To == "" || this.Body == "" {
		return errors.New("Email record not complete enough to send.")
	}

	emailSendQueue <- this

	return nil
}

func EmailSendQueueHandler() {
	for email := range emailSendQueue {

		if err := sendSES(email); err != nil {
			log.Println(err.Error())
		}

		d, _ := time.ParseDuration("200ms")
		time.Sleep(d)
	}
}

// Gets emails from S3, saves them as messages.
func ProcessS3Emails(bucketName string, limit int64) (int, error) {
	svc := s3.New(session.New())

	params := &s3.ListObjectsInput{
		Bucket:  aws.String(bucketName),
		MaxKeys: aws.Int64(limit),
	}

	resp, err := svc.ListObjects(params)
	if err != nil {
		return 0, err
	}

	for _, obj := range resp.Contents {

		// Get the S3 object
		params := &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(*obj.Key),
		}

		resp, err := svc.GetObject(params)

		if err != nil {
			log.Println(err.Error())
			continue
		}

		// Parse the email
		m, err := mail.ReadMessage(resp.Body)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		body, _ := ioutil.ReadAll(m.Body)

		from := strings.Split(m.Header.Get("From"), "@")

		// Save the message
		msg := &Message{
			Network:  DomainToNetwork(from[1]),
			UUID:     from[0],
			Message:  string(body),
			Outgoing: 0,
		}

		if err := msg.Save(); err != nil {
			log.Println(err.Error())
			continue
		}

		// Push to message process queue...

		// Remove the object
		delParams := &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(*obj.Key),
		}

		if _, err := svc.DeleteObject(delParams); err != nil {
			log.Println(err.Error())
			continue
		}
	}

	return len(resp.Contents), nil
}

func sendSES(email *Email) error {
	svc := ses.New(session.New())
	params := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(email.To),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Data: aws.String(email.Body),
				},
			},
			Subject: &ses.Content{
				Data: aws.String(email.Subject),
			},
		},
		Source: aws.String("sms@iwillvote.us"),
		ReplyToAddresses: []*string{
			aws.String("sms@iwillvote.us"),
		},
	}

	if _, err := svc.SendEmail(params); err != nil {
		return err
	}

	return nil
}

func sendGmail(email *Email) error {
	if os.Getenv("GMAIL_ADDRESS") == "" || os.Getenv("GMAIL_PASSWORD") == "" {
		return errors.New("GMail not configured.")
	}

	email.From = os.Getenv("GMAIL_ADDRESS")

	msg := "From: " + email.From + "\n" +
		"To: " + email.To + "\n" +
		"Subject: " + email.Subject + "\n\n" +
		email.Body

	log.Printf("Sending email to %s\n", email.To)

	err := smtp.SendMail(
		"smtp.gmail.com:587",
		smtp.PlainAuth("", email.From, os.Getenv("GMAIL_PASSWORD"), "smtp.gmail.com"),
		email.From,
		[]string{email.To},
		[]byte(msg))

	if err != nil {
		log.Printf("Error sending email to %s -- %s\n", email.To, err.Error())
		return err
	}

	return nil
}
