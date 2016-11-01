package gosmtp

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"time"
	"github.com/pborman/uuid"
)

const QueueSize = 100

type EmailSender struct {
	ServerAddr  string
	ServerPort  int
	SenderEmail string
	Username    string
	Password    string
	UseTLS      bool
	conn        *smtp.Client
	queue       chan *Task
}

func (e *EmailSender) AddQueue(t *Task) <-chan error {
	if e.queue == nil {
		e.queue = make(chan *Task, QueueSize)
	}
	t.err = make(chan error, 1)
	e.queue <- t
	return t.err
}

func (e *EmailSender) run(done chan<- interface{}) error {
	timer := time.NewTimer(10 * time.Second)
	if e.conn == nil || e.conn.Hello("localhost") != nil {

		if e.UseTLS {
			config := &tls.Config{
				ServerName: e.ServerAddr,
			}

			if conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", e.ServerAddr, 465), config); err != nil {
				return fmt.Errorf("[EMAIL] SERVER Connect Failed，%s", err)
			} else {
				var err error
				if e.conn, err = smtp.NewClient(conn, e.ServerAddr); err != nil {
					return fmt.Errorf("[EMAIL] SERVER Auth failed，%s", err)
				}
			}

		} else {

			var err error
			if e.conn, err = smtp.Dial(fmt.Sprintf("%s:%d", e.ServerAddr, e.ServerPort)); err != nil {
				return fmt.Errorf("[EMAIL] SERVER Connect Failed，%s", err)
			}
		}
		if e.Username != "" {
			auth := smtp.PlainAuth("", e.Username, e.Password, e.ServerAddr)
			if err := e.conn.Auth(auth); err != nil {
				return fmt.Errorf("[EMAIL] SERVER Auth failed，%s", err)
			}
		}
	}
	go func() {
	loop:
		for {
			select {
			case t := <-e.queue:
				if err := e.send(t); err != nil {
					select {
					case t.err <- err:
					case <-timer.C:
					}
				}
				close(t.err)
			case <-timer.C:
				break loop
			}
			timer.Reset(10 * time.Second)
		}
		done <- "ok"
		e.conn.Close()
		e.conn.Quit()
	}()
	return nil
}

func (e *EmailSender) send(t *Task) error {
	if err := e.conn.Mail(e.SenderEmail); err != nil {
		return fmt.Errorf("[EMAIL] Server can't accept sender，%s，%s", e.SenderEmail, err)
	}
	for _, r := range t.To {
		if err := e.conn.Rcpt(r); err != nil {
			return fmt.Errorf("[EMAIL] Server can't accept recipient，%s，%s", r, err)
		}
	}
	if wc, err := e.conn.Data(); err == nil {
		boundary := fmt.Sprintf("%x", md5.Sum(t.Content))
		wc.Write([]byte("Message-ID:<" + uuid.New() + "@hotelnabe.com.tw>\n"))
		wc.Write([]byte("Subject: " + "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(t.Subject)) + "?=\n"))
		wc.Write([]byte("MIME-Version: 1.0" + "\n"))
		wc.Write([]byte("Date: " + time.Now().Format(time.RFC1123Z) + "\n"))
		wc.Write([]byte("From: " + e.SenderEmail + "\n"))
		wc.Write([]byte("To: undisclosed-recipients:;\n"))
		wc.Write([]byte("Content-Type: multipart/mixed; boundary=b" + boundary + "\n\n"))
		wc.Write([]byte("This is a multi-part message in MIME format.\n\n--b" + boundary + "\n"))
		wc.Write([]byte("Content-Type: text/html;charset=UTF-8\n"))
		wc.Write([]byte("Content-Transfer-Encoding: base64\n\n"))
		body := base64.StdEncoding.EncodeToString(t.Content)
		for len(body) > 76 {
			wc.Write([]byte(body[0:76]))
			wc.Write([]byte("\n"))
			body = body[76:]
		}
		wc.Write([]byte(body[0:]))
		wc.Write([]byte("\n\n--b"+boundary+"--"))
		wc.Close()
	}else{
		return err
	}
	return nil
}

func (e *EmailSender) Init() <-chan error {
	er := make(chan error, 1)
	go func() {
		done := make(chan interface{}, 1)
		for {
			if len(e.queue) > 0 {
				if err := e.run(done); err != nil {
					er <- err
					continue
				}
				<-done
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return er
}

type Task struct {
	Subject string
	To      []string
	Content []byte
	err     chan error
}
