package gosmtp

import(
	"net/smtp"
	"fmt"
	"encoding/base64"
	"time"
	"crypto/md5"
)

type EmailSender struct{
	ServerAddr string
	SenderEmail string
	Username string
	Password string
	conn *smtp.Client
	queue chan *Task
}

func (e *EmailSender) AddQueue(t *Task) <-chan string {
	if e.queue == nil {
		e.queue = make(chan *Task,100)
	}
	r := make(chan string,1)
	t.result = r
	e.queue<-t
	return r
}

func (e *EmailSender) run(done chan <- interface{}) error{
	timer := time.NewTimer(10 * time.Second)
	if e.conn == nil || e.conn.Hello("localhost")!=nil{
		var err error
		if e.conn, err = smtp.Dial(e.ServerAddr);err!=nil{
			return fmt.Errorf("[EMAIL] SERVER連接錯誤，%s",err)
		}
		if e.Username != ""{
			auth := smtp.PlainAuth("", e.Username, e.Password, e.ServerAddr)
			if err := e.conn.Auth(auth);err !=nil {
				return fmt.Errorf("[EMAIL] SERVER認證錯誤，%s",err)
			}
		}
	}
	go func(){
		loop:
		for{
			select{
				case t := <-e.queue:
					if err := e.send(t);err == nil{
						t.result <- "Success"
					}else{
						t.result <- err.Error()
					}
				case <- timer.C:
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

func (e *EmailSender) send(t *Task) error{
	if err := e.conn.Mail(e.SenderEmail); err != nil {
		return fmt.Errorf("[EMAIL] 無法指定寄件者，%s，%s",e.SenderEmail,err)
	}
	for _, r := range(t.To){
		if err := e.conn.Rcpt(r);err != nil{
			return fmt.Errorf("[EMAIL] 無法指定收件者，%s，%s",r,err)
		}
	}
	if wc,err:=e.conn.Data();err == nil{
		boundary := fmt.Sprintf("%x",md5.Sum(t.Content))
		wc.Write([]byte("Subject: " + "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(t.Subject)) +"?=\n"))
		wc.Write([]byte("MIME-Version: 1.0" + "\n"))
		wc.Write([]byte("Content-Type: multipart/mixed; boundary = b" + boundary + "\n"))
		wc.Write([]byte("This is a multi-part message in MIME format.\n\n--b" + boundary + "\n"))
		wc.Write([]byte("Content-Type: TEXT/html;charset=uft-8\n"))
		wc.Write([]byte("Content-Transfer-Encoding: base64\n\n"))
		body := base64.StdEncoding.EncodeToString(t.Content)
		for len(body)>76{
			wc.Write([]byte(body[0:76]))
			wc.Write([]byte("\n"))
			body = body[76:]		
		}
		wc.Write([]byte(body[0:]))		
		wc.Write([]byte("\n\n--b"+boundary))
		wc.Close()
	}
	return nil
}

func (e *EmailSender) Init() <-chan error{
	er := make(chan error,1)
	go func(){
		done := make(chan interface{},1)
		for{
			if len(e.queue)>0{
				if err := e.run(done);err!=nil{
					er <- err
					continue
				}				
				<- done
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return er
}

type Task struct{
	Subject string
	To []string
	Content []byte
	result chan string
}