package writemail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type MailAccount struct {
	ServerName string
	Address    string
	Password   string
}

type MailMsg struct {
	Account MailAccount
	Name    string
	To      []string
	Subject string
}

func (m MailMsg) Write(p []byte) (n int, err error) {

	from := mail.Address{m.Name, m.Account.Address}
	to := mail.Address{"", strings.Join(m.To, ",")}
	subj := m.Subject
	body := string(p)

	body = strings.Replace(body, "\n", "<br>", -1)
	body = strings.Replace(body, "\t", "&nbsp; &nbsp; &nbsp; &nbsp;", -1)

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	servername := m.Account.ServerName

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", m.Account.Address, m.Account.Password, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	c, err := smtp.Dial(servername)
	if err != nil {
		return 0, err
	}

	err = c.StartTLS(tlsconfig)
	if err != nil {
		return 0, err
	}
	// Auth
	if err = c.Auth(auth); err != nil {
		return 0, err
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return 0, err
	}

	for i := range m.To {
		if err = c.Rcpt(m.To[i]); err != nil {
			return 0, err
		}
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return 0, err
	}

	n, err = w.Write([]byte(message))
	if err != nil {
		return n, err
	}

	err = w.Close()
	if err != nil {
		return n, err
	}

	err = c.Quit()
	if err != nil {
		return 0, err
	}

	return n, nil
}
