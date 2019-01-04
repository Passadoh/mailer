// Package mailer contains a utility to send an smtp
package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// Mail contains the information related to email.
type Mail struct {
	SenderID string
	Password string
	ToIds    []string
	CcIds    []string
	BccIds   []string
	Subject  string
	Body     string
}

// SMTPServer contains information related to smtp-server.
type SMTPServer struct {
	Host string
	Port string
}

func (s *SMTPServer) serverName() string {
	return s.Host + ":" + s.Port
}

func (mail *Mail) buildMessage() string {
	mime := "1.0;"
	contenType := "text/html; charset=\"UTF-8\";"
	message := ""
	message += fmt.Sprintf("MIME-version: %s\r\n", mime)
	message += fmt.Sprintf("Content-Type: %s\r\n", contenType)
	message += fmt.Sprintf("From: %s\r\n", mail.SenderID)

	if len(mail.ToIds) > 0 {
		message += fmt.Sprintf("To: %s\r\n", strings.Join(mail.ToIds, ";"))
	}
	if len(mail.CcIds) > 0 {
		message += fmt.Sprintf("Cc: %s\r\n", strings.Join(mail.CcIds, ";"))
	}

	message += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	message += "\r\n" + mail.Body

	return message
}

// Send expects a Mail struct and SMTPServer struct
func Send(mail Mail, smtpServer SMTPServer) {
	messageBody := mail.buildMessage()

	//build an auth
	auth := smtp.PlainAuth("", mail.SenderID, mail.Password, smtpServer.Host)

	// Gmail will reject connection if it's not secure
	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer.Host,
	}

	conn, err := tls.Dial("tcp", smtpServer.serverName(), tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	client, err := smtp.NewClient(conn, smtpServer.Host)
	if err != nil {
		log.Panic(err)
	}

	// step 1: Use Auth
	if err = client.Auth(auth); err != nil {
		log.Panic(err)
	}

	// step 2: add all from and to
	if err = client.Mail(mail.SenderID); err != nil {
		log.Panic(err)
	}
	// for _, k := range mail.ToIds {
	// 	if err = client.Rcpt(k); err != nil {
	// 		log.Panic(err)
	// 	}
	// }

	receivers := append(mail.ToIds, mail.CcIds...)
	receivers = append(receivers, mail.BccIds...)
	for _, k := range receivers {
		log.Println("sending to: ", k)
		if err = client.Rcpt(k); err != nil {
			log.Panic(err)
		}
	}

	// Data
	w, err := client.Data()
	if err != nil {
		log.Panic(err)
	}

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
